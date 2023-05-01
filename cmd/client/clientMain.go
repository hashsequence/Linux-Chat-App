package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	clientLib "github.com/hashsequence/Linux-Chat-App/pkg/client"
	linuxChatAppPb "github.com/hashsequence/Linux-Chat-App/pkg/pb/LinuxChatAppPb"
	utils "github.com/hashsequence/Linux-Chat-App/pkg/utils"
)

var (
	crt  = "./ssl/client-cert.pem"
	key  = "./ssl/client-key.pem"
	ca   = "./ssl/ca-cert.pem"
	addr = "localhost:50051"
)

var userName string
var clientChatRoomNames map[string]string = map[string]string{}

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
	defer func() {
		client.DeleteUser(context.Background(), &linuxChatAppPb.DeleteUserRequest{
			UserName: userName,
		})
	}()

	input := make(chan string)
	done := make(chan struct{})
	scan := bufio.NewScanner(os.Stdin)
	userNameCreated := false
	go func() {
		for scan.Scan() {
			if !userNameCreated {
				fmt.Println("please Create UserName:")
			}
			input <- scan.Text()
		}
	}()
	func() {
		for cmd := range input {
			args := utils.GetArgsArr(cmd)
			if len(args) == 1 && args[0] == "d" {
				close(done)
				return
			} else if !userNameCreated && len(args) == 1 {
				req := &linuxChatAppPb.CreateUserNameRequest{
					UserName: args[1],
				}
				resp, err := client.CreateUser(context.Background(), req)
				if resp.GetSuccess() {
					fmt.Printf("Username: %v created successfully!\n", resp.GetUserName())
					userNameCreated = true
				} else {
					fmt.Printf("Username: %v failed to be created, please try again.\n", resp.GetUserName())
				}
				if err != nil {
					fmt.Println(err)
				}
			} else if len(args) == 1 && args[0] == "viewChatRooms" {
				req := &linuxChatAppPb.ViewListOfChatRoomsRequest{}
				resp, err := client.ViewListOfChatRooms(context.Background(), req)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println(resp)
				}
			} else if len(args) == 1 && args[0] == "viewUsers" {
				req := &linuxChatAppPb.ViewListOfUsersRequest{}
				resp, err := client.ViewListOfUsers(context.Background(), req)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println(resp)
				}
			} else if len(args) >= 3 && args[0] == "createChatRoom" {
				if userName != "" {
					req := &linuxChatAppPb.CreateChatRoomRequest{
						ChatRoomName: args[1],
						Users:        args[2:],
					}
					resp, err := client.CreateChatRoom(context.Background(), req)
					clientChatRoomNames[resp.GetChatRoomName()] = resp.GetChatRoomName()
					if userName == "" {
						userName = resp.GetHostUserName()
					}
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println(resp)
					}
				} else {
					fmt.Println("Must Create username")
				}
			} else if len(args) == 2 && args[0] == "joinChatRoom" {
				if userName != "" {
					req := &linuxChatAppPb.JoinChatRoomRequest{
						UserName:     userName,
						ChatRoomName: args[2],
					}
					resp, err := client.JoinChatRoom(context.Background(), req)
					if resp.GetSuccess() {
						clientChatRoomNames[resp.GetChatRoomName()] = resp.GetChatRoomName()
					}
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println(resp)
					}
				} else {
					fmt.Println("Must Create username")
				}
			} else if len(args) == 2 && args[0] == "leaveChatRoom" {
				if userName != "" {
					req := &linuxChatAppPb.LeaveChatRoomRequest{
						UserName:     userName,
						ChatRoomName: args[2],
					}
					resp, err := client.LeaveChatRoom(context.Background(), req)
					if resp.GetSuccess() {
						clientChatRoomNames[resp.GetChatRoomName()] = resp.GetChatRoomName()
					}
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println(resp)
					}
				} else {
					fmt.Println("Must Create username")
				}
			}
		}
	}()
}
