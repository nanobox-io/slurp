package ssh

import (
	"sync"
)

var (
	// copy of all non-committed users
	authUsers []string

	// mutex ensures updates to authUsers are atomic
	mutex = sync.Mutex{}
)

// Add an authorized user
func AddUser() (string, error) {
	user := genUser()

	mutex.Lock()
	authUsers = append(authUsers, user)
	mutex.Unlock()

	return user, nil
}

// Remove an authorized user
func DelUser(user string) error {
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
