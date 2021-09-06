package streaming

import (
	"sync"
	"time"
)

type rateLimitManager struct {
	rateLimitPerSecond          int
	clientRequestStatistics     map[RemoteAddress]int
	mx                          *sync.RWMutex
	channelConnectionCloseEvent chan RemoteAddress
}

func newRateLimitManager(rateLimitPerSecond int, poolSizeClients int) *rateLimitManager {
	this := new(rateLimitManager)
	this.mx = new(sync.RWMutex)
	this.rateLimitPerSecond = rateLimitPerSecond
	this.clientRequestStatistics = make(map[RemoteAddress]int)
	this.channelConnectionCloseEvent = make(chan RemoteAddress, poolSizeClients)
	go this.checkClientStatistics()
	return this
}

func (this *rateLimitManager) startNewClientStatistic(clientRemoteAddress RemoteAddress) {
	this.mx.Lock()
	defer this.mx.Unlock()
	this.clientRequestStatistics[clientRemoteAddress] = 0
}

func (this *rateLimitManager) updateClientStatistic(clientRemoteAddress RemoteAddress) {
	this.mx.Lock()
	defer this.mx.Unlock()
	if stat, exist := this.clientRequestStatistics[clientRemoteAddress]; !exist {
		return
	} else {
		this.clientRequestStatistics[clientRemoteAddress] = stat + 1
	}
}

func (this *rateLimitManager) deleteClientStatistic(clientRemoteAddress RemoteAddress) {
	this.mx.Lock()
	defer this.mx.Unlock()
	if _, exist := this.clientRequestStatistics[clientRemoteAddress]; !exist {
		return
	} else {
		delete(this.clientRequestStatistics, clientRemoteAddress)
	}
}

func (this *rateLimitManager) checkClientStatistics() {
	var (
		ticker = time.NewTicker(time.Second)
	)
	for {
		select {
		case <-ticker.C:
			for client, stat := range this.clientRequestStatistics {
				if stat >= this.rateLimitPerSecond {
					this.channelConnectionCloseEvent <- client
				} else {
					this.clientRequestStatistics[client] = 0
				}
			}
		}
	}
}
