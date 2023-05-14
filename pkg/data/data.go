package data

import (
	"fmt"
	"sync"
	"time"

	linuxChatAppPb "github.com/hashsequence/Linux-Chat-App/pkg/pb/LinuxChatAppPb"
	"github.com/hashsequence/Linux-Chat-App/pkg/utils"
)

const CHATROOM_SIZE = 100

type User struct {
	userName  string
	chatRooms map[string]*ChatRoom
}

type Message struct {
	msg          string
	chatRoomName string
	userName     string
	timeStamp    string
}
type ChatRoom struct {
	chatRoomName string
	users        map[string]*User
	private      bool
	createdBy    string
	messages     []Message
}

type DataStore struct {
	sync.RWMutex
	users        map[string]*User
	chatRooms    map[string]*ChatRoom
	messageQueue chan Message
	streams      map[string]map[string]linuxChatAppPb.LinuxChatAppService_SendMessageServer
	ttl          int64
	Done         chan struct{}
	afk          map[string]bool
}

func (this *Message) GetMsg() string {
	return this.msg
}

func (this *Message) GetChatRoomName() string {
	return this.chatRoomName
}

func (this *Message) GetUserName() string {
	return this.userName
}

func (this *Message) GetTimeStamp() string {
	return this.timeStamp
}

func (this *Message) ToString() string {
	return this.userName + " | " + this.msg + " | " + this.timeStamp
}

func (this *DataStore) AddMessage(stream linuxChatAppPb.LinuxChatAppService_SendMessageServer, chatRoomName string, userName string, msg string, timeStamp string) *Message {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	if _, ok := this.users[userName]; !ok {
		fmt.Printf("%v does not exist\n", userName)
		return nil
	}
	delete(this.afk, userName)
	//check if user is in chatRoom
	if _, ok := this.chatRooms[chatRoomName]; !ok {
		fmt.Printf("Chatroom %v does not exist.\n", chatRoomName)
		return nil
	}
	if _, ok := this.chatRooms[chatRoomName].users[userName]; !ok {
		fmt.Println(userName + " is not in chatRoom " + chatRoomName + ".")
		return nil
	}
	var newMsg *Message
	if _, ok := this.streams[chatRoomName]; !ok {
		this.streams[chatRoomName] = map[string]linuxChatAppPb.LinuxChatAppService_SendMessageServer{}
	}
	if _, ok := this.streams[chatRoomName][userName]; !ok {
		this.streams[chatRoomName][userName] = stream
	}
	fmt.Printf("%v | %v | %v | %v\n", chatRoomName, userName, msg, timeStamp)
	if chatRoom, ok := this.chatRooms[chatRoomName]; ok {
		newMsg = &Message{
			msg:          msg,
			chatRoomName: chatRoomName,
			userName:     userName,
			timeStamp:    timeStamp,
		}
		chatRoom.messages = append(chatRoom.messages)
		this.messageQueue <- *newMsg
	}
	return newMsg
}

func (this *DataStore) addMessage(stream linuxChatAppPb.LinuxChatAppService_SendMessageServer, chatRoomName string, userName string, msg string, timeStamp string) *Message {
	var newMsg *Message
	if _, ok := this.streams[chatRoomName]; !ok {
		this.streams[chatRoomName] = map[string]linuxChatAppPb.LinuxChatAppService_SendMessageServer{}
	}
	if _, ok := this.streams[chatRoomName][userName]; !ok {
		this.streams[chatRoomName][userName] = stream
	}
	fmt.Printf("%v | %v | %v | %v\n", chatRoomName, userName, msg, timeStamp)
	if chatRoom, ok := this.chatRooms[chatRoomName]; ok {
		newMsg = &Message{
			msg:          msg,
			chatRoomName: chatRoomName,
			userName:     userName,
			timeStamp:    timeStamp,
		}
		chatRoom.messages = append(chatRoom.messages)
		this.messageQueue <- *newMsg
	}
	return newMsg
}

func (this *DataStore) CreateUser(userName string) {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	delete(this.afk, userName)
	if _, ok := this.users[userName]; !ok {
		this.users[userName] = &User{
			userName:  userName,
			chatRooms: map[string]*ChatRoom{},
		}
	}
}

func (this *DataStore) DeleteUser(userName string) bool {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	if user, ok := this.users[userName]; ok {
		chatRooms := user.chatRooms
		for key := range chatRooms {
			fmt.Printf("force %v to leave %v.\n", userName, key)
			this.leaveChatRoom(userName, key)
		}
		delete(this.users, userName)
		return true
	}
	return false
}

