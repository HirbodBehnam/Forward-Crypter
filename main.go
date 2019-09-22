package main

import (
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/urfave/cli"
	"golang.org/x/crypto/chacha20poly1305"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
)

var To string
var Listen string
var Verbose = false
var BufferSize int
var upgrader = websocket.Upgrader{EnableCompression:true}
var chacha cipher.AEAD
var nonce = make([]byte,24)

func server(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	proxy, err := net.Dial("tcp", To)
	if err != nil {
		log.Println("Error dialing the server:",err)
		return
	}
	defer proxy.Close()
	defer c.Close()
	go func() { //Get the data from the source
		buf := make([]byte, BufferSize)
		for {
			nr, er := proxy.Read(buf)
			if nr > 0 {
				crypt := chacha.Seal(nil,nonce,buf[0:nr],nil)
				ew := c.WriteMessage(websocket.BinaryMessage,crypt)
				if ew != nil {
					LogVerbose("Error on writing data to client:",ew)
					break
				}
			}
			if er != nil {
				if er != io.EOF {
					LogVerbose("Error on reading data from server client:",er)
				}
				break
			}
		}
	}()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			LogVerbose("Error reading the message from client:", err)
			break
		}
		res, err := chacha.Open(nil,nonce,message,nil)
		if err != nil{
			log.Println(err)
			break
		}
		_, err = proxy.Write(res)
		if err != nil{
			LogVerbose("Error writing to server:", err)
			break
		}
	}
}

func main() {
	key := ""
	app := cli.NewApp()
	app.Name = "Forward Crypt"
	app.Version = "1.0.0"
	app.Usage = "Forward your encrypted packets over websocket"
	app.Author = "Hirbod Behnam"
	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "listen, l",
			Required: true,
			Usage: "The port that proxy listen on.",
			Destination: &Listen,
		},
		cli.StringFlag{
			Name: "to, t",
			Required: true,
			Usage: "Where the packets must be send. Server address on client application",
			Destination: &To,
		},
		cli.StringFlag{
			Name: "key, k",
			Usage: "The key of the server and client encryption",
			Destination: &key,
		},
		cli.IntFlag{
			Name: "buffer",
			Usage: "The buffer size in bytes",
			Value: 64 * 1024,
			Destination: &BufferSize,
		},
		cli.BoolFlag{
			Name: "verbose",
			Usage: "Enable verbose mode",
			Destination: &Verbose,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "Run as server app",
			Action:  func(c *cli.Context) error {
				fmt.Println("Server mode: true")
				fmt.Println("Forward to:",To)
				fmt.Println("Listen on",Listen)
				LogVerbose("Verbose is on and staring")
				{
					var err error
					chacha, err = chacha20poly1305.NewX(keyToByte(key))
					if err != nil{
						panic(err)
					}
				}

				http.HandleFunc("/", server)
				log.Fatal(http.ListenAndServe(":" + Listen, nil))
				return nil
			},
		},
		{
			Name:    "client",
			Aliases: []string{"c"},
			Usage:   "Run as client",
			Action:  func(c *cli.Context) error {
				fmt.Println("Server mode: false")
				fmt.Println("Forward to:",To)
				fmt.Println("Listen on",Listen)
				LogVerbose("Verbose is on and staring")
				{
					var err error
					chacha, err = chacha20poly1305.NewX(keyToByte(key))
					if err != nil{
						panic(err)
					}
				}

				ln, err := net.Listen("tcp", "127.0.0.1:" + Listen)
				if err != nil {
					panic(err)
				}

				for {
					conn, err := ln.Accept()
					if err != nil {
						LogVerbose("Error accepting connection:",err)
						continue
					}

					go func(conn net.Conn) {
						defer conn.Close()
						u := url.URL{Scheme: "ws", Host: To, Path: "/"}

						c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
						if err != nil {
							log.Println("Error dialing server:", err)
						}
						defer c.Close()
						go func() {
							for {
								_, message, err := c.ReadMessage()
								if err != nil {
									LogVerbose("Error reading from the server:", err)
									return
								}
								res, e := chacha.Open(nil,nonce,message,nil)
								if e != nil{
									log.Println(e)
									break
								}

								_ , err = conn.Write(res)
								if err != nil {
									LogVerbose("Error writing to local connection:", err)
									return
								}
							}
						}()
						buf := make([]byte, BufferSize)
						for {
							nr, er := conn.Read(buf)
							if nr > 0 {
								crypt := chacha.Seal(nil,nonce,buf[0:nr],nil)
								ew := c.WriteMessage(websocket.BinaryMessage,crypt)
								if ew != nil {
									LogVerbose("Error on writing to server:",ew)
									break
								}
							}
							if er != nil {
								if er != io.EOF {
									LogVerbose("Error on reading from you:", er)
								}
								break
							}
						}
					}(conn)
				}
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func LogVerbose(v ...interface{})  {
	if Verbose{
		log.Println(v)
	}
}

func keyToByte(key string) []byte {
	h := sha256.New()
	h.Write([]byte(key))
	return h.Sum(nil)
}