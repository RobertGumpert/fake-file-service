package fileservice

import (
	"fmt"
	"log"
	"protoservice/src/streaming"
	"strings"

	"google.golang.org/protobuf/proto"
)

type Event struct {
	Context *streaming.Context
	OK      bool
	Error   error
}

type Service struct {
	websocketEngine              *streaming.Engine
	poolSession                  *poolSessionManager
	RootPath                     string
	StoragePath                  string
	SessionOpeningEventChannel   chan Event
	SessionClosingEventChannel   chan Event
	FileFrameReceiveEventChannel chan Event
}

func NewService(websocketEngine *streaming.Engine, rootPath, storagePath string) *Service {
	this := new(Service)
	this.websocketEngine = websocketEngine
	this.poolSession = newPoolSessionManager()
	this.RootPath = rootPath
	this.StoragePath = storagePath
	this.SessionClosingEventChannel = make(chan Event)
	this.SessionOpeningEventChannel = make(chan Event)
	this.FileFrameReceiveEventChannel = make(chan Event)
	return this
}

func (this *Service) HandleReceivingFileFrames(context *streaming.Context) {
	fileFrame := new(FileStreamingRequest)
	err := proto.Unmarshal(context.Message, fileFrame)
	if err != nil {
		log.Println(
			fmt.Sprintf(
				"FILESERVICE [ERROR]: Receiving file frame from client [%s] failed. [error: %s]",
				string(context.ClientRemoteAddress),
				err.Error(),
			),
		)
		this.FileFrameReceiveEventChannel <- Event{
			Context: context,
			OK:      false,
			Error:   err,
		}
		return
	}
	session, err := this.poolSession.get(uuidCode(fileFrame.GetSessionUuid()))
	if err != nil {
		log.Println(
			fmt.Sprintf(
				"FILESERVICE [ERROR]: Receiving file frame from client [%s] failed. [error: %s]",
				string(context.ClientRemoteAddress),
				err.Error(),
			),
		)
		this.FileFrameReceiveEventChannel <- Event{
			Context: context,
			OK:      false,
			Error:   err,
		}
		return
	}
	err = session.appendFileBytes(fileFrame.GetStreamingFrame())
	if err != nil {
		log.Println(
			fmt.Sprintf(
				"FILESERVICE [ERROR]: Receiving file frame from client [%s] failed. [error: %s]",
				string(context.ClientRemoteAddress),
				err.Error(),
			),
		)
		this.FileFrameReceiveEventChannel <- Event{
			Context: context,
			OK:      false,
			Error:   err,
		}
		return
	}
	if fileFrame.GetLastFrame() {
		err = session.writeToDisk()
		if err != nil {
			log.Println(
				fmt.Sprintf(
					"FILESERVICE [ERROR]: Closing of the session. Receiving file frames from client [%s], with session [%s], completed failed. [%s]",
					string(context.ClientRemoteAddress),
					fileFrame.GetSessionUuid(),
					err.Error(),
				),
			)
		}
		err = this.poolSession.delete(uuidCode(fileFrame.GetSessionUuid()))
		if err != nil {
			log.Println(
				fmt.Sprintf(
					"FILESERVICE [ERROR]: Closing of the session. Receiving file frames from client [%s], with session [%s], completed failed. [%s]",
					string(context.ClientRemoteAddress),
					fileFrame.GetSessionUuid(),
					err.Error(),
				),
			)
			this.SessionClosingEventChannel <- Event{
				Context: context,
				OK:      false,
				Error:   err,
			}
			return
		}
		log.Println(
			fmt.Sprintf(
				"FILESERVICE [OK]: Closing of the session. Receiving file frames from client [%s], with session [%s], completed successfully",
				string(context.ClientRemoteAddress),
				fileFrame.GetSessionUuid(),
			),
		)
		this.SessionClosingEventChannel <- Event{
			Context: context,
			OK:      true,
			Error:   nil,
		}
		return
	}
	log.Println(
		fmt.Sprintf(
			"FILESERVICE [OK]: Receiving file frames from client [%s], with session [%s], completed successfully",
			string(context.ClientRemoteAddress),
			fileFrame.GetSessionUuid(),
		),
	)
	this.FileFrameReceiveEventChannel <- Event{
		Context: context,
		OK:      true,
		Error:   nil,
	}
}

func (this *Service) HandleOpenSession(context *streaming.Context) {
	sessionStart := new(HandshakeRequest)
	err := proto.Unmarshal(context.Message, sessionStart)
	if err != nil {
		log.Println(
			fmt.Sprintf(
				"FILESERVICE [ERROR]: An error occurred while opening a session with the client [%s]. [error: %s]",
				string(context.ClientRemoteAddress),
				err.Error(),
			),
		)
		this.SessionOpeningEventChannel <- Event{
			Context: context,
			OK:      false,
			Error:   err,
		}
		return
	}
	if sessionStart.GetRemoteAddress() == "" {
		log.Println(
			fmt.Sprintf(
				"FILESERVICE [ERROR]: An error occurred while opening a session with the client [%s]. [error: non exist address in context]",
				"???",
			),
		)
		this.SessionOpeningEventChannel <- Event{
			Context: context,
			OK:      false,
			Error:   err,
		}
		return
	}
	websocketClient, err := this.websocketEngine.PoolClients.Get(
		streaming.RemoteAddress(
			sessionStart.GetRemoteAddress(),
		),
	)
	if err != nil {
		log.Println(
			fmt.Sprintf(
				"FILESERVICE [ERROR]: An error occurred while opening a session with the client [%s]. [error: %s]",
				string(context.ClientRemoteAddress),
				err.Error(),
			),
		)
		this.SessionOpeningEventChannel <- Event{
			Context: context,
			OK:      false,
			Error:   err,
		}
		return
	}
	if websocketClient == nil {
		log.Println(
			fmt.Sprintf(
				"FILESERVICE [ERROR]: An error occurred while opening a session with the client [%s]. [error: none exist client in pool]",
				string(context.ClientRemoteAddress),
			),
		)
		this.SessionOpeningEventChannel <- Event{
			Context: context,
			OK:      false,
			Error:   err,
		}
		return
	}
	session := newSession(
		streaming.RemoteAddress(sessionStart.RemoteAddress),
		strings.Join([]string{
			this.RootPath,
			this.StoragePath,
		}, "/"),
	)
	err = this.poolSession.push(session)
	if err != nil {
		log.Println(
			fmt.Sprintf(
				"FILESERVICE [ERROR]: An error occurred while opening a session with the client [%s]. [error: %s]",
				string(context.ClientRemoteAddress),
				err.Error(),
			),
		)
		this.SessionOpeningEventChannel <- Event{
			Context: context,
			OK:      false,
			Error:   err,
		}
		return
	}
	log.Println(
		fmt.Sprintf(
			"FILESERVICE [OK]: Opening a session with the client [%s] completed successfully. [uuid: %s]",
			string(context.ClientRemoteAddress),
			session.sessionUUID.String(),
		),
	)
	this.SessionOpeningEventChannel <- Event{
		Context: context,
		OK:      true,
		Error:   nil,
	}
}