func (this *DataStore) deleteUser(userName string) bool {
	if user, ok := this.users[userName]; ok {
		chatRooms := user.chatRooms
		for key := range chatRooms {
			fmt.Printf("force %v to leave %v.\n", userName, key)
			this.leaveChatRoom(userName, key)
		}
		delete(this.users, userName)
		return true
	}
	return false
}

func (this *DataStore) GetUsers(clientUserName string) []string {
	this.RLock()
	defer func() {
		this.RUnlock()
	}()
	delete(this.afk, clientUserName)
	usersArr := make([]string, len(this.users))
	i := 0
	for _, val := range this.users {
		usersArr[i] = val.userName
		i++
	}
	return usersArr
}

func (this *DataStore) GetChatRooms(clientUserName string) map[string]*ChatRoom {
	this.RLock()
	defer func() {
		this.RUnlock()
	}()
	delete(this.afk, clientUserName)
	return this.chatRooms
}

func (this *DataStore) AddChatRoom(clientUserName string, chatRoom *ChatRoom) {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	delete(this.afk, clientUserName)
	if len(this.chatRooms) < CHATROOM_SIZE {
		if _, ok := this.chatRooms[chatRoom.chatRoomName]; !ok {
			this.chatRooms[chatRoom.chatRoomName] = chatRoom
		}
	}
}

func (this *DataStore) AddUser(userName string, chatRoomName string) {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	delete(this.afk, userName)
	//if chatroom does not exist then create it

	if _, ok := this.chatRooms[chatRoomName]; !ok {
		if len(this.chatRooms) >= CHATROOM_SIZE {
			fmt.Printf("max number of chatrooms has reached max capacity at %v.\n", CHATROOM_SIZE)
			return
		}
		this.chatRooms[chatRoomName] = NewChatRoom(userName, chatRoomName, []string{userName}, false)
	}
	//if username does not exist create user and set chatRooms of user to the chatRoom and set chatRoom to contain the user
	if _, ok := this.users[userName]; !ok {
		this.users[userName] = &User{
			userName: userName,
			chatRooms: map[string]*ChatRoom{
				chatRoomName: this.chatRooms[chatRoomName],
			},
		}
		this.chatRooms[chatRoomName].users[userName] = this.users[userName]
	} else {
		//if user exists then set current chatroom to contain the user, and user's chatRooms to contain current chatRoom
		this.chatRooms[chatRoomName].users[userName] = this.users[userName]
		this.users[userName].chatRooms[chatRoomName] = this.chatRooms[chatRoomName]
	}
}

func (this *DataStore) leaveChatRoom(userName string, chatRoomName string) bool {
	success := true
	//if user exist then delete user's chatroom for chatRoomName
	if _, ok := this.users[userName]; ok {
		fmt.Printf("Deleting %v from %v's ChatRooms.\n", chatRoomName, userName)
		delete(this.users[userName].chatRooms, chatRoomName)
	} else {
		success = false
	}
	//if user exists and chatRoom exists then delete chatroom's user pointing to userName
	if _, ok := this.chatRooms[chatRoomName]; ok && success {
		fmt.Printf("Deleting %v from chatRoom %v.\n", userName, chatRoomName)
		delete(this.chatRooms[chatRoomName].users, userName)
	} else {
		success = false
	}

	if _, ok := this.streams[chatRoomName]; ok && success {
		if _, ok2 := this.streams[chatRoomName][userName]; ok2 {
			fmt.Printf("Removing user: %v from chatroom: %v\n", userName, chatRoomName)
			this.addMessage(this.streams[chatRoomName][userName], chatRoomName, userName, userName+" leaving "+chatRoomName+"...", utils.GetTimeStamp())
			delete(this.streams[chatRoomName], userName)
		}
	} else {
		success = false
	}
	//if user's chatRoom is 0 then delete user from users
	//if len(this.users[userName].chatRooms) < 1 {
	//	delete(this.users, userName)
	//}
	//if chatRoom has 0 users then delete chatRoom
	if len(this.chatRooms[chatRoomName].users) < 1 {
		fmt.Printf("Chatroom %v has no more users, deleting chatRoom...\n", chatRoomName)
		delete(this.chatRooms, chatRoomName)
		if _, ok := this.streams[chatRoomName]; ok {
			fmt.Printf("Chatroom %v has no more users, deleting chatRoom streams...\n", chatRoomName)
			delete(this.streams, chatRoomName)
		}
	}

	return success
}

