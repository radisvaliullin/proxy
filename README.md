# proxy
A TCP Proxy with support:
* load-balancing
* AuthN via mTLS

Implemented for education purpose. Implemented following best practices but do not support much feautures.\
Just accept mTLS clients and forward to tcp server.\
Example of how need write right Go concurrent code with right Go style.

## Description
* A simple mTLS proxy.
* Client connected via mTLS.
* For each client need generate certificate associated with self-signer parent client root certificate. (see [keycertgen example](#keycertgen-examples))
* For identify clients server will use client self-signer root certificate. (see [keycertgen example](#keycertgen-examples))
* For indetify server need generate self-signer server root certificate. (see [keycertgen example](#keycertgen-examples))
* Clients also should be listed in config.yaml file in auth clients section. (see [example.config.yaml](./config/example.config.yaml))
* Proxy accept connections and forward clients to upstream list servers. Proxy balance connections using a least connection method.
* For clients can be set limit of connection number. Each client can be limited for hist own list of upstreams in range of upstreams.

## CMD usage
command line apps

### keycertgen examples
gen proxy server CA self-signed certificate and key
```
go run cmd/keycertge/main.go -isca
```

gen client root CA self-signed certificate and key
```
go run cmd/keycertgen/main.go -key clientcakey -cert clientcacert -isca
```

gen end client certificate and key (provide parent CA cert and key)
```
go run cmd/keycertgen/main.go -key clientkey -cert clientcert -clientid client@client.org -parentkey clientcakey -parentcert clientcacert
```

### run proxy
```
go run cmd/proxy/main.go
```
or build and run from binary
```
go build -o bin/proxy cmd/proxy/main.go
./bin/proxy
```
