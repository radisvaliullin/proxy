# proxy
A TCP Proxy with support:
* load-balancing
* AuthN via mTLS

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
```
