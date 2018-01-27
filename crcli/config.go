package crcli

import (
	"fmt"
	"time"

	"github.com/coderockit/viper"
	"github.com/juju/loggo"
)

func LoadConfiguration(configDir string) {
	viper.SetConfigType("json")
	viper.SetConfigName("coderockit") // name of config file (without extension)
	if configDir != "" {
		viper.AddConfigPath(configDir) // optional configDir passed in from the command line
	}
	viper.AddConfigPath(".")                  // looking for config in the working directory
	viper.AddConfigPath("$HOME/.coderockit/") // call multiple times to add many search paths
	viper.AddConfigPath("/etc/coderockit/")   // path to look for the config file in
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error reading coderockit.json config file: %s \n", err))
	}

	loggo.ConfigureLoggers(viper.GetString("logger.config"))
	logger := loggo.GetLogger("coderockit.cli.config")

	if configDir != "" {
		logger.Debugf("CodeRockIt --config directory: %q\n", configDir)
	}
	filename, err := viper.GetConfigFile()
	logger.Debugf("Using config file: %s\n", filename)
}

// GetInt(key string) : int
func ConfInt(name string, def int) int {
	if viper.IsSet(name) {
		return viper.GetInt(name)
	}
	return def
}

// Get(key string) : interface{}
func Conf(name string, def interface{}) interface{} {
	if viper.IsSet(name) {
		return viper.Get(name)
	}
	return def
}

// GetBool(key string) : bool
func ConfBool(name string, def bool) bool {
	if viper.IsSet(name) {
		return viper.GetBool(name)
	}
	return def
}

// GetFloat64(key string) : float64
func ConfFloat64(name string, def float64) float64 {
	if viper.IsSet(name) {
		return viper.GetFloat64(name)
	}
	return def
}

// GetString(key string) : string
func ConfString(name string, def string) string {
	if viper.IsSet(name) {
		return viper.GetString(name)
	}
	return def
}

// GetStringMap(key string) : map[string]interface{}
func ConfStringMap(name string, def map[string]interface{}) map[string]interface{} {
	if viper.IsSet(name) {
		return viper.GetStringMap(name)
	}
	return def
}

// GetStringMapString(key string) : map[string]string
func ConfStringMapString(name string, def map[string]string) map[string]string {
	if viper.IsSet(name) {
		return viper.GetStringMapString(name)
	}
	return def
}

// GetStringSlice(key string) : []string
func ConfStringSlice(name string, def []string) []string {
	if viper.IsSet(name) {
		return viper.GetStringSlice(name)
	}
	return def
}

// GetTime(key string) : time.Time
func ConfTime(name string, def time.Time) time.Time {
	if viper.IsSet(name) {
		return viper.GetTime(name)
	}
	return def
}

// GetDuration(key string) : time.Duration
func ConfDuration(name string, def time.Duration) time.Duration {
	if viper.IsSet(name) {
		return viper.GetDuration(name)
	}
	return def
}
