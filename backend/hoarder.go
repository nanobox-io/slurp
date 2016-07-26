package backend

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/nanobox-io/slurp/config"
)

type hoarder struct {
	proto string
}

// ensure hoarder is up
func (self hoarder) initialize() error {
	_, err := self.rest("GET", "ping", nil)
	return err
}

// get blob from hoarder and return Reader for piping to next command
func (self hoarder) readBlob(id string) (io.ReadCloser, error) {
	res, err := self.rest("GET", "blobs/"+id, nil)
	if err != nil { // prevent panic if no res
		return nil, err
	}
	return res.Body, err
}

// pipe blob to hoarder
func (self hoarder) writeBlob(id string, blob io.Reader) error {
	_, err := self.rest("POST", "blobs/"+id, blob)
	return err
}

// rest is a helper method http client to interact with hoarder
func (self hoarder) rest(method, path string, body io.Reader) (*http.Response, error) {
	config.Log.Trace("[client] - %v hoarder/%v", method, path)
	var client *http.Client
	client = http.DefaultClient
	uri := fmt.Sprintf("%s://%s/%s", self.proto, storeAddr, path)

	// if insecure is false, verify cert
	if config.Insecure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-AUTH-TOKEN", config.StoreToken)
	res, err := client.Do(req)
	if err != nil {
		// return original error to client
		return nil, err
	}
	if res.StatusCode == 401 {
		return nil, fmt.Errorf("401 Unauthorized. Please specify backend api token (-T 'backend-token')")
	}
	return res, nil
}
