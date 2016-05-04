// Package "config" is a central location for configuration options. It also contains
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
	ApiToken   = "secret"                   // Token for API Access
	ApiAddress = "127.0.0.1:1566"           // Listen address for the API
	BuildDir   = "/var/db/slurp/build/"     // Build staging directory
	ConfigFile = ""                         // Configuration file to load
	Insecure   = false                      // Disable tls key checking (client) and listen on http (server)
	LogLevel   = "info"                     // Log level to output [fatal|error|info|debug|trace]
	SshAddr    = "127.0.0.1:1567"           // Address ssh server will listen on (ip:port combo)
	SshHostKey = "/var/db/slurp/slurp_rsa"  // SSH host (private) key file
	StoreAddr  = "hoarder://127.0.0.1:7410" // Storage host address
	StoreSSL   = false                      // Disable tls key checking (client) and listen on http (server)
	StoreToken = ""                         // Storage auth token
	Version    = false                      // Print version info and exit

	Log lumber.Logger // Central logger for slurp
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
	cmd.PersistentFlags().BoolVarP(&StoreSSL, "store-ssl", "I", StoreSSL, "Enable tls certificate verification when connecting to storage")

	cmd.Flags().BoolVarP(&Version, "version", "v", Version, "Print version info and exit")
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
	viper.SetDefault("store-ssl", StoreSSL)
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
	StoreSSL = viper.GetBool("store-ssl")
	StoreToken = viper.GetString("store-token")

	return nil
}
