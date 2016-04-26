package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jcelliott/lumber"

	"github.com/nanobox-io/slurp/backend"
	"github.com/nanobox-io/slurp/config"
)

func TestMain(m *testing.M) {
	// clean test dir
	os.RemoveAll("/tmp/slurpMain")

	// manually configure
	initialize()

	args := strings.Split("-b /tmp/slurpMain/ -l fatal -k /tmp/slurp_rsa -s 127.0.0.1:1568 -a 127.0.0.1:1564", " ")
	slurp.SetArgs(args)

	// start api
	go slurp.Execute()
	<-time.After(time.Second)
	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/slurpMain")

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

////////////////////////////////////////////////////////////////////////////////
// PRIVS
////////////////////////////////////////////////////////////////////////////////

// manually configure and start internals
func initialize() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))
	config.ApiToken = ""

	// check for hoarder
	err := backend.Initialize()
	if err != nil {
		fmt.Printf("Backend init failed, skipping tests - %v\n", err)
		os.Exit(0)
	}
}

// hit api and return response body
func rest(method, route, data string) ([]byte, error) {
	body := bytes.NewBuffer([]byte(data))

	req, _ := http.NewRequest(method, fmt.Sprintf("https://%s%s", config.ApiAddress, route), body)
	req.Header.Add("X-AUTH-TOKEN", "")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Unable to %v %v - %v", method, route, err)
	}
	defer res.Body.Close()

	b, _ := ioutil.ReadAll(res.Body)

	return b, nil
}