func (this *DataStore) LeaveChatRoom(userName string, chatRoomName string) bool {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	delete(this.afk, userName)
	success := true
	//if user exist then delete user's chatroom for chatRoomName
	if _, ok := this.users[userName]; ok {
		fmt.Printf("Deleting %v from %v's ChatRooms.\n", chatRoomName, userName)
		delete(this.users[userName].chatRooms, chatRoomName)
	} else {
		success = false
	}
	//if user exists and chatRoom exists then delete chatroom's user pointing to userName
	if _, ok := this.chatRooms[chatRoomName]; ok && success {
		fmt.Printf("Deleting %v from chatRoom %v.\n", userName, chatRoomName)
		delete(this.chatRooms[chatRoomName].users, userName)
	} else {
		success = false
	}

	if _, ok := this.streams[chatRoomName]; ok && success {
		if _, ok2 := this.streams[chatRoomName][userName]; ok2 {
			fmt.Printf("Removing user: %v from chatroom: %v\n", userName, chatRoomName)
			this.addMessage(this.streams[chatRoomName][userName], chatRoomName, userName, userName+" leaving "+chatRoomName+"...", utils.GetTimeStamp())
			delete(this.streams[chatRoomName], userName)
		}
	} else {
		success = false
	}
	//if user's chatRoom is 0 then delete user from users
	//if len(this.users[userName].chatRooms) < 1 {
	//	delete(this.users, userName)
	//}
	//if chatRoom has 0 users then delete chatRoom
	if len(this.chatRooms[chatRoomName].users) < 1 {
		fmt.Printf("Chatroom %v has no more users, deleting chatRoom...\n", chatRoomName)
		delete(this.chatRooms, chatRoomName)
		if _, ok := this.streams[chatRoomName]; ok {
			fmt.Printf("Chatroom %v has no more users, deleting chatRoom streams...\n", chatRoomName)
			delete(this.streams, chatRoomName)
		}
	}

	return success
}

func (this *DataStore) SendOutMessages() {
	for {
		select {
		case <-this.Done:
			return
		case msg := <-this.messageQueue:
			fmt.Printf("Sending out Message %v\n", msg)
			for userName, stream := range this.streams[msg.GetChatRoomName()] {
				//send out messages to Chatroom if msg does not belong to sender
				if userName != msg.userName {
					stream.Send(&linuxChatAppPb.MessageResponse{
						ChatRoomName: msg.GetChatRoomName(),
						ChatRow:      msg.ToString(),
					})
				}
			}
		}
	}
}

func (this *DataStore) UserExists(userName string) bool {
	this.RLock()
	defer func() {
		this.RUnlock()
	}()
	if _, ok := this.users[userName]; ok {
		return true
	}
	return false
}

func NewUser(userName string) *User {
	return &User{
		userName: userName,
	}
}

func NewChatRoom(userName string, chatRoomName string, users []string, private bool) *ChatRoom {
	userMap := map[string]*User{}
	nChatRoom := &ChatRoom{
		chatRoomName: chatRoomName,
		users:        userMap,
		private:      private,
		createdBy:    userName,
		messages:     []Message{},
	}
	for _, val := range users {
		userMap[val] = &User{val, map[string]*ChatRoom{
			chatRoomName: nChatRoom,
		},
		}
	}
	return nChatRoom
}

func (this *DataStore) refreshAfkList() {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	for key := range this.afk {
		fmt.Printf("Deleting user %v since user has been afk for %v seconds\n", key, this.ttl)
		this.deleteUser(key)
		delete(this.afk, key)
	}
	//add users to new cycle
	for key := range this.users {
		this.afk[key] = true
	}
}

func (this *DataStore) initRoutines() {
	fmt.Println("datastore initRoutines...")
	ticker := time.NewTicker(time.Duration(this.ttl) * time.Second)
	go func() {

		for {
			select {
			case <-ticker.C:
				this.refreshAfkList()
			case <-this.Done:
				fmt.Println("Shutting down routines")
				ticker.Stop()
				return
			}
		}
	}()
}

func NewDataStore(messageQueueSize int, ttl_Sec int64, done chan struct{}) *DataStore {

	ds := &DataStore{
		users:        map[string]*User{},
		chatRooms:    map[string]*ChatRoom{},
		messageQueue: make(chan Message, messageQueueSize),
		streams:      map[string]map[string]linuxChatAppPb.LinuxChatAppService_SendMessageServer{},
		ttl:          ttl_Sec,
		Done:         done,
		afk:          map[string]bool{},
	}
	ds.initRoutines()
	return ds
}
