package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"

	"github.com/hashsequence/Linux-Chat-App/pkg/data"
	linuxChatAppPb "github.com/hashsequence/Linux-Chat-App/pkg/pb/LinuxChatAppPb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type LinuxChatServer struct {
	linuxChatAppPb.UnimplementedLinuxChatAppServiceServer
	dataStore *data.DataStore
}

func NewLinuxChatServer(messageQueueSize int, ttl_Sec int64, done chan struct{}) *LinuxChatServer {
	return &LinuxChatServer{
		dataStore: data.NewDataStore(messageQueueSize, ttl_Sec, done),
	}
}

func (this *LinuxChatServer) Serve(caCrt, serverCrt, serverKey, addr string) error {
	// Load the certificates from disk
	defer func() {
		close(this.dataStore.Done)
	}()

	certificate, err := tls.LoadX509KeyPair(serverCrt, serverKey)
	if err != nil {
		return fmt.Errorf("could not load server key pair: %v", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caCrt)
	if err != nil {
		return fmt.Errorf("could not read ca certificate: %v", err)
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return fmt.Errorf("failed to append ca certs")
	}

	// Create the channel to listen on
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("could not list on %v: %v", addr, err)
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		//InsecureSkipVerify: true,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	// Create the gRPC server with the credentials
	srv := grpc.NewServer(grpc.Creds(creds))

	// Register the handler object
	linuxChatAppPb.RegisterLinuxChatAppServiceServer(srv, this)
	//run thread to send out messages
	go this.dataStore.SendOutMessages()
	// Serve and Listen
	if err := srv.Serve(lis); err != nil {
		return fmt.Errorf("grpc serve error: %v", err)
	}

	return nil
}

func (this *LinuxChatServer) CreateUser(ctx context.Context, req *linuxChatAppPb.CreateUserNameRequest) (*linuxChatAppPb.CreateUserNameResponse, error) {
	fmt.Printf("CreateUser requested for %v\n", req.GetUserName())
	this.dataStore.CreateUser(req.GetUserName())
	resp := &linuxChatAppPb.CreateUserNameResponse{
		UserName: req.GetUserName(),
	}
	return resp, nil
}

func (this *LinuxChatServer) DeleteUser(ctx context.Context, req *linuxChatAppPb.DeleteUserRequest) (*linuxChatAppPb.DeleteUserResponse, error) {
	fmt.Printf("DeleteUser requested for %v\n", req.GetUserName())
	isDeleted := this.dataStore.DeleteUser(req.GetUserName())
	resp := &linuxChatAppPb.DeleteUserResponse{
		UserName: req.GetUserName(),
		Success:  isDeleted,
	}
	return resp, nil
}

func (this *LinuxChatServer) CreateChatRoom(ctx context.Context, req *linuxChatAppPb.CreateChatRoomRequest) (*linuxChatAppPb.CreateChatRoomResponse, error) {
	fmt.Printf("CreateChatRoom %v requested for %v\n", req.GetChatRoomName(), req.GetUserName())
	chatRoom := data.NewChatRoom(req.GetUserName(), req.GetChatRoomName(), req.GetUsers(), false)
	this.dataStore.AddChatRoom(req.GetUserName(), chatRoom)

	resp := &linuxChatAppPb.CreateChatRoomResponse{
		HostUserName: req.GetUserName(),
		ChatRoomName: req.GetChatRoomName(),
	}

	return resp, nil
}

func (this *LinuxChatServer) JoinChatRoom(ctx context.Context, req *linuxChatAppPb.JoinChatRoomRequest) (*linuxChatAppPb.JoinChatRoomResponse, error) {
	fmt.Printf("%v is joining %v\n", req.GetUserName(), req.GetChatRoomName())
	this.dataStore.AddUser(req.GetUserName(), req.GetChatRoomName())
	resp := &linuxChatAppPb.JoinChatRoomResponse{
		UserName:     req.GetUserName(),
		ChatRoomName: req.GetChatRoomName(),
		Success:      true,
	}
	return resp, nil
}

func (this *LinuxChatServer) LeaveChatRoom(ctx context.Context, req *linuxChatAppPb.LeaveChatRoomRequest) (*linuxChatAppPb.LeaveChatRoomResponse, error) {
	fmt.Printf("%v is leaving %v\n", req.GetUserName(), req.GetChatRoomName())
	success := this.dataStore.LeaveChatRoom(req.GetUserName(), req.GetChatRoomName())
	resp := &linuxChatAppPb.LeaveChatRoomResponse{
		UserName:     req.GetUserName(),
		ChatRoomName: req.GetChatRoomName(),
		Success:      success,
	}

	if !success {
		return resp, fmt.Errorf("LeaveChatRoom | %v Failed To Leave ChatRoom %v\n", req.GetUserName(), req.GetChatRoomName())
	}
	return resp, nil
}

func (this *LinuxChatServer) SendMessage(stream linuxChatAppPb.LinuxChatAppService_SendMessageServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error while reading client stream %v\n", err)
		}

		this.dataStore.AddMessage(stream, req.GetChatRoomName(), req.GetUserName(), req.GetMessage(), req.GetTimeStamp())

		//senderErr := stream.Send(&linuxChatAppPb.MessageResponse{
		//	ChatRoomName: req.GetChatRoomName(),
		//	ChatRow:      newMsg.ToString(),
		//})
		//if senderErr != nil {
		//	return fmt.Errorf("Error while sending data to cclient: %v\n", err)
		//}
	}
}

func (this *LinuxChatServer) ViewListOfUsers(ctx context.Context, req *linuxChatAppPb.ViewListOfUsersRequest) (*linuxChatAppPb.ViewListOfUsersResponse, error) {
	resp := &linuxChatAppPb.ViewListOfUsersResponse{
		Users: this.dataStore.GetUsers(req.GetUserName()),
	}
	return resp, nil
}

func (this *LinuxChatServer) ViewListOfChatRooms(ctx context.Context, req *linuxChatAppPb.ViewListOfChatRoomsRequest) (*linuxChatAppPb.ViewListOfChatRoomsResponse, error) {
	chatRoomsMap := this.dataStore.GetChatRooms(req.GetUserName())
	chatRooms := []string{}
	for key := range chatRoomsMap {
		chatRooms = append(chatRooms, key)
	}
	resp := &linuxChatAppPb.ViewListOfChatRoomsResponse{
		ChatRoomNames: chatRooms,
	}
	return resp, nil
}

func (this *LinuxChatServer) Ping(ctx context.Context, req *linuxChatAppPb.PingRequest) (*linuxChatAppPb.PingResponse, error) {
	resp := &linuxChatAppPb.PingResponse{}
	return resp, nil
}
