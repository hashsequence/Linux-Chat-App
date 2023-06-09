package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	clientLib "github.com/hashsequence/Linux-Chat-App/pkg/client"
	"github.com/hashsequence/Linux-Chat-App/pkg/pb/LinuxChatAppPb"
	linuxChatAppPb "github.com/hashsequence/Linux-Chat-App/pkg/pb/LinuxChatAppPb"
	utils "github.com/hashsequence/Linux-Chat-App/pkg/utils"
)

var (
	crt              = "./ssl/client-cert.pem"
	key              = "./ssl/client-key.pem"
	ca               = "./ssl/ca-cert.pem"
	addr             = "localhost:50051"
	heatbeatInterval = 300 // 300 seconds
)

var userName string
var clientChatRoomNames map[string]bool = map[string]bool{}
var chatRoomSessions map[string](LinuxChatAppPb.LinuxChatAppService_SendMessageClient) = map[string](LinuxChatAppPb.LinuxChatAppService_SendMessageClient){}
var Help map[string]string = map[string]string{
	"viewChatRooms":  "get a list of available chatRooms\n Example: viewChatRooms",
	"viewUsers":      "get a list of users logged into server\n Example: viewUsers ",
	"createChatRoom": "create chatroom\n createChatRoom <chatRoomName>\nExample: createChatRoom r1\n",
	"joinChatRoom":   "join chatrooom\n joinChatRoom <chatRoomName>\nExample: createChatRoom r1\n",
	"leaveChatRoom":  "leave chatRoom\n leaveChatRoom <chatRoomName>\nExample: createChatRoom r1\n",
	"send":           "send message to chat service\n send <chatRoomName> <msg>\nExample: createChatRoom r1 Hello World!\n",
}

const RootHelpMessage = "\nCommands:\nviewChatRooms: get a list of available chatRooms\nviewUsers: get a list of users logged into server\ncreateChatRoom: create chatroom\njoinChatRoom: join chatrooom\nleaveChatRoom: leave chatRoom\nsend: send message to chat service\n\n"

