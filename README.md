# Forward Crypter
A program to encrypt and forward packets via websocket
## Why this?
You can use this just like shadowsocks or brooks. Use this program to forward a socks5 server on your system and use it. You can also forward an HTTP proxy.
## Install
Use the executables from [releases](https://github.com/HirbodBehnam/Forward-Crypter/releases) and download one for your server and client.
### Server
```
./fc -l 8080 -k "password" server
```
Run this command. It will get all traffics from port 8080 and decrypt them with "password".
### Client
```
 ./fc -l 1080 -k "password" client --forward 127.0.0.1:8080 -s 1.1.1.1:1080
```
This will get all incoming traffics from port 1080, encrypts them with the key "password" and forwards them to 1.1.1.1:1080. In server, they will be forwarded to 127.0.0.1:8080.

This also means that you can expose the local server on your own PC.
### Building
```bash
go get github.com/gorilla/websocket
go get github.com/urfave/cli
go get golang.org/x/crypto/chacha20poly1305
go build main.go
```
