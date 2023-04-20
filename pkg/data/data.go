package data

import(
	"sync"
	"fmt"
	"path/filepath"
	linuxChatAppPb "github.com/hashsequence/Linux-Job-Worker/pkg/pb"
)

type User struct {
	userName string
}

type ChatRoom struct {
	hostUserName string
	chatRoomName string
	chatRoomId string
}

type DataStore {
	users map[string]*User
	chatRooms map[string]*ChatRoom
}