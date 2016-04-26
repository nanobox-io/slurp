// Package "api" defines the routes accessible and the logic when they are hit.
package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/pat"
	"github.com/nanobox-io/golang-nanoauth"

	"github.com/nanobox-io/slurp/config"
)

var (
	badJson      = errors.New("Bad JSON Syntax Received in Body")
	bodyReadFail = errors.New("Body Read Failed")
)

type (
	apiError struct {
		ErrorString string `json:"error"`
	}
	apiMsg struct {
		MsgString string `json:"msg"`
	}
)

// start the web server
func StartApi() error {
	if config.Insecure {
		config.Log.Info("Api listening at http://%s...", config.ApiAddress)
		return http.ListenAndServe(config.ApiAddress, routes())
	}

	var auth nanoauth.Auth
	cert, err := nanoauth.Generate("slurp.nanobox.io")
	if err != nil {
		return err
	}
	auth.Certificate = cert
	auth.Header = "X-AUTH-TOKEN"

	config.Log.Info("Api listening at https://%s...", config.ApiAddress)
	return auth.ListenAndServeTLS(config.ApiAddress, config.ApiToken, routes())
}

// api routes
func routes() *pat.Router {
	router := pat.New()

	// keep "/stages" so a build named "ping" won't break anything
	router.Post("/stages", addStage)
	router.Put("/stages/{buildId}", commitStage)
	router.Delete("/stages/{buildId}", deleteStage)

	router.Get("/ping", pong)

	return router
}

// write the json body and log the request
func writeBody(rw http.ResponseWriter, req *http.Request, v interface{}, status int) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	// print the error only if there is one
	var msg map[string]string
	json.Unmarshal(b, &msg)

	var errMsg string
	if msg["error"] != "" {
		errMsg = msg["error"]
	}

	config.Log.Debug("%s %d %s %s %s", req.RemoteAddr, status, req.Method, req.RequestURI, errMsg)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	rw.Write(append(b, byte('\n')))

	return nil
}

// parseBody parses the json body into v
func parseBody(req *http.Request, v interface{}) error {

	// read the body
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		config.Log.Error(err.Error())
		return bodyReadFail
	}
	defer req.Body.Close()

	// parse body and store in v
	err = json.Unmarshal(b, v)
	if err != nil {
		return badJson
	}

	return nil
}

// reply pong (life check)
func pong(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("pong\n"))
}
