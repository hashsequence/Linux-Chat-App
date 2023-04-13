package server

import (
	"fmt"
	linuxChatAppPb "github.com/hashsequence/Linux-Chat-App/pkg/pb"
	"crypto/tls"
	"crypto/x509"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"net"	
)

type LinuxChatServer struct {

}

func NewLinuxChatServer() *LinuxChatServer {
	return &LinuxChatServer{}
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
	linuxChatAppPb.RegisterLinuxChatAppServiceServer(srv, s)

	// Serve and Listen
	if err := srv.Serve(lis); err != nil {
		return fmt.Errorf("grpc serve error: %v", err)
	}

	return nil
}