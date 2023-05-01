package data

import "sync"

const CHATROOM_SIZE = 100

type User struct {
	userName  string
	chatRooms map[string]*ChatRoom
}

type ChatRoom struct {
	chatRoomName string
	users        map[string]*User
	private      bool
	createdBy    string
	messages     [][]string
}

type DataStore struct {
	sync.RWMutex
	users     map[string]*User
	chatRooms map[string]*ChatRoom
}

func (this *DataStore) CreateUser(userName string) bool {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	if _, ok := this.users[userName]; !ok {
		this.users[userName] = &User{
			userName:  userName,
			chatRooms: map[string]*ChatRoom{},
		}
		return true
	}
	return false
}

func (this *DataStore) DeleteUser(userName string) bool {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	if user, ok := this.users[userName]; ok {
		chatRooms := user.chatRooms
		for key := range chatRooms {
			delete(chatRooms[key].users, userName)
		}
		delete(this.users, userName)
		return true
	}
	return false
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

func (this *DataStore) GetChatRooms() map[string]*ChatRoom {
	this.RLock()
	defer func() {
		this.RUnlock()
	}()
	return this.chatRooms
}

func (this *DataStore) AddChatRoom(chatRoom *ChatRoom) {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
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
	//if chatroom does not exist then create it
	if _, ok := this.chatRooms[chatRoomName]; !ok {
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

func (this *DataStore) LeaveChatRoom(userName string, chatRoomName string) bool {
	this.Lock()
	defer func() {
		this.Unlock()
	}()
	success := true
	//if user exist then delete user's chatroom for chatRoomName
	if _, ok := this.users[userName]; ok {
		delete(this.users[userName].chatRooms, chatRoomName)
	} else {
		success = false
	}
	//if user exists and chatRoom exists then delete chatroom's user pointing to userName
	if _, ok := this.chatRooms[chatRoomName]; ok && success {
		delete(this.chatRooms[chatRoomName].users, userName)
	} else {
		success = false
	}
	//if user's chatRoom is 0 then delete user from users
	//if len(this.users[userName].chatRooms) < 1 {
	//	delete(this.users, userName)
	//}
	//if chatRoom has 0 users then delete chatRoom
	if len(this.chatRooms[chatRoomName].users) < 1 {
		delete(this.chatRooms, chatRoomName)
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
	nChatRoom := &ChatRoom{
		chatRoomName: chatRoomName,
		users:        userMap,
		private:      private,
		createdBy:    userName,
		messages:     [][]string{},
	}
	for _, val := range users {
		userMap[val] = &User{val, map[string]*ChatRoom{
			chatRoomName: nChatRoom,
		},
		}
	}
	return nChatRoom
}

func NewDataStore() *DataStore {
	return &DataStore{
		users:     map[string]*User{},
		chatRooms: map[string]*ChatRoom{},
	}
}
