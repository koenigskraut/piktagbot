package main

import (
	"sync"
	"time"
)

type session struct {
	userID     int64
	clientHash string
}

type sessionMap struct {
	m map[string]*session
	sync.RWMutex
}

var oneTimeSessions = sessionMap{m: make(map[string]*session)}

const sessionExpirationLimit = time.Minute * 10

func (sm *sessionMap) createSession(userID int64, hash string, clientHash string) {
	session := &session{
		userID:     userID,
		clientHash: clientHash,
	}
	sm.Lock()
	defer sm.Unlock()
	if _, ok := sm.m[hash]; ok {
		return
	}
	sm.m[hash] = session
	go func(s *sessionMap, h string) {
		time.Sleep(sessionExpirationLimit)
		s.delete(h)
	}(sm, hash)
}

func (sm *sessionMap) peek(hash string) (*session, bool) {
	sm.RLock()
	session, ok := sm.m[hash]
	sm.RUnlock()
	return session, ok
}

func (sm *sessionMap) delete(hash string) {
	sm.Lock()
	delete(sm.m, hash)
	sm.Unlock()
}
