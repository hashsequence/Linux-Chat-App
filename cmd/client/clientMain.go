package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	clientLib "github.com/hashsequence/Linux-Chat-App/pkg/client"
	linuxChatAppPb "github.com/hashsequence/Linux-Chat-App/pkg/pb/LinuxChatAppPb"
)

var (
	crt  = "./ssl/client-cert.pem"
	key  = "./ssl/client-key.pem"
	ca   = "./ssl/ca-cert.pem"
	addr = "localhost:50051"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
		fmt.Println("Client Closing")
	}()
	client, err := clientLib.CreateClient(ca, crt, key, addr)
	if err != nil {
		log.Fatalf("could not create client stream %v: %v", addr, err)
	}

	input := make(chan string)
	done := make(chan struct{})
	go func() {
		scan := bufio.NewScanner(os.Stdin)
		for scan.Scan() {
			input <- scan.Text()
		}
	}()
	func() {
		for cmd := range input {
			if cmd == "d" {
				close(done)
				return
			} else if cmd == "viewChatRooms" {
				req := &linuxChatAppPb.ViewListOfChatRoomsRequest{}
				resp, err := client.ViewListOfChatRooms(context.Background(), req)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println(resp)
				}
			} else if cmd == "viewUsers" {
				req := &linuxChatAppPb.ViewListOfUsersRequest{}
				resp, err := client.ViewListOfUsers(context.Background(), req)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println(resp)
				}
			}
		}
	}()
}
