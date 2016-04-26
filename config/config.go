// config is a central location for configuration options. It also contains
// config file parsing logic.
package config

import (
	"fmt"
	"path/filepath"

	"github.com/jcelliott/lumber"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ApiToken   = "secret"
	ApiAddress = "127.0.0.1:1566"
	BuildDir   = "/var/db/slurp/build/"
	ConfigFile = ""
	Insecure   = false
	LogLevel   = "info"
	SshAddr    = "127.0.0.1:1567"
	SshHostKey = "/var/db/slurp/slurp_rsa" // "host-file"
	StoreAddr  = "hoarder://127.0.0.1:7410"
	StoreToken = ""

	Server = true
	Log    lumber.Logger
)

// AddFlags adds the available cli flags
func AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&ApiToken, "api-token", "t", ApiToken, "Token for API Access")
	cmd.PersistentFlags().StringVarP(&ApiAddress, "api-address", "a", ApiAddress, "Listen address for the API")
	cmd.PersistentFlags().StringVarP(&BuildDir, "build-dir", "b", BuildDir, "Build staging directory")
	cmd.PersistentFlags().StringVarP(&ConfigFile, "config-file", "c", ConfigFile, "Configuration file to load")
	cmd.PersistentFlags().BoolVarP(&Insecure, "insecure", "i", Insecure, "Disable tls key checking (client) and listen on http (server)")
	cmd.PersistentFlags().StringVarP(&LogLevel, "log-level", "l", LogLevel, "Log level to output [fatal|error|info|debug|trace]")

	cmd.PersistentFlags().StringVarP(&SshAddr, "ssh-addr", "s", SshAddr, "Address ssh server will listen on (ip:port combo)")
	cmd.PersistentFlags().StringVarP(&SshHostKey, "ssh-host", "k", SshHostKey, "SSH host (private) key file")

	cmd.PersistentFlags().StringVarP(&StoreAddr, "store-addr", "S", StoreAddr, "Storage host address")
	cmd.PersistentFlags().StringVarP(&StoreToken, "store-token", "T", StoreToken, "Storage auth token")
}

// LoadConfigFile reads the specified config file
func LoadConfigFile() error {
	if ConfigFile == "" {
		return nil
	}

	// Set defaults to whatever might be there already
	viper.SetDefault("api-token", ApiToken)
	viper.SetDefault("api-address", ApiAddress)
	viper.SetDefault("build-dir", BuildDir)
	viper.SetDefault("config-file", ConfigFile)
	viper.SetDefault("insecure", Insecure)
	viper.SetDefault("log-level", LogLevel)
	viper.SetDefault("ssh-addr", SshAddr)
	viper.SetDefault("ssh-host", SshHostKey)
	viper.SetDefault("store-addr", StoreAddr)
	viper.SetDefault("store-token", StoreToken)

	filename := filepath.Base(ConfigFile)
	viper.SetConfigName(filename[:len(filename)-len(filepath.Ext(filename))])
	viper.AddConfigPath(filepath.Dir(ConfigFile))

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Failed to read config file - %v", err)
	}

	// Set values. Config file will override commandline
	ApiToken = viper.GetString("api-token")
	ApiAddress = viper.GetString("api-address")
	BuildDir = viper.GetString("build-dir")
	ConfigFile = viper.GetString("config-file")
	Insecure = viper.GetBool("insecure")
	LogLevel = viper.GetString("log-level")
	SshAddr = viper.GetString("ssh-addr")
	SshHostKey = viper.GetString("ssh-host")
	StoreAddr = viper.GetString("store-addr")
	StoreToken = viper.GetString("store-token")

	return nil
}
