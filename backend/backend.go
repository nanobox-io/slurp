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

var Backend blobReadWriter // exportable for testing.. todo: if needed

func Initialize() error {
	var err error
	var u *url.URL
	u, err = url.Parse(config.StoreAddr)
	if err != nil {
		return fmt.Errorf("Failed to parse db connection - %v", err)
	}
	switch u.Scheme {
	case "hoarder":
		Backend = &hoarder{}
	default:
		Backend = &hoarder{}
	}

	config.StoreAddr = u.Host
	return Backend.initialize()
}

// ReadBlob reads a blob from a storage backend
func ReadBlob(id string) (io.ReadCloser, error) {
	return Backend.readBlob(id)
}

// WriteBlob writes a blob to a storage backend
func WriteBlob(id string, blob io.Reader) error {
	return Backend.writeBlob(id, blob)
}
