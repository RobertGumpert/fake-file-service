package application

import (
	"fmt"
	"log"
	"protoservice/src/fileservice"
	"protoservice/src/streaming"
)

type FileServiceManager struct {
	websocketEngine *streaming.Engine
	fileService     *fileservice.Service
}

func NewFileServiceManager(websocketEngine *streaming.Engine, rootPath, storagePath string) *FileServiceManager {
	this := new(FileServiceManager)
	this.websocketEngine = websocketEngine
	this.fileService = fileservice.NewService(
		websocketEngine,
		rootPath,
		storagePath,
	)
	this.websocketEngine.Handle("/send/file", this.fileService.HandleReceivingFileFrames)
	this.websocketEngine.Handle("/session/open", this.fileService.HandleOpenSession)
	go this.waitCloseSession()
	go this.waitFileFrame()
	go this.waitOpenSession()
	return this
}

func (this *FileServiceManager) waitCloseSession() {
	for event := range this.fileService.SessionClosingEventChannel {
		if event.OK {
			log.Println(
				fmt.Sprintf(
					"FILESERVICE-MANAGER [OK]: session close with client [%s].",
					event.Context.ClientRemoteAddress,
				),
			)
		} else {
			log.Println(
				fmt.Sprintf(
					"FILESERVICE-MANAGER [ERROR]: session close with client [%s] have error. [error: %s]",
					event.Context.ClientRemoteAddress,
					event.Error.Error(),
				),
			)
		}
		this.websocketEngine.CloseConnectionClient(event.Context.ClientRemoteAddress)
	}
}

func (this *FileServiceManager) waitOpenSession() {
	for event := range this.fileService.SessionOpeningEventChannel {
		if event.OK {
			log.Println(
				fmt.Sprintf(
					"FILESERVICE-MANAGER [OK]: session open with client [%s].",
					event.Context.ClientRemoteAddress,
				),
			)
		} else {
			log.Println(
				fmt.Sprintf(
					"FILESERVICE-MANAGER [ERROR]: session open with client [%s] have error. [error: %s]",
					event.Context.ClientRemoteAddress,
					event.Error.Error(),
				),
			)
			this.websocketEngine.CloseConnectionClient(event.Context.ClientRemoteAddress)
		}
	}
}

func (this *FileServiceManager) waitFileFrame() {
	for event := range this.fileService.FileFrameReceiveEventChannel {
		if event.OK {
			log.Println(
				fmt.Sprintf(
					"FILESERVICE-MANAGER [OK]: file frame receive by client [%s].",
					event.Context.ClientRemoteAddress,
				),
			)
		} else {
			log.Println(
				fmt.Sprintf(
					"FILESERVICE-MANAGER [ERROR]: file frame receive by client [%s]. [error: %s]",
					event.Context.ClientRemoteAddress,
					event.Error.Error(),
				),
			)
			this.websocketEngine.CloseConnectionClient(event.Context.ClientRemoteAddress)
		}
	}
}