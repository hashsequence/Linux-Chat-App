package client 


import (
	"fmt"
	//"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc"
	//"google.golang.org/grpc/status"
	"crypto/x509"
	"crypto/tls"
	"io/ioutil"
	linuxChatAppPb "github.com/hashsequence/Linux-Chat-App/pkg/pb/LinuxChatAppPb"
)

func CreateClient(caCrt, clientCrt, clientKey, addr string) (linuxJobWorkerPb.LinuxJobWorkerServiceClient, error) {

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
	
    conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))//,grpc.WithMaxMsgSize(1024*1024*50))
    if err != nil {
        return nil, fmt.Errorf("could not dial %s: %s", addr, err)
    } 

    // Initialize the client and make the request
	client := linuxJobWorkerPb.NewLinuxJobWorkerServiceClient(conn)
	return client, err
	
}