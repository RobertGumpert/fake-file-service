package application

import (
	"log"
	"net/http"
	"protoservice/src/streaming"

	"github.com/gin-gonic/gin"
)

type HttpEngine struct {
	HttpEngine         *gin.Engine
	RunHttpEngine      func()
	websocketEngine    *streaming.Engine
	fileServiceManager *FileServiceManager
}

func NewHttpEngine(websocketEngine *streaming.Engine, port string, rootPath, storagePath string) *HttpEngine {
	engine := gin.New()
	this := new(HttpEngine)
	this.HttpEngine = engine
	this.websocketEngine = websocketEngine
	this.fileServiceManager = NewFileServiceManager(
		websocketEngine,
		rootPath,
		storagePath,
	)
	//
	engine.GET("/ws", this.openWebsocket)
	this.RunHttpEngine = func() {
		err := engine.Run(port)
		if err != nil {
			log.Fatal(err)
		}
	}
	return this
}

func (this *HttpEngine) openWebsocket(context *gin.Context) {
	_, err := this.websocketEngine.NewClient(
		http.ResponseWriter(context.Writer),
		context.Request,
	)
	if err != nil {
		context.AbortWithStatus(http.StatusLocked)
		return
	}
}
