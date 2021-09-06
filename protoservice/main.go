package main

import (
	"protoservice/src/streaming"
)

// protoc --go_out=. src/proto/*.proto
// export PATH=$PATH:/Users/vlad/go/bin

func main() {
	engine := streaming.NewEngine(5, 10)
	engine.Handle(streaming.URI("/session/open"), func(context *streaming.Context) {
		
	})
	engine.Handle(streaming.URI("/file/send"), func(context *streaming.Context) {
		
	})
}