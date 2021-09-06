package fileservice

import (
	"bufio"
	"bytes"
	"os"
	"protoservice/src/streaming"
	"github.com/google/uuid"
)

type session struct {
	storagePath         string
	remoteClientAddress streaming.RemoteAddress
	sessionUUID         uuid.UUID
	fileBuffer          *bytes.Buffer
}

func newSession(remoteClientAddress streaming.RemoteAddress, storagePath string) *session {
	this := new(session)
	this.remoteClientAddress = remoteClientAddress
	this.sessionUUID = uuid.New()
	this.fileBuffer = bytes.NewBuffer(nil)
	return this
}

func (this *session) appendFileBytes(fileBytes []byte) error {
	_, err := this.fileBuffer.Write(fileBytes)
	if err != nil {
		return err
	}
	return nil
}

func (this *session) writeToDisk() error {
	file, err := os.Create(this.storagePath)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	_, err = writer.Write(this.fileBuffer.Bytes())
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}
