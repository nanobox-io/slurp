package api

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jcelliott/lumber"

	"github.com/nanobox-io/slurp/backend"
	"github.com/nanobox-io/slurp/config"
)

func TestMain(m *testing.M) {
	// clean test dir
	os.RemoveAll("/tmp/slurpApi")

	// manually configure
	initialize()

	// start api
	go StartApi()
	<-time.After(2 * time.Second)
	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/slurpApi")

	os.Exit(rtn)
}

func TestPing(t *testing.T) {
	body, err := rest("GET", "/ping", "")
	if err != nil {
		t.Error(err)
	}
	if string(body) != "pong\n" {
		t.Errorf("%q doesn't match expected out", body)
	}
}

func TestAddStage(t *testing.T) {
	oldRead := cryptoRead
	cryptoRead = func(b []byte) (int, error) {
		copy(b, []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
		return 16, nil
	}
	defer func() { cryptoRead = oldRead }()

	body, err := rest("POST", "/stages", "{\"new-id\": \"newbuild\"}")
	if err != nil {
		t.Error(err)
	}

	if string(body) != "{\"secret\":\"01010101010101010101010101010101\"}\n" {
		t.Errorf("%q doesn't match expected out", body)
	}

	// badjson
	body, err = rest("POST", "/stages", "{\"new-id\"newbuild\"}")
	if err != nil {
		t.Error(err)
	}
	if string(body) != "{\"error\":\"Bad JSON Syntax Received in Body\"}\n" {
		t.Errorf("%q doesn't match expected out", body)
	}

	// missing payload
	body, err = rest("POST", "/stages", "{}")
	if err != nil {
		t.Error(err)
	}
	if string(body) != "{\"error\":\"Missing Payload Data\"}\n" {
		t.Errorf("%q doesn't match expected out", body)
	}
}

func TestCommitStage(t *testing.T) {
	body, err := rest("PUT", "/stages/newbuild", "")
	if err != nil {
		t.Error(err)
	}
	if string(body) != "{\"msg\":\"Success\"}\n" {
		t.Errorf("%q doesn't match expected out", body)
	}
}

func TestDeleteStage(t *testing.T) {
	body, err := rest("DELETE", "/stages/newbuild", "")
	if err != nil {
		t.Error(err)
	}
	if string(body) != "{\"msg\":\"Success\"}\n" {
		t.Errorf("%q doesn't match expected out", body)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVS
////////////////////////////////////////////////////////////////////////////////

// manually configure and start internals
func initialize() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	config.ApiToken = ""
	config.BuildDir = "/tmp/slurpApi/"
	config.LogLevel = "fatal"
	config.SshHostKey = "/tmp/slurp_rsa"
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))

	// initialize backend
	err := backend.Initialize()
	if err != nil {
		fmt.Printf("Backend init failed, skipping tests - %v\n", err)
		os.Exit(0)
	}
}

// hit api and return response body
func rest(method, route, data string) ([]byte, error) {
	body := bytes.NewBuffer([]byte(data))

	req, _ := http.NewRequest(method, fmt.Sprintf("%s%s", config.ApiAddress, route), body)
	req.Header.Add("X-AUTH-TOKEN", "")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Unable to %v %v - %v", method, route, err)
	}
	defer res.Body.Close()

	b, _ := ioutil.ReadAll(res.Body)

	return b, nil
}
