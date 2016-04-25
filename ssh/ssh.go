package ssh

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	mrand "math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/nanopack/slurp/config"
)

// characters for username
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// random seed
var seed = time.Now().UnixNano()

// Check for host key, generate and write to a file if none exist
func init() {
	// seed random user generation
	mrand.Seed(seed)

	// check if key exists
	_, err := os.Stat(config.SshHostKey)
	if err == nil {
		return
	}

	// generate a new host key
	_, hostPrv, err := genKeyPair()
	if err != nil {
		config.Log.Fatal("Failed to generate host key - %v", err)
	}

	// ensure keyfile directory exists
	err = os.MkdirAll(filepath.Dir(config.SshHostKey), 0755)
	if err != nil {
		config.Log.Fatal("Failed to create host key directory - %v", err)
	}

	// store new host key
	key := []byte(hostPrv)
	err = ioutil.WriteFile(config.SshHostKey, key, 0600)
	if err != nil {
		config.Log.Fatal("Failed to write host key to file - %v", err)
	}
}

// generate a random user for rsync clients
func genUser() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = letterBytes[mrand.Intn(len(letterBytes))]
	}
	return (string(b))
}

// genKeyPair make a pair of public and private keys for SSH access.
// Public key is encoded in the format for inclusion in an OpenSSH authorized_keys file.
// Private Key generated is PEM encoded
func genKeyPair() (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	var private bytes.Buffer
	if err := pem.Encode(&private, privateKeyPEM); err != nil {
		return "", "", err
	}

	// create ssh.PublicKey
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	return string(ssh.MarshalAuthorizedKey(pub)), private.String(), nil
}

// gets the host key from file
func getKey() ([]byte, error) {
	key, err := ioutil.ReadFile(config.SshHostKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to read host key from file - %v", err)
	}
	return key, nil
}

// StartSSH starts the ssh server that will handle the rsync
func Start() error {
	// get host key
	hostPrv, err := getKey()
	if err != nil {
		return fmt.Errorf("Failed to get key - %v", err)
	}

	// parse key
	pvtKeySigner, err := ssh.ParsePrivateKey(hostPrv)
	if err != nil {
		return fmt.Errorf("Failed to parse private key - %v", err)
	}

	// initialize ssh config
	sshConfig := &ssh.ServerConfig{
		PublicKeyCallback: userAuth,
		ServerVersion:     "SSH-2.0-NANOPACK-SLURP",
		AuthLogCallback:   logAuth,
	}

	// add host key
	sshConfig.AddHostKey(pvtKeySigner)

	// start tcp server
	serverSocket, err := net.Listen("tcp", config.SshAddr)
	if err != nil {
		return fmt.Errorf("Failed to listen for rsync - %v", err)
	}

	config.Log.Info("Listening for rsync on '%v'...", config.SshAddr)

	// accept connections
	go func() {
		for {
			conn, err := serverSocket.Accept()
			if err != nil {
				config.Log.Error("Failed to accept connection - %v", err)
				continue
			}
			config.Log.Trace("Got connection")
			go handleConnection(conn, sshConfig)
		}
	}()
	return nil
}

// logAuth logs when a user is attempting to authenticate
func logAuth(conn ssh.ConnMetadata, method string, err error) {
	config.Log.Debug("User '%v' connecting from '%v' with '%v' method '%v'", conn.User(), conn.RemoteAddr().String(), string(conn.ClientVersion()), method)
}

// authenticate connection based on username
func userAuth(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	config.Log.Trace("Attempting to auth user: '%v'", conn.User())
	// assign a new var here to prevent issues using a user as its deleted
	for _, permittedUser := range authUsers {
		if conn.User() == permittedUser {
			config.Log.Debug("User: '%v' authorized", conn.User())
			return nil, nil
		}
	}
	config.Log.Error("User: '%v' not found!", conn.User())
	return nil, fmt.Errorf("User not found!")
}

// handle tcp connection
func handleConnection(conn net.Conn, sshConfig *ssh.ServerConfig) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, sshConfig)
	if err != nil {
		config.Log.Error("Failed to handshake - %v", err)
		return
	}
	config.Log.Debug("Handshake successful")

	defer sshConn.Close()

	// service incoming request channel
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			config.Log.Debug("Unknown channel type - %v", newChannel.ChannelType())
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		handleChannel(newChannel)
	}
}

// handle ssh connections
func handleChannel(newChannel ssh.NewChannel) {
	channel, requests, err := newChannel.Accept()
	if err != nil {
		config.Log.Error("Failed to accept channel request - %v", err)
		return
	}

	go func(in <-chan *ssh.Request) {
		for req := range in {
			config.Log.Trace("Req recieved - %v", req.Type)
			ok := false
			switch req.Type {
			case "exec":
				ok = true
				if len(req.Payload) < 4 {
					config.Log.Debug("Payload Too Small")
					ok = false
					continue // todo: or break?
				}

				waitedRun(channel, string(req.Payload[4:]))
			case "env":
				ok = true
			}
			req.Reply(ok, nil)
		}
	}(requests)
}

// run command (rsync server)
func waitedRun(channel ssh.Channel, command string) {
	defer channel.Close()

	config.Log.Trace("Command: '%v'", command)
	cmd := exec.Command("/bin/bash", "-c", command)
	cmd.Dir = config.BuildDir // todo: isolate maybe?

	// connect stdin/out to the ssh pipe
	cmd.Stdin = channel
	cmd.Stdout = channel
	cmd.Stderr = channel.Stderr()

	// todo: will change once exec.Command has hardcoded values
	config.Log.Trace("running command %+v", cmd.Args)

	// start running the command
	err := cmd.Start()
	if err != nil || cmd.Process == nil {
		config.Log.Fatal("Failed to run command - %v", err)
		return
	}

	config.Log.Trace("PID: %v\n", cmd.Process.Pid)

	// using cmd.Wait(), the PID gets killed, but it gets stuck on a c.goroutine (the stdin io.Copy() one)
	// and doesn't return, hence the implementation.
	state, err := cmd.Process.Wait()
	cmd.ProcessState = state
	if err != nil {
		config.Log.Fatal("Failed to wait - %v", err)
		// return // todo: ? or let go?
	}

	// release resources associated to process
	cmd.Process.Release()

	// check exit status
	exitStatusBuffer := []byte{0, 0, 0, 0}
	if strings.Contains(state.String(), "exit status") {
		status := strings.Split(state.String(), " ")[2]
		if status != "0" {
			// exit 1
			exitStatusBuffer = []byte{0, 0, 0, 1}
		}
	} else {
		// exit 2
		exitStatusBuffer = []byte{0, 0, 0, 2}
	}

	// return exit status to client
	channel.SendRequest("exit-status", true, exitStatusBuffer)
	config.Log.Trace("Command's exit-status returned")
}
