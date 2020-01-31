package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/nanobox-io/slurp/core"
)

// for whatever reason, these need to be exported so json.[un]marshal can utilize it
type build struct {
	OldId string `json:"old-id"` // build to fetch from storage
	NewId string `json:"new-id"` // build to stage and store
}

type auth struct {
	AuthSecret string `json:"secret"`
}

// addStage prepares a directory for receiving the new build. If an old build is specified,
// that build is fetched from hoarder, otherwise a new directory is created.
func addStage(rw http.ResponseWriter, req *http.Request) {
	var stage build
	err := parseBody(req, &stage)
	if err != nil {
		writeBody(rw, req, apiError{err.Error()}, http.StatusBadRequest)
		return
	}

	if stage.NewId == "" {
		writeBody(rw, req, apiError{"Missing Payload Data"}, http.StatusInternalServerError)
		return
	}

	secret, err := generateSecret()
	if err != nil {
		writeBody(rw, req, apiError{ErrorString: "internal error"}, http.StatusInternalServerError)
		return
	}

	// stage the build
	err = slurp.AddStage(stage.OldId, stage.NewId, secret)
	if err != nil {
		writeBody(rw, req, apiError{err.Error()}, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, auth{AuthSecret: secret}, http.StatusOK)
}

// commitStage is called once the local build is synced with the staged build. It will
// compress and upload the staged build to hoarder. CommitStage will also remove the
// user for security.
func commitStage(rw http.ResponseWriter, req *http.Request) {
	// PUT /stages/{buildId}
	buildId := req.URL.Query().Get(":buildId")

	// commit the staged build
	err := slurp.CommitStage(buildId)
	if err != nil {
		writeBody(rw, req, apiError{err.Error()}, http.StatusInternalServerError)
		return
	}

	// delete the staged build
	err = slurp.DeleteStage(buildId)
	if err != nil {
		writeBody(rw, req, apiError{err.Error()}, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}

// deleteStage removes the staged build directory
func deleteStage(rw http.ResponseWriter, req *http.Request) {
	// DELETE /stages/{buildId}
	buildId := req.URL.Query().Get(":buildId")

	// delete the staged build
	err := slurp.DeleteStage(buildId)
	if err != nil {
		writeBody(rw, req, apiError{err.Error()}, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}

// generateSecret creates a new cryptographically secure secret
func generateSecret() (string, error) {
	var b [32]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
