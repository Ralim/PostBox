package main

import "github.com/ralim/PostBox/webserver"

func main() {
	server := webserver.NewServer()
	server.StartWebserver()

}
