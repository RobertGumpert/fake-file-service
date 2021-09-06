package test

import (
	"net/http/httptest"
	"protoservice/src/application"
	"protoservice/src/streaming"
)

type FakeServer struct {
	HttpEngine      *application.HttpEngine
	WebsocketEngine *streaming.Engine
	TestServer      *httptest.Server
}

// /Users/vlad/Job/sber/prototest/protoservice/storage/fileservice
func NewFakeServer() *FakeServer {
	this := new(FakeServer)
	this.WebsocketEngine = streaming.NewEngine(5, 100)
	this.HttpEngine = application.NewHttpEngine(
		this.WebsocketEngine,
		"",
		"/Users/vlad/Job/sber/prototest/protoservice",
		"storage/fileservice",
	)
	this.TestServer = httptest.NewServer(this.HttpEngine.HttpEngine)
	return this
}
