// Package "backend" is the layer between slurp and the long stored blobs.
package backend

import (
	"fmt"
	"io"
	"net/url"

	"github.com/nanopack/slurp/config"
)

type blobReadWriter interface {
	initialize() error
	readBlob(id string) (io.ReadCloser, error)
	writeBlob(id string, blob io.Reader) error
}

// the pluggable (future) backend
var backend blobReadWriter

// Initialize prepares the backend and ensures it is available
func Initialize() error {
	var err error
	var u *url.URL
	u, err = url.Parse(config.StoreAddr)
	if err != nil {
		return fmt.Errorf("Failed to parse backend connection - %v", err)
	}
	switch u.Scheme {
	case "hoarder":
		backend = &hoarder{}
	default:
		backend = &hoarder{}
	}

	config.StoreAddr = u.Host
	return backend.initialize()
}

// ReadBlob reads a blob from a storage backend
func ReadBlob(id string) (io.ReadCloser, error) {
	return backend.readBlob(id)
}

// WriteBlob writes a blob to a storage backend
func WriteBlob(id string, blob io.Reader) error {
	return backend.writeBlob(id, blob)
}
