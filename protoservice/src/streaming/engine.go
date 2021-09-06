package streaming

import (
	"fmt"
	"log"
	"net/http"

	"google.golang.org/protobuf/proto"
)

type Engine struct {
	PoolClients  *poolClientsManager
	poolHandlers *handlersManager
	rateLimiter  *rateLimitManager
}

func NewEngine(poolSizeClients int, rateLimitPerSecond int) *Engine {
	this := new(Engine)
	this.PoolClients = newPoolClientsManager(poolSizeClients)
	this.poolHandlers = newHandlersManager()
	this.rateLimiter = newRateLimitManager(rateLimitPerSecond, poolSizeClients)
	go this.waitRateLimiterEvents()
	return this
}

func (this *Engine) Handle(uri URI, handle Handler) {
	this.poolHandlers.registerHandler(uri, handle)
}

func (this *Engine) SendMessageClient(clientRemoteAddress RemoteAddress, message []byte) error {
	client, err := this.PoolClients.Get(clientRemoteAddress)
	if err != nil {
		return err
	}
	err = client.sendMessage(message)
	return err
}

func (this *Engine) NewClient(w http.ResponseWriter, r *http.Request) (RemoteAddress, error) {
	if this.PoolClients.isFilled() {
		log.Println(
			fmt.Sprintf(
				"STREAMING [ERROR]: Connection websocket with client open failed. [error: %s]",
				ErrorPoolClientIsFilled.Error(),
			),
		)
		return RemoteAddress(""), ErrorPoolClientIsFilled
	}
	client, err := newClient(w, r, this.redirectMessageToHandler)
	if err != nil {
		log.Println(
			fmt.Sprintf(
				"STREAMING [ERROR]: Connection websocket with client [%s] open failed. [error: %s]",
				client.websocketRemoteAddress,
				err.Error(),
			),
		)
		return RemoteAddress(""), err
	}
	err = this.PoolClients.push(client)
	if err != nil {
		log.Println(
			fmt.Sprintf(
				"STREAMING [ERROR]: Connection websocket with client [%s] open failed. [error: %s]",
				client.websocketRemoteAddress,
				err.Error(),
			),
		)
		return RemoteAddress(""), err
	}
	this.rateLimiter.startNewClientStatistic(client.websocketRemoteAddress)
	log.Println(
		fmt.Sprintf(
			"STREAMING [OK]: Connection websocket with client [%s] open successfully",
			client.websocketRemoteAddress,
		),
	)
	go client.receiveMessage()
	return client.websocketRemoteAddress, nil
}

func (this *Engine) CloseConnectionClient(clientRemoteAddress RemoteAddress) {
	client, err := this.PoolClients.Get(clientRemoteAddress)
	if err != nil {
		log.Println(
			fmt.Sprintf(
				"STREAMING [ERROR]: Attempt to close a connection with a non-existent client [%s]",
				string(clientRemoteAddress),
			),
		)
		return
	}
	client.closeConnection()
	err = this.PoolClients.delete(client.websocketRemoteAddress)
	if err != nil {
		log.Println(
			fmt.Sprintf(
				"STREAMING [ERROR]: Attempt to delete a non-existent client [%s]",
				string(clientRemoteAddress),
			),
		)
		return
	}
	this.rateLimiter.deleteClientStatistic(clientRemoteAddress)
	log.Println(
		fmt.Sprintf(
			"STREAMING [OK]: Connection with client [%s] closed successfully",
			string(clientRemoteAddress),
		),
	)
}

func (this *Engine) redirectMessageToHandler(clientRemoteAddress RemoteAddress, message []byte) {
	request := new(Request)
	this.rateLimiter.updateClientStatistic(clientRemoteAddress)
	if err := proto.Unmarshal(message, request); err != nil {
		log.Println(
			fmt.Sprintf(
				"STREAMING [ERROR]: An error occurred while receiving a message from the client [%s] and transmitting it to the handler: [%s]",
				string(clientRemoteAddress),
				err.Error(),
			),
		)
		return
	}
	uri := request.GetUri()
	handler, err := this.poolHandlers.getHandler(URI(uri))
	if err != nil {
		log.Println(
			fmt.Sprintf(
				"STREAMING [ERROR]: An error occurred while receiving a message from the client [%s] and transmitting it to the handler: [%s]",
				string(clientRemoteAddress),
				err.Error(),
			),
		)
		return
	}
	log.Println(
		fmt.Sprintf(
			"STREAMING [OK]: Redirect on handler [%s] by request client [%s] successfully",
			uri,
			string(clientRemoteAddress),
		),
	)
	handler(&Context{
		ClientRemoteAddress: clientRemoteAddress,
		Message:             message,
		Error:               nil,
	})
}

func (this *Engine) waitRateLimiterEvents() {
	for clientRemoteAddress := range this.rateLimiter.channelConnectionCloseEvent {
		client, err := this.PoolClients.Get(clientRemoteAddress)
		if err != nil {
			continue
		}
		client.closeConnection()
		this.PoolClients.delete(clientRemoteAddress)
		this.rateLimiter.deleteClientStatistic(clientRemoteAddress)
		log.Println(
			fmt.Sprintf(
				"STREAMING [WARNING]: Сlient [%s] exceeded the allowed number of requests and was disconnected.",
				string(clientRemoteAddress),
			),
		)
	}
}
