package ssh_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/jcelliott/lumber"

	"github.com/nanobox-io/slurp/config"
	"github.com/nanobox-io/slurp/ssh"
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
	<-time.After(3 * time.Second)

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

	exec.Command("bash", "-c", "ssh-keygen -t rsa -b 2048 -C 'slurp@test' -N '' -f /tmp/slurp-usr").Run()
	cmd := exec.Command("rsync", "-v", "--delete", "-aR", ".", "-e", "ssh -i /tmp/slurp-usr -p 1567 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null", "sshTest@127.0.0.1:sshTest")
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
