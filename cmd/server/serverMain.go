package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	server "github.com/hashsequence/Linux-Chat-App/pkg/server"
)

const (
	crt              = "./ssl/server-cert.pem"
	key              = "./ssl/server-key.pem"
	caCrt            = "./ssl/ca-cert.pem"
	addr             = "localhost:50051"
	messageQueueSize = 1000
	ttl              = 1800 //1800 seconds or 30 minutes
)

func main() {
	done := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Cleanup...")
		close(done)
		os.Exit(1)
	}()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
		fmt.Println("Server Closing")
		close(done)
	}()

	fmt.Println("Starting Server")

	s := server.NewLinuxChatServer(messageQueueSize, ttl, done)
	if err := s.Serve(caCrt, crt, key, addr); err != nil {
		panic(err)
	}

}
