// Package "slurp" contains the core logic to fetch, pipe, and un/compress builds.
package slurp

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/nanobox-io/slurp/backend"
	"github.com/nanobox-io/slurp/config"
	"github.com/nanobox-io/slurp/ssh"
)

type build struct {
	ID     string
	Secret string
}

var (
	// copy of all non-committed builds
	builds []build

	// mutex ensures updates to builds are atomic
	mutex = sync.Mutex{}
)

// todo: slurp restart persistance? regenerate builds from config.BuildDir contents

// AddStage fetches the build "oldId" from the backend, uncompresses it to "newId",
// generates, and returns, a new user secret for rsyncing.
// Bash equivalent:
//  `curl localhost:7410/blobs/oldId | tar -C buildDir/newId -zxf -`
func AddStage(oldId, newId, secret string) error {
	// prepare location for extraction
	err := os.MkdirAll(config.BuildDir+"/"+newId, 0755)
	if err != nil {
		return fmt.Errorf("Failed to create build dir - %v", err)
	}

	// backend.ReadBlob(oldId) | tar -C buildDir/newId -zxf -
	if oldId != "" {
		// stream last build from backend
		res, err := backend.ReadBlob(oldId)
		if err != nil {
			return fmt.Errorf("Failed to get old build - %v", err)
		}

		config.Log.Trace("Fetched build")

		// prepare to extract to new build dir
		cmd := exec.Command("tar", "--atime-preserve", "-C", newId, "-zxf", "-")
		// cmd.Dir = "/tmp"
		cmd.Dir = config.BuildDir

		// pipe build to extract command
		cmd.Stdin = res

		config.Log.Trace("Running extract command '%v'", cmd.Args)
		// err := cmd.Run()
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Failed to extract build to dir '%s' - %v", out, err)
		}

		config.Log.Trace("Extracted build")
		res.Close()
	}

	err = ssh.AddUser(secret)
	if err != nil {
		return fmt.Errorf("Failed to add user - %v", err)
	}

	mutex.Lock()
	builds = append(builds, build{
		ID:     newId,
		Secret: secret,
	})
	mutex.Unlock()

	return nil
}

// CommitStage compresses the new build, uploads it to the backend and removes
// the user secret from the ssh server.
// Bash equivalent:
//  `tar -C buildDir/buildId -czf - . | curl localhost:7410/blobs/newId -T -`
func CommitStage(buildId string) error {
	// remove user first
	secret, err := getUser(buildId)
	if err == nil {
		err = ssh.DelUser(secret)
		if err != nil {
			return fmt.Errorf("Failed to remove user - %v", err)
		}
	}

	// don't buffer (free the rams)
	blobReader, blobWriter := io.Pipe()

	config.Log.Trace("Preparing to compress '%v'", config.BuildDir+"/"+buildId)

	// check for existing build
	_, err = os.Stat(config.BuildDir + "/" + buildId)
	if err != nil {
		return fmt.Errorf("Build dir doesn't exist - %v", err)
	}

	// tar -C buildDir/buildId -czf - . | backend.WriteBlob(buildId)
	// prepare to compress build dir
	cmd := exec.Command("tar", "-C", config.BuildDir+"/"+buildId, "-czf", "-", ".")
	// cmd.Dir = "/tmp"
	cmd.Dir = config.BuildDir

	// keep the modified time unchanged when compressing (keep md5 the same)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GZIP=-n")

	// pipe compressed build to write command
	cmd.Stdout = blobWriter

	config.Log.Trace("Running compress command '%v'", cmd.Args)

	// prep writing build to backend
	echan := make(chan error, 1)

	// start stream to backend
	go func() {
		echan <- backend.WriteBlob(buildId, blobReader)
	}()

	// compress the build
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to compress build - %v", err)
		// the error `io: read/write on closed pipe` here is likely due to
		// wrong backend protocol (http/https) resolve with different scheme
		// for hoarder 'hoarder[s]://'
	}

	config.Log.Trace("Compressed build")

	// if the command finished, blobWriter is done
	blobWriter.Close()

	// wait for WriteBlob to finish
	err = <-echan
	if err != nil {
		return fmt.Errorf("Failed to write build - %v", err)
	}

	config.Log.Trace("Uploaded build")

	return nil
}

// DeleteStage removes files for a specific build.
func DeleteStage(buildId string) error {
	// remove user first
	secret, err := getUser(buildId)
	if err != nil {
		err = ssh.DelUser(secret)
		if err != nil {
			return fmt.Errorf("Failed to remove user - %v", err)
		}
	}

	config.Log.Trace("Removing '%v'", config.BuildDir+"/"+buildId)

	// remove build files
	err = os.RemoveAll(config.BuildDir + "/" + buildId)
	if err != nil {
		return fmt.Errorf("Failed to remove build dir - %v", err)
	}

	// remove cached build
	mutex.Lock()
	for i := range builds {
		if builds[i].ID == buildId {
			builds = append(builds[:i], builds[i+1:]...)
			break
		}
	}
	mutex.Unlock()

	return nil
}

// getUser gets the user secret corresponding to an uncommitted build.
func getUser(buildId string) (string, error) {
	for _, build := range builds {
		if build.ID == buildId {
			return build.Secret, nil
		}
	}
	return "", fmt.Errorf("No Build Found")
}
