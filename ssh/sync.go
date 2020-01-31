package ssh

import (
	"sync"

	"github.com/nanobox-io/slurp/config"
)

type user struct {
	Name    string
	BuildID string
}

var (
	// copy of all non-committed users
	authUsers []user

	// mutex ensures updates to authUsers are atomic
	mutex = sync.Mutex{}
)

// Add an authorized user
func AddUser(name, directory string) error {
	config.Log.Trace("Adding user %v dir %v", name, directory)
	mutex.Lock()
	authUsers = append(authUsers, user{
		Name:    name,
		BuildID: directory,
	})
	mutex.Unlock()

	return nil
}

// Remove an authorized user
func DelUser(user string) error {
	config.Log.Trace("Removing user %v", user)
	mutex.Lock()
	for i := range authUsers {
		if authUsers[i].Name == user {
			authUsers = append(authUsers[:i], authUsers[i+1:]...)
			break
		}
	}
	mutex.Unlock()

	return nil
}
