mod:
	go mod init github.com/hashsequence/Linux-Chat-App

mod-tidy:
	go mod tidy

run-server:
	go run cmd/server/serverMain.go

run-client:
	go run cmd/client/clientMain.go