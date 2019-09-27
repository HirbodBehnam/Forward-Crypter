# Forward Crypter
A program to encrypt and forward packets via websocket
## Why this?
You can use this just like shadowsocks or brooks. Use this program to forward a socks5 server on your system and use it. You can also forward an HTTP proxy.
## Install
Use the executables from [releases](https://github.com/HirbodBehnam/Forward-Crypter/releases) and download one for your server and client.
### Server
```
./sf server -l 8080 -t 127.0.0.1:8888 -k "password"
```
Run this command. It will get all traffics from port 8080, decrypt them with "password" and forwards them to 127.0.0.1:8888.
### Client
```
 ./sf client -l 1080 -t 1.1.1.1:8080 -k "password"
```
This will get all incoming traffics from port 1080, encryptes them with the key "password" and forwards them to port 1.1.1.1:1080.
### Building
```bash
go get github.com/gorilla/websocket
go get github.com/urfave/cli
go get golang.org/x/crypto/chacha20poly1305
go build main.go
```
