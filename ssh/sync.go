package ssh

import (
	"sync"

	"github.com/nanopack/slurp/config"
)

var (
	// copy of all non-committed users
	authUsers []string

	// mutex ensures updates to authUsers are atomic
	mutex = sync.Mutex{}
)

// Add an authorized user
func AddUser(user string) error {
	config.Log.Trace("Adding user %v", user)
	mutex.Lock()
	authUsers = append(authUsers, user)
	mutex.Unlock()

	return nil
}

// Remove an authorized user
func DelUser(user string) error {
	config.Log.Trace("Removing user %v", user)
	mutex.Lock()
	for i := range authUsers {
		if authUsers[i] == user {
			authUsers = append(authUsers[:i], authUsers[i+1:]...)
			break
		}
	}
	mutex.Unlock()

	return nil
}
