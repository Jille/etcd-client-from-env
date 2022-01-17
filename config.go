// Package clientconfig constructs a Config for connecting to etcd.
//
// You can use the simple mode and only call Get and use our defaults.
// If you want to customize defaults, either do that on Get's return value, or first call Defaults, modify it and then call Apply to read the environment variables.
package clientconfig

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Get is the easiest way to get a clientv3.Config if you don't have any defaults that have less priority than client configuration.
func Get() (clientv3.Config, error) {
	return Apply(Defaults())
}

// Defaults are the defaults used by this library, but you can overwrite them.
// After overwriting them, pass the Config to Apply to get the configuration from the environment.
func Defaults() clientv3.Config {
	return clientv3.Config{
		DialTimeout:      15 * time.Second,
		AutoSyncInterval: 5 * time.Minute,
	}
}

// Apply reads the environment variables and returns a modified copy of the given config.
func Apply(c clientv3.Config) (clientv3.Config, error) {
	settings := map[string]string{}
	for _, k := range []string{"ETCD_ENDPOINTS", "ETCD_USERNAME", "ETCD_PASSWORD", "ETCD_USERNAME_AND_PASSWORD", "ETCD_INSECURE_SKIP_VERIFY", "ETCD_SERVER_CA", "ETCD_CLIENT_CERT", "ETCD_CLIENT_KEY"} {
		ev := os.Getenv(k)
		fn := os.Getenv(k + "_FILE")
		if ev != "" && fn != "" {
			return c, fmt.Errorf("conflicting value for %s: both %s and %s_FILE are set", k, k, k)
		} else if ev != "" {
			settings[k] = ev
		} else if fn != "" {
			b, err := ioutil.ReadFile(fn)
			if err != nil {
				return c, fmt.Errorf("error reading %q (for %s_FILE): %v", fn, k, err)
			}
			settings[k] = string(b)
		}
	}
	if v := settings["ETCD_ENDPOINTS"]; v != "" {
		c.Endpoints = strings.Split(v, ",")
	}
	if v := settings["ETCD_USERNAME_AND_PASSWORD"]; v != "" {
		if settings["ETCD_USERNAME"] != "" || settings["ETCD_PASSWORD"] != "" {
			return c, errors.New("you can't set both ETCD_USERNAME_AND_PASSWORD and ETCD_USERNAME or ETCD_PASSWORD")
		}
		sp := strings.SplitN(v, ":", 2)
		if len(sp) != 2 {
			return c, errors.New("invalid ETCD_USERNAME_AND_PASSWORD: user and password should be separated with a colon (:)")
		}
		settings["ETCD_USERNAME"] = sp[0]
		settings["ETCD_PASSWORD"] = sp[1]
	}
	if v := settings["ETCD_USERNAME"]; v != "" {
		c.Username = v
	}
	if v := settings["ETCD_PASSWORD"]; v != "" {
		c.Password = v
	}
	if v := settings["ETCD_INSECURE_SKIP_VERIFY"]; v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return c, fmt.Errorf("failed to parse ETCD_INSECURE_SKIP_VERIFY as bool (%q)", v)
		}
		if c.TLS == nil {
			c.TLS = new(tls.Config)
		}
		c.TLS.InsecureSkipVerify = b
	}
	if v := settings["ETCD_SERVER_CA"]; v != "" {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(v)) {
			return c, errors.New("certificate(s) in ETCD_SERVER_CA(_FILE) were invalid PEM certificates")
		}
		if c.TLS == nil {
			c.TLS = new(tls.Config)
		}
		c.TLS.RootCAs = pool
	}
	vc, vk := settings["ETCD_CLIENT_CERT"], settings["ETCD_CLIENT_KEY"]
	if vc != "" && vk != "" {
		crt, err := tls.X509KeyPair([]byte(vc), []byte(vk))
		if err != nil {
			return c, fmt.Errorf("failed to parse ETCD_CLIENT_CERT+ETCD_CLIENT_KEY: %v", err)
		}
		if c.TLS == nil {
			c.TLS = new(tls.Config)
		}
		c.TLS.Certificates = []tls.Certificate{crt}
	} else if vc != "" || vk != "" {
		return c, errors.New("either both of ETCD_CLIENT_CERT(_FILE) and ETCD_CLIENT_KEY(_FILE) must be given or neither")
	}
	return c, nil
}
