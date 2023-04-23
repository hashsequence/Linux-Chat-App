package data

import "sync"

const CHATROOM_SIZE = 100

type User struct {
	userName string
}

type ChatRoom struct {
	chatRoomName string
	users        map[string]*User
	private      bool
	createdBy    string
}

type DataStore struct {
	sync.RWMutex
	users     map[string]*User
	chatRooms map[string]*ChatRoom
}

func (this *DataStore) GetUsers() []string {
	this.RLock()
	defer func() {
		this.RUnlock()
	}()
	usersArr := make([]string, len(this.users))
	i := 0
	for _, val := range this.users {
		usersArr[i] = val.userName
		i++
	}
	return usersArr
}

func (this *DataStore) AddUser(userName string, chatRoomName string) bool {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	success := true
	if _, ok := this.users[userName]; !ok {
		this.users[userName] = &User{userName}
	} else {
		success = false
	}
	if _, ok := this.chatRooms[chatRoomName]; !ok && success {
		this.chatRooms[chatRoomName].users[userName] = this.users[userName]
		success = true
	} else {
		success = false
	}
	return false
}

func (this *DataStore) DeleteUser(user *User) bool {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	if _, ok := this.users[user.userName]; !ok {
		delete(this.users, user.userName)
		return true
	}
	return false
}

func (this *DataStore) GetChatRooms() map[string]*ChatRoom {
	this.RLock()
	defer func() {
		this.RUnlock()
	}()
	return this.chatRooms
}

func (this *DataStore) AddChatRoom(chatRoom *ChatRoom) bool {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	if len(this.chatRooms) < CHATROOM_SIZE {
		if _, ok := this.chatRooms[chatRoom.chatRoomName]; !ok {
			this.chatRooms[chatRoom.chatRoomName] = chatRoom
			return true
		}
	}
	return false
}

func (this *DataStore) DeleteChatRoom(chatRoom *ChatRoom) bool {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	if _, ok := this.chatRooms[chatRoom.chatRoomName]; !ok {
		delete(this.chatRooms, chatRoom.chatRoomName)
		return true
	}
	return false
}

func (this *DataStore) LeaveChatRoom(userName string, chatRoomName string) bool {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	success := true
	if _, ok := this.chatRooms[chatRoomName]; ok {
		delete(this.chatRooms[chatRoomName].users, userName)
		success = true
	} else {
		success = false
	}
	if _, ok := this.users[userName]; ok && success {
		delete(this.users, userName)
		success = true
	} else {
		success = false
	}
	return success
}

func NewUser(userName string) *User {
	return &User{
		userName: userName,
	}
}

func NewChatRoom(userName string, chatRoomName string, users []string, private bool) *ChatRoom {
	userMap := map[string]*User{}
	for _, val := range users {
		userMap[val] = &User{val}
	}
	return &ChatRoom{
		chatRoomName: chatRoomName,
		users:        userMap,
		private:      private,
		createdBy:    userName,
	}
}

func NewDataStore() *DataStore {
	return &DataStore{
		users:     map[string]*User{},
		chatRooms: map[string]*ChatRoom{},
	}
}
