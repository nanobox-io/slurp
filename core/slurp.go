// Package "slurp" contains the core logic to fetch, pipe, and un/compress builds.
package slurp

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/nanobox-io/slurp/backend"
	"github.com/nanobox-io/slurp/config"
	"github.com/nanobox-io/slurp/ssh"
)

var (
	// copy of all non-committed builds
	builds []string

	// mutex ensures updates to builds are atomic
	mutex = sync.Mutex{}
)

// todo: slurp restart persistance? regenerate builds from config.BuildDir contents

// AddStage fetches the build "oldId" from the backend, uncompresses it to "newId",
// generates, and returns, a new user secret for rsyncing.
// Bash equivalent:
//  `curl localhost:7410/blobs/oldId | tar -C buildDir/newId -zxf -`
func AddStage(oldId, newId string) error {
	// prepare location for extraction
	err := os.MkdirAll(config.BuildDir+"/"+newId, 0755)
	if err != nil {
		return fmt.Errorf("Failed to create build dir - %v", err)
	}

	// backend.ReadBlob(oldId) | tar -C buildDir/newId -zxf -
	if oldId != "" {
		// todo: can this stream somehow? send a *reader to readblob, goroutine readblob
		//   then by assigning the *reader to cmd.Stdin, it shouldn't block (keep from
		//   storing build in memory)?? maybe
		// get last build from backend
		res, err := backend.ReadBlob(oldId)
		if err != nil {
			return fmt.Errorf("Failed to get old build - %v", err)
		}

		config.Log.Trace("Fetched build")

		// prepare to extract to new build dir
		cmd := exec.Command("tar", "-C", newId, "-zxf", "-")
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

	err = ssh.AddUser(newId)
	if err != nil {
		return fmt.Errorf("Failed to add user - %v", err)
	}

	mutex.Lock()
	builds = append(builds, newId)
	mutex.Unlock()

	return nil
}

// CommitStage compresses the new build, uploads it to the backend and removes
// the user secret from the ssh server.
// Bash equivalent:
//  `tar -C buildDir/buildId -czf - . | curl localhost:7410/blobs/newId --data-binary @-`
func CommitStage(buildId string) error {
	// remove user first
	err := getUser(buildId)
	if err == nil {
		err = ssh.DelUser(buildId)
		if err != nil {
			return fmt.Errorf("Failed to remove user - %v", err)
		}
	}

	// define buffer
	var blob bytes.Buffer

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

	// pipe compressed build to write command
	cmd.Stdout = &blob

	config.Log.Trace("Running compress command '%v'", cmd.Args)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to compress build - %v", err)
	}
	config.Log.Trace("Compressed build")

	// write build to backend
	err = backend.WriteBlob(buildId, &blob)
	if err != nil {
		return fmt.Errorf("Failed to write build - %v", err)
	}
	config.Log.Trace("Uploaded build")

	return nil
}

// DeleteStage removes files for a specific build.
func DeleteStage(buildId string) error {
	// remove user first
	err := getUser(buildId)
	if err != nil {
		err = ssh.DelUser(buildId)
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
		if builds[i] == buildId {
			builds = append(builds[:i], builds[i+1:]...)
			break
		}
	}
	mutex.Unlock()

	return nil
}

// getUser gets the user secret corresponding to an uncommitted build.
func getUser(buildId string) error {
	for _, build := range builds {
		if build == buildId {
			return nil
		}
	}
	return fmt.Errorf("No Build Found")
}
