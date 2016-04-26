package ssh_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/jcelliott/lumber"

	"github.com/nanopack/slurp/config"
	"github.com/nanopack/slurp/ssh"
)

func TestMain(m *testing.M) {
	// clean test dir
	os.RemoveAll("/tmp/slurpSsh")
	os.RemoveAll("/tmp/slurp_rsa")
	os.RemoveAll("/tmp/sshTest")

	// manually configure
	initialize()

	// start ssh server
	go ssh.Start()
	<-time.After(2 * time.Second)

	rtn := m.Run()

	// clean test dir
	os.RemoveAll("/tmp/slurpSsh")
	os.RemoveAll("/tmp/slurp_rsa")
	os.RemoveAll("/tmp/sshTest")

	os.Exit(rtn)
}

func TestAddUser(t *testing.T) {
	err := ssh.AddUser("sshTest")
	if err != nil {
		t.Error(err)
	}
}

func TestCommitStage(t *testing.T) {
	err := os.MkdirAll("/tmp/sshTest", 0755)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	file, err := os.Create("/tmp/sshTest/file")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, err = file.WriteString("SomeThing")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	cmd := exec.Command("rsync", "-v", "--delete", "-aR", ".", "-e", "ssh -p 1567", "sshTest@127.0.0.1:sshTest")
	cmd.Dir = "/tmp/sshTest/"
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("OUT: ", string(out))
		t.Error(err)
		t.FailNow()
	}
}

func TestDelUser(t *testing.T) {
	err := ssh.DelUser("sshTest")
	if err != nil {
		t.Error(err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVS
////////////////////////////////////////////////////////////////////////////////

// manually configure and start internals
func initialize() {
	config.BuildDir = "/tmp/slurpSsh/"
	config.LogLevel = "fatal"
	config.SshHostKey = "/tmp/slurp_rsa"
	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))

	// prepare build dir
	err := os.MkdirAll(config.BuildDir, 0755)
	if err != nil {
		fmt.Printf("Failed to create build dir - %v\n", err)
		os.Exit(1)
	}
}
