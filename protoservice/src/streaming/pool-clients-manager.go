package streaming

import "sync"

type clientTuple struct {
	Address RemoteAddress
	Client  *client
}

type poolClientsManager struct {
	size int
	pool map[RemoteAddress]*client
	mx   *sync.RWMutex
}

func newPoolClientsManager(size int) *poolClientsManager {
	this := new(poolClientsManager)
	this.size = size
	this.mx = new(sync.RWMutex)
	this.pool = make(map[RemoteAddress]*client)
	return this
}

func (this *poolClientsManager) length() int {
	this.mx.RLock()
	defer this.mx.RUnlock()
	return len(this.pool)
}

func (this *poolClientsManager) isFilled() bool {
	if this.length()+1 >= this.size {
		return true
	}
	return false
}

func (this *poolClientsManager) push(client *client) error {
	if this.isFilled() {
		return ErrorPoolClientIsFilled
	}
	if client == nil {
		return ErrorClientObjectIsNil
	}
	if client.websocketRemoteAddress == RemoteAddress("") {
		return ErrorClientObjectIsNil
	}
	this.mx.Lock()
	defer this.mx.Unlock()
	this.pool[client.websocketRemoteAddress] = client
	return nil
}

func (this *poolClientsManager) Get(remoteAddress RemoteAddress) (*client, error) {
	this.mx.RLock()
	defer this.mx.RUnlock()
	if client, exist := this.pool[remoteAddress]; !exist {
		return nil, ErrorClientObjectIsNil
	} else {
		return client, nil
	}
}

func (this *poolClientsManager) delete(remoteAddress RemoteAddress) error {
	this.mx.Lock()
	defer this.mx.Unlock()
	if _, exist := this.pool[remoteAddress]; !exist {
		return ErrorClientObjectIsNil
	}
	delete(this.pool, remoteAddress)
	return nil
}

func (this *poolClientsManager) Iterate() <-chan clientTuple {
	var (
		channel = make(chan clientTuple, this.length())
		write   = func(channel chan clientTuple, this *poolClientsManager) {
			for remoteAddress, client := range this.pool {
				channel <- clientTuple{
					Address: remoteAddress,
					Client:  client,
				}
			}
			close(channel)
		}
	)
	go write(channel, this)
	return channel
}
