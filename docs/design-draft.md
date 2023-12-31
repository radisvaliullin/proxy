1. The proxy should accept TCP connection with the mTLS layer.
1. Proxy should authenticate clients using mTLS and Bearer token passed through x509 certificate (common name field).
1. Proxy should load-balance client connections to one of available servers.
1. Proxy should authorize clients for which available servers it can be forward.
1. Certificates, Client identity, Client authorizations should be set via configuration (config file, env var, cli arguments).
1. Proxy should support connection limits for clients.
1. Proxy should provide a least connection balancing.
1. Proxy should provide health check of server aliveness and balance only for the live of them.
