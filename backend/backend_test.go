package backend_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/jcelliott/lumber"

	"github.com/nanobox-io/slurp/backend"
	"github.com/nanobox-io/slurp/config"
)

func TestMain(m *testing.M) {
	// manually configure
	initialize()

	rtn := m.Run()

	os.Exit(rtn)
}

func TestWriteBlob(t *testing.T) {
	body := bytes.Buffer{}
	body.Write([]byte("big-build"))

	err := backend.WriteBlob("test", &body)
	if err != nil {
		t.Error(err)
	}
}

func TestReadBlob(t *testing.T) {
	body, err := backend.ReadBlob("test")
	if err != nil {
		t.Error(err)
	}
	buff := make([]byte, 9)
	body.Read(buff)

	if string(buff) != "big-build" {
		t.Errorf("%q doesn't match expected out", body)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVS
////////////////////////////////////////////////////////////////////////////////

// manually configure and start internals
func initialize() {
	config.LogLevel = "fatal"
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))

	// initialize backend
	err := backend.Initialize()
	if err != nil {
		fmt.Printf("Backend init failed, skipping tests - %v\n", err)
		os.Exit(0)
	}
}
