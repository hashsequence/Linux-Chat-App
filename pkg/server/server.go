package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
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

func NewLinuxChatServer() *LinuxChatServer {
	return &LinuxChatServer{
		dataStore: data.NewDataStore(),
	}
}

func (this *LinuxChatServer) Serve(caCrt, serverCrt, serverKey, addr string) error {
	// Load the certificates from disk
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

	// Serve and Listen
	if err := srv.Serve(lis); err != nil {
		return fmt.Errorf("grpc serve error: %v", err)
	}

	return nil
}

func (this *LinuxChatServer) CreateUser(ctx context.Context, req *linuxChatAppPb.CreateUserNameRequest) (*linuxChatAppPb.CreateUserNameResponse, error) {

	isCreated := this.dataStore.CreateUser(req.GetUserName())
	resp := &linuxChatAppPb.CreateUserNameResponse{
		UserName: req.GetUserName(),
		Success:  isCreated,
	}
	return resp, nil
}

func (this *LinuxChatServer) DeleteUser(ctx context.Context, req *linuxChatAppPb.DeleteUserRequest) (*linuxChatAppPb.DeleteUserResponse, error) {

	isDeleted := this.dataStore.DeleteUser(req.GetUserName())
	resp := &linuxChatAppPb.DeleteUserResponse{
		UserName: req.GetUserName(),
		Success:  isDeleted,
	}
	return resp, nil
}

func (this *LinuxChatServer) CreateChatRoom(ctx context.Context, req *linuxChatAppPb.CreateChatRoomRequest) (*linuxChatAppPb.CreateChatRoomResponse, error) {

	chatRoom := data.NewChatRoom(req.GetUserName(), req.GetChatRoomName(), req.GetUsers(), false)
	this.dataStore.AddChatRoom(chatRoom)

	resp := &linuxChatAppPb.CreateChatRoomResponse{
		HostUserName: req.GetUserName(),
		ChatRoomName: req.GetChatRoomName(),
	}

	return resp, nil
}

func (this *LinuxChatServer) JoinChatRoom(ctx context.Context, req *linuxChatAppPb.JoinChatRoomRequest) (*linuxChatAppPb.JoinChatRoomResponse, error) {
	this.dataStore.AddUser(req.GetUserName(), req.GetChatRoomName())
	resp := &linuxChatAppPb.JoinChatRoomResponse{
		Success: true,
	}
	return resp, nil
}

func (this *LinuxChatServer) LeaveChatRoom(ctx context.Context, req *linuxChatAppPb.LeaveChatRoomRequest) (*linuxChatAppPb.LeaveChatRoomResponse, error) {

	success := this.dataStore.LeaveChatRoom(req.GetUserName(), req.GetChatRoomName())
	resp := &linuxChatAppPb.LeaveChatRoomResponse{
		Success: success,
	}

	if !success {
		return resp, fmt.Errorf("LeaveChatRoom | %v Failed To Leave ChatRoom %v\n", req.GetUserName(), req.GetChatRoomName())
	}
	return resp, nil
}

func (this *LinuxChatServer) SendMessage(stream linuxChatAppPb.LinuxChatAppService_SendMessageServer) error {
	return nil
}

func (this *LinuxChatServer) ViewListOfUsers(ctx context.Context, req *linuxChatAppPb.ViewListOfUsersRequest) (*linuxChatAppPb.ViewListOfUsersResponse, error) {
	resp := &linuxChatAppPb.ViewListOfUsersResponse{
		Users: this.dataStore.GetUsers(),
	}
	return resp, nil
}

func (this *LinuxChatServer) ViewListOfChatRooms(ctx context.Context, req *linuxChatAppPb.ViewListOfChatRoomsRequest) (*linuxChatAppPb.ViewListOfChatRoomsResponse, error) {
	chatRoomsMap := this.dataStore.GetChatRooms()
	chatRooms := []string{}
	for key := range chatRoomsMap {
		chatRooms = append(chatRooms, key)
	}
	resp := &linuxChatAppPb.ViewListOfChatRoomsResponse{
		ChatRoomNames: chatRooms,
	}
	return resp, nil
}
