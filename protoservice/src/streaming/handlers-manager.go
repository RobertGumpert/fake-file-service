package streaming

import "sync"

type handlersManager struct {
	handlers     map[URI]Handler
	mx           *sync.RWMutex
}

func newHandlersManager() *handlersManager {
	this := new(handlersManager)
	this.handlers = make(map[URI]Handler, 0)
	this.mx = new(sync.RWMutex)
	return this
}

func (this *handlersManager) registerHandler(uri URI, handle Handler) {
	this.mx.Lock()
	defer this.mx.Unlock()
	if _, exist := this.handlers[uri]; exist {
		return
	}
	this.handlers[uri] = handle
}

func (this *handlersManager) getHandler(uri URI) (Handler, error) {
	this.mx.RLock()
	defer this.mx.RUnlock()
	if handle, exist := this.handlers[uri]; !exist {
		return nil, ErrorHandlerIsntExist
	} else {
		return handle, nil
	}
}

