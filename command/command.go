package command

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"

	"github.com/nanopack/slurp/api"
	"github.com/nanopack/slurp/backend"
	"github.com/nanopack/slurp/config"
	"github.com/nanopack/slurp/ssh"
)

var (
	runServer bool
	Slurp     = &cobra.Command{
		Use:   "slurp",
		Short: "slurp - build intermediary",
		Long:  ``,

		Run: startSlurp,
	}
)

func init() {
	config.AddFlags(Slurp)
}

func startSlurp(ccmd *cobra.Command, args []string) {
	if err := config.LoadConfigFile(); err != nil {
		config.Log.Fatal("Failed to read config - %v", err)
		os.Exit(1)
	}

	if !config.Server {
		ccmd.HelpFunc()(ccmd, args)
		return
	}

	config.Log = lumber.NewConsoleLogger(lumber.LvlInt(config.LogLevel))

	// initialize backend
	err := backend.Initialize()
	if err != nil {
		config.Log.Fatal("Backend init failed - %v", err)
		os.Exit(1)
	}

	// start ssh server
	err = ssh.Start()
	if err != nil {
		config.Log.Fatal("SSH server start failed - %v", err)
		os.Exit(1)
	}

	// start api
	err = api.StartApi()
	if err != nil {
		config.Log.Fatal("Api start failed - %v", err)
		os.Exit(1)
	}
	return
}

func rest(path string, method string, body io.Reader) (*http.Response, error) {
	var client *http.Client
	client = http.DefaultClient
	uri := fmt.Sprintf("https://%s:%s/%s", config.ApiHost, config.ApiPort, path)

	if config.Insecure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		panic(err)
	}
	req.Header.Add("X-AUTH-TOKEN", config.ApiToken)
	res, err := client.Do(req)
	if err != nil {
		// if requesting `https://` failed, server may have been started with `-i`, try `http://`
		uri = fmt.Sprintf("http://%s:%s/%s", config.ApiHost, config.ApiPort, path)
		req, er := http.NewRequest(method, uri, body)
		if er != nil {
			panic(er)
		}
		req.Header.Add("X-AUTH-TOKEN", config.ApiToken)
		var err2 error
		res, err2 = client.Do(req)
		if err2 != nil {
			// return original error to client
			return nil, err
		}
	}
	if res.StatusCode == 401 {
		return nil, fmt.Errorf("401 Unauthorized. Please specify api token (-t 'token')")
	}
	return res, nil
}

func fail(format string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("%v\n", format), args...)
	os.Exit(1)
}
