package commands

import (
	"github.com/gotd/td/tg"
	db "github.com/koenigskraut/piktagbot/database"
	"sync"
)

type UserUnderLock struct {
	sync.Mutex
	DBUser *db.User
}

type MessageSemaphore struct {
	sync.Mutex
	data map[int64]*UserUnderLock
}

func NewMessageSemaphore() MessageSemaphore {
	return MessageSemaphore{
		data: map[int64]*UserUnderLock{},
	}
}

func (ms *MessageSemaphore) GetCurrentLock(userID int64) (lockedUser *UserUnderLock) {
	ms.Lock()
	if v, ok := ms.data[userID]; !ok {
		lockedUser = &UserUnderLock{}
		ms.data[userID] = lockedUser
	} else {
		lockedUser = v
	}
	ms.Unlock()
	return
}

func (ms *MessageSemaphore) MessageUserFromUpdate(update *tg.UpdateNewMessage) (*tg.Message, *db.User) {
	m := update.Message.(*tg.Message)
	userID := m.PeerID.(*tg.PeerUser).UserID
	user := ms.GetCurrentLock(userID).DBUser
	return m, user
}
