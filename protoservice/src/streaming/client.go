package streaming

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type client struct {
	httpRemoteAddress      RemoteAddress
	websocketRemoteAddress RemoteAddress
	connection             *websocket.Conn
	mx                     *sync.Mutex
	callback               func(clientRemoteAddress RemoteAddress, message []byte)
	connectionIsClosed     bool
}

func newClient(w http.ResponseWriter, r *http.Request, callback func(clientRemoteAddress RemoteAddress, message []byte)) (*client, error) {
	this := new(client)
	upgrader := &websocket.Upgrader{}
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(
			fmt.Sprintf(
				"STREAMING [ERROR]: Protocol switching, for client [%s], happened with an error: [%s]",
				this.websocketRemoteAddress,
				err.Error(),
			),
		)
		return nil, err
	}
	this.connection = connection
	this.websocketRemoteAddress = RemoteAddress(connection.RemoteAddr().String())
	this.callback = callback
	log.Println(
		fmt.Sprintf(
			"STREAMING [OK]: Protocol switch, for client [%s], succeeded",
			this.websocketRemoteAddress,
		),
	)
	return this, err
}

func (this *client) sendMessage(message []byte) error {
	if !this.connectionIsClosed {
		return nil
	}
	this.mx.Lock()
	defer this.mx.Unlock()
	err := this.connection.WriteMessage(
		websocket.BinaryMessage,
		message,
	)
	if err != nil {
		log.Println(
			fmt.Sprintf(
				"STREAMING [ERROR]: Sending a message to the client [%s] failed: [%s]",
				this.websocketRemoteAddress,
				err.Error(),
			),
		)
		return err
	}
	log.Println(
		fmt.Sprintf(
			"STREAMING [OK]: Sending a message to the client [%s] succeeded",
			this.websocketRemoteAddress,
		),
	)
	return nil
}

func (this *client) receiveMessage() {
	defer this.closeConnection()
	for {
		if this.connectionIsClosed {
			return
		}
		_, message, err := this.connection.ReadMessage()
		if err != nil {
			log.Println(
				fmt.Sprintf(
					"STREAMING [ERROR]: An error message was received from the client [%s]: [%s].",
					this.websocketRemoteAddress,
					err.Error(),
				),
			)
			return
		}
		go this.callback(this.websocketRemoteAddress, message)
	}
}

func (this *client) closeConnection() {
	if this.connectionIsClosed {
		return
	}
	this.connectionIsClosed = true
	if err := this.connection.Close(); err != nil {
		log.Println(
			fmt.Sprintf(
				"STREAMING [ERROR]: Connection with client [%s] closed with error [%s]",
				this.websocketRemoteAddress,
				err,
			),
		)
	} else {
		log.Println(
			fmt.Sprintf(
				"STREAMING [OK]: Connection with client [%s] closed",
				this.websocketRemoteAddress,
			),
		)
	}
}
