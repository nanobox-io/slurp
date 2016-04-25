package api

import (
	"net/http"

	"github.com/nanopack/slurp/core"
)

type build struct {
	// for whatever reason, these need to be exported so json.[un]marshal can utilize it
	OldId string `json:"old-id"`
	NewId string `json:"new-id"`
}

type auth struct {
	AuthSecret string `json:"secret"`
}

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

	// stage the build
	secret, err := slurp.AddStage(stage.OldId, stage.NewId)
	if err != nil {
		writeBody(rw, req, apiError{err.Error()}, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, auth{secret}, http.StatusOK)
}

func commitStage(rw http.ResponseWriter, req *http.Request) {
	// PUT /stages/{buildId}
	buildId := req.URL.Query().Get(":buildId")

	// commit the staged build
	err := slurp.CommitStage(buildId)
	if err != nil {
		writeBody(rw, req, apiError{err.Error()}, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}

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
