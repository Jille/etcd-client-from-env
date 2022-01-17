# etcd-client-from-env

This library reads environment variables and returns a [clientv3.Config](https://pkg.go.dev/go.etcd.io/etcd/client/v3#Config).

It makes it easy to write tools against etcd that give the user control over how to connect to etcd.

It currently supports the following settings (but we welcome contributions). Each setting can also be passed by setting k_FILE (like ETCD_SERVER_CA_FILE) to a filename from where to read the value.

- ETCD_ENDPOINTS: A comma separated list of etcd endpoints. (required)
- ETCD_USERNAME: Username for etcd authentication.
- ETCD_PASSWORD: Password for etcd authentication.
- ETCD_USERNAME_AND_PASSWORD: username:password pair (separated by a colon) for etcd authentication.
- ETCD_INSECURE_SKIP_VERIFY: "true" to disable verification of the etcd server certificate (insecure).
- ETCD_SERVER_CA: PEM encoded CA certificate that has signed the server certificate.
- ETCD_CLIENT_CERT: PEM encoded certificate for CN authentication.
- ETCD_CLIENT_KEY: PEM encoded private key for CN authentication.

All settings are optional except ETCD_ENDPOINTS. If you pass a username, you must also pass a password and vice versa. Same for a client cert/key.
