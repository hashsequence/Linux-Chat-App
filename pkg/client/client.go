package client

import (
	"context"
	"fmt"
	"io"

	//"google.golang.org/grpc/codes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	//"google.golang.org/grpc/status"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	linuxChatAppPb "github.com/hashsequence/Linux-Chat-App/pkg/pb/LinuxChatAppPb"
	"github.com/hashsequence/Linux-Chat-App/pkg/utils"
)

func recieveNewMessages(stream linuxChatAppPb.LinuxChatAppService_SendMessageClient, done chan struct{}, errChannel chan<- error) {

	res, err := stream.Recv()
	if err == io.EOF {
		errChannel <- err
		return
	}
	if err != nil {
		fmt.Printf("error while recieving %v\n", err)
		errChannel <- err
		return
	}
	fmt.Println(res.GetChatRoomName() + " | " + res.GetChatRow())

}

func InitConnection(client linuxChatAppPb.LinuxChatAppServiceClient, chatRoomSessions map[string](linuxChatAppPb.LinuxChatAppService_SendMessageClient), userName string, chatRoomName string, done chan struct{}, errChannel chan<- error) {
	//create connection with chat service
	if _, ok := chatRoomSessions[userName]; !ok {
		fmt.Printf("Creating stream for chatroom: %v\n", chatRoomName)
		stream, err := client.SendMessage(context.Background())
		if err != nil {
			fmt.Println(fmt.Errorf("error while calling SendMessage RPC: %v\n", err))
			errChannel <- err
			return
		}
		chatRoomSessions[chatRoomName] = stream

		//send initial message to register (chatRoom,user) with message queue in chat service
		stream.Send(&linuxChatAppPb.MessageRequest{
			UserName:     userName,
			ChatRoomName: chatRoomName,
			Message:      userName + " joined " + chatRoomName + "...",
			TimeStamp:    utils.GetTimeStamp(),
		})
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					recieveNewMessages(stream, done, errChannel)
				}
			}
		}()
	}
}

func CreateClient(caCrt, clientCrt, clientKey, addr string) (linuxChatAppPb.LinuxChatAppServiceClient, error) {

	// Load the client certificates from disk
	certificate, err := tls.LoadX509KeyPair(clientCrt, clientKey)
	if err != nil {
		return nil, fmt.Errorf("could not load client key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caCrt)
	if err != nil {
		return nil, fmt.Errorf("could not read ca certificate: %s", err)
	}

	// Append the certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, fmt.Errorf("failed to append ca certs")
	}

	creds := credentials.NewTLS(&tls.Config{
		ServerName:   addr, // NOTE: this is required!
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	// Create a connection with the TLS credentials

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds)) //,grpc.WithMaxMsgSize(1024*1024*50))
	if err != nil {
		return nil, fmt.Errorf("could not dial %s: %s", addr, err)
	}

	// Initialize the client and make the request
	client := linuxChatAppPb.NewLinuxChatAppServiceClient(conn)
	return client, err

}