func main() {

	input := make(chan string)
	done := make(chan struct{})
	errChannel := make(chan error)
	scan := bufio.NewScanner(os.Stdin)
	userNameCreated := false

	client, err := clientLib.CreateClient(ca, crt, key, addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	cleanUp := func() {
		fmt.Printf("Client Logging off, deleting %v from ChatService\n", userName)
		client.DeleteUser(context.Background(), &linuxChatAppPb.DeleteUserRequest{
			UserName: userName,
		})
		close(done)
		close(errChannel)
		fmt.Println("Client Closing")
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
		cleanUp()
	}()

	go func() {
		errMsg := <-errChannel
		cleanUp()
		panic(errMsg)
	}()

	go func() {
		for scan.Scan() {
			input <- scan.Text()
		}
	}()

	ticker := time.NewTicker(time.Duration(heatbeatInterval) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				if len(chatRoomSessions) > 0 {
					_, err := client.Ping(context.Background(), &linuxChatAppPb.PingRequest{
						UserName: userName,
					})

					if err != nil {
						fmt.Println("\nPing request failed")
						errChannel <- err
						return
					}
				}
			case <-done:
				fmt.Println("\nShutting down client because heartbeat ping failed")
				ticker.Stop()
				return
			}
		}
	}()
	if !userNameCreated {
		fmt.Println("\nplease Create UserName:")
	}
	func() {
		for cmd := range input {
			select {
			case <-done:
				return
			default:
				args := utils.GetArgsArr(cmd)
				if len(args) == 1 && args[0] == "d" {
					return
				} else if !userNameCreated {
					fmt.Printf("Attempting to create user: %v\n", args[0])
					req := &linuxChatAppPb.CreateUserNameRequest{
						UserName: args[0],
					}
					resp, err := client.CreateUser(context.Background(), req)
					if resp.GetUserName() == args[0] {
						fmt.Printf("Username: %v created successfully!\n", resp.GetUserName())
						userNameCreated = true
						userName = resp.GetUserName()
					} else {
						fmt.Printf("Username: %v failed to be created, please try again.\n", resp.GetUserName())
						fmt.Println("please Create UserName:")
					}
					if err != nil {
						errChannel <- err
					}
				} else if len(args) == 1 && args[0] == "viewChatRooms" {
					req := &linuxChatAppPb.ViewListOfChatRoomsRequest{
						UserName: userName,
					}
					resp, err := client.ViewListOfChatRooms(context.Background(), req)
					if err != nil {
						errChannel <- err
					} else {
						fmt.Println(resp.GetChatRoomNames())
					}
				} else if len(args) == 1 && args[0] == "viewUsers" {
					req := &linuxChatAppPb.ViewListOfUsersRequest{
						UserName: userName,
					}
					resp, err := client.ViewListOfUsers(context.Background(), req)
					if err != nil {
						errChannel <- err
					} else {
						fmt.Println(resp.GetUsers())
					}
				} else if len(args) >= 2 && args[0] == "createChatRoom" {
					if userName != "" {
						if _, ok := clientChatRoomNames[args[1]]; ok {
							fmt.Printf("Chatroom %v already created.\n", args[1])

						} else {
							req := &linuxChatAppPb.CreateChatRoomRequest{
								ChatRoomName: args[1],
								UserName:     userName,
								Users:        append([]string{userName}, args[2:]...),
							}
							resp, err := client.CreateChatRoom(context.Background(), req)
							clientChatRoomNames[resp.GetChatRoomName()] = true
							fmt.Printf("Adding %v to clientChatRoomNames\n", resp.GetChatRoomName())
							if userName == "" {
								userName = resp.GetHostUserName()
							}
							//create connection with chat service
							clientLib.InitConnection(client, chatRoomSessions, userName, resp.GetChatRoomName(), done, errChannel)
							if err != nil {
								errChannel <- err
							} else {
								fmt.Println(resp)
							}
						}
					} else {
						fmt.Println("Must Create username")
					}
				} else if len(args) == 2 && args[0] == "joinChatRoom" {
					if userName != "" {
						if _, ok := clientChatRoomNames[args[1]]; ok {
							fmt.Printf("Chatroom %v already joined.\n", args[1])
						} else {
							req := &linuxChatAppPb.JoinChatRoomRequest{
								UserName:     userName,
								ChatRoomName: args[1],
							}
							resp, err := client.JoinChatRoom(context.Background(), req)
							if resp.GetSuccess() {
								clientChatRoomNames[resp.GetChatRoomName()] = true
								fmt.Printf("Adding %v to clientChatRoomNames\n", resp.GetChatRoomName())
							}
							//create connection with chat service
							clientLib.InitConnection(client, chatRoomSessions, userName, resp.GetChatRoomName(), done, errChannel)
							if err != nil {
								errChannel <- err
							} else {
								fmt.Println(resp)
							}
						}
					} else {
						fmt.Println("Must Create username")
					}
				} else if len(args) == 2 && args[0] == "leaveChatRoom" {
					if userName != "" {
						req := &linuxChatAppPb.LeaveChatRoomRequest{
							UserName:     userName,
							ChatRoomName: args[1],
						}
						resp, err := client.LeaveChatRoom(context.Background(), req)
						if resp.GetSuccess() {
							clientChatRoomNames[resp.GetChatRoomName()] = false
							if _, ok := chatRoomSessions[resp.GetChatRoomName()]; ok {
								chatRoomSessions[resp.GetChatRoomName()].CloseSend()
								delete(chatRoomSessions, resp.GetChatRoomName())
								fmt.Printf("removing %v from chatRoomSessions\n", resp.GetChatRoomName())
							}
							delete(clientChatRoomNames, resp.GetChatRoomName())
							fmt.Printf("removing %v from clientChatRoomNames\n", resp.GetChatRoomName())
						}
						if err != nil {
							errChannel <- err
						} else {
							fmt.Println(resp)
						}
					} else {
						fmt.Println("Must Create username")
					}
				} else if len(args) >= 2 && args[0] == "send" {
					if _, ok := clientChatRoomNames[args[1]]; ok {
						chatRoomSessions[args[1]].Send(&linuxChatAppPb.MessageRequest{
							UserName:     userName,
							ChatRoomName: args[1],
							Message:      strings.Join(args[2:], " "),
							TimeStamp:    utils.GetTimeStamp(),
						})
					} else {
						fmt.Printf("%v is not in chatRoom %v\n", userName, args[1])
					}
				} else if len(args) == 1 && args[0] == "help" {
					fmt.Printf(RootHelpMessage)
				} else if len(args) == 2 && args[0] == "help" {
					if _, ok := Help[args[1]]; ok {
						fmt.Printf(Help[args[1]])
					} else {
						fmt.Println("Command does not exist.")
					}
				}
				if userNameCreated {
					fmt.Printf(userName + ": ")
				}
			}
		}
	}()

}
