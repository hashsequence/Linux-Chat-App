package data

import "sync"

type User struct {
	userName string
}

type ChatRoom struct {
	chatRoomName string
	users        []string
	private      bool
	createdBy    string
}

type DataStore struct {
	sync.RWMutex
	users     map[string]*User
	chatRooms map[string]*ChatRoom
}

func (this *DataStore) GetUsers() map[string]*User {
	this.RLock()
	defer func() {
		this.RUnlock()
	}()
	return this.users
}

func (this *DataStore) AddUser(user *User) bool {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	if _, ok := this.users[user.userName]; !ok {
		this.users[user.userName] = user
		return true
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
	if _, ok := this.chatRooms[chatRoom.chatRoomName]; !ok {
		this.chatRooms[chatRoom.chatRoomName] = chatRoom
		return true
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

func NewUser(userName string) *User {
	return &User{
		userName: userName,
	}
}

func NewChatRoom(userName string, chatRoomName string, users []string, private bool) *ChatRoom {
	return &ChatRoom{
		chatRoomName: chatRoomName,
		users:        users,
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
