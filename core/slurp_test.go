package slurp_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/jcelliott/lumber"

	"github.com/nanobox-io/slurp/backend"
	"github.com/nanobox-io/slurp/config"
	"github.com/nanobox-io/slurp/core"
)

func TestMain(m *testing.M) {
	// clean test dir
	os.RemoveAll("/tmp/slurpCore")

	// manually configure
	initialize()

	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/slurpCore")

	os.Exit(rtn)
}

func TestAddStage(t *testing.T) {
	err := slurp.AddStage("", "core-new", "sekret")
	if err != nil {
		t.Error(err)
	}

	// use build from api_test
	err = slurp.AddStage("newbuild", "core-new", "sekret2")
	if err != nil {
		t.Error(err)
	}
}

func TestCommitStage(t *testing.T) {
	err := slurp.CommitStage("core-new")
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteStage(t *testing.T) {
	err := slurp.DeleteStage("core-new")
	if err != nil {
		t.Error(err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVS
////////////////////////////////////////////////////////////////////////////////

// manually configure and start internals
func initialize() {
	config.BuildDir = "/tmp/slurpCore/"
	config.LogLevel = "fatal"
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))

	// initialize backend
	err := backend.Initialize()
	if err != nil {
		fmt.Printf("Backend init failed, skipping tests - %v\n", err)
		os.Exit(0)
	}
}
