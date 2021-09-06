package fileservice

import (
	"errors"
	"sync"
)

type (
	uuidCode string
)

type poolSessionManager struct {
	mx   *sync.RWMutex
	pool map[uuidCode]*session
}

func newPoolSessionManager() *poolSessionManager {
	this := new(poolSessionManager)
	this.mx = new(sync.RWMutex)
	this.pool = make(map[uuidCode]*session)
	return this
}

func (this *poolSessionManager) push(session *session) error {
	this.mx.Lock()
	defer this.mx.Unlock()
	if session, exist := this.pool[uuidCode(session.sessionUUID.String())]; exist {
		return errors.New("Session [" + session.sessionUUID.String() + "] is exist with client [" + string(session.remoteClientAddress) + "]")
	} else {
		this.pool[uuidCode(session.sessionUUID.String())] = session
		return nil
	}
}

func (this *poolSessionManager) get(uuidCode uuidCode) (*session, error) {
	this.mx.RLock()
	defer this.mx.RUnlock()
	if session, exist := this.pool[uuidCode]; exist {
		return session, nil
	} else {
		return nil, errors.New("Session none exist.")
	}
}

func (this *poolSessionManager) delete(uuidCode uuidCode) error {
	this.mx.Lock()
	defer this.mx.Unlock()
	if _, exist := this.pool[uuidCode]; exist {
		delete(this.pool, uuidCode)
		return nil
	} else {
		return errors.New("Session none exist.")
	}
}
