package test

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)


func TestFileStreamingServerFlow(t *testing.T) {
	fakeServer := NewFakeServer()
	defer fakeServer.TestServer.Close()
	u, err := url.Parse(fakeServer.TestServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	fakeClient := NewFakeClient(u)
	_, err = http.Get(fakeClient.TestServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Minute)
}
