package test

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"protoservice/src/fileservice"
	"protoservice/src/streaming"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type FakeClient struct {
	TestServer *httptest.Server
	backend    *url.URL
	connection *websocket.Conn
	done       chan struct{}
}

func NewFakeClient(backend *url.URL) *FakeClient {
	this := new(FakeClient)
	this.TestServer = httptest.NewServer(this)
	this.backend = backend
	this.done = make(chan struct{})
	return this
}

func (this *FakeClient) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := this.ConnectWithServerByWS("/ws")
	if err != nil {
		log.Println(err)
	}
}

func (this *FakeClient) ConnectWithServerByWS(endpoint string) error {
	this.backend.Scheme = "ws"
	this.backend.Path = endpoint
	connection, _, err := websocket.DefaultDialer.Dial(this.backend.String(), nil)
	if err != nil {
		return err
	}
	this.connection = connection
	go this.handleWS()
	return nil
}

func (this *FakeClient) handleWS() {

	// /Users/vlad/Job/sber/prototest/protoservice/storage/test

	requestHandshake := &fileservice.HandshakeRequest{
		RemoteAddress: this.connection.LocalAddr().String(),
	}
	reqbts, err := proto.Marshal(requestHandshake)
	if err != nil {
		log.Println(err)
		return
	}
	request := &streaming.Request{
		Uri:   "/session/open",
		Frame: reqbts,
	}
	readyMessage, err := proto.Marshal(request)
	if err != nil {
		log.Println(err)
		return
	}
	err = this.connection.WriteMessage(
		websocket.BinaryMessage,
		readyMessage,
	)
	if err != nil {
		log.Println(err)
		return
	}
	//
	_, respbts, err := this.connection.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}
	if err != nil {
		log.Println(err)
		return
	}
	responseHandshake := new(fileservice.HandshakeResponce)
	err = proto.Unmarshal(respbts, responseHandshake)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Start session from client ", this.connection.LocalAddr().String(), " with uuid ", responseHandshake.GetSessionUuid())

	file, err := ioutil.ReadFile("/Users/vlad/Job/sber/prototest/protoservice/storage/test")
	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		err := this.connection.Close()
		if err != nil {
			log.Println("Fnish :", err)
		}
	}()
	go func() {
		defer close(this.done)
		for {
			_, _, err := this.connection.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
		}
	}()

	countSlices := int(len(file) / 5)
	from, to := int(0), countSlices
	timer := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-this.done:
			return
		case <-timer.C:
			var (
				bts       []byte
				lastframe = false
			)
			if to < len(file) {
				bts = file[from:to]
			} else {
				bts = file[from : len(file)-1]
				lastframe = true
			}
			frame := &fileservice.FileStreamingRequest{
				LastFrame:      lastframe,
				StreamingFrame: bts,
				SessionUuid:    responseHandshake.GetSessionUuid(),
			}
			if message, err := proto.Marshal(frame); err != nil {
				log.Println(err)
				return
			} else {
				request := &streaming.Request{
					Uri:   "/send/file",
					Frame: message,
				}
				readyMessage, err := proto.Marshal(request)
				if err != nil {
					log.Println(err)
					return
				}
				err = this.connection.WriteMessage(
					websocket.BinaryMessage,
					readyMessage,
				)
				if err != nil {
					log.Println(err)
					return
				}
			}
			from = from + to
			to = to + 1
			if lastframe {
				return
			}
		}
	}
}
