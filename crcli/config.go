package crcli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/coderockit/viper"
	"github.com/juju/loggo"
	"gopkg.in/urfave/cli.v1"
)

const codeRockItDirName = ".coderockit"

var ConfigLogger loggo.Logger
var CmdsLogger loggo.Logger
var CrcliLogger loggo.Logger
var FileioLogger loggo.Logger
var HashLogger loggo.Logger

var apiAccessTokens []interface{}

func InitViper(configDir string) error {
	viper.SetConfigType("json")
	viper.SetConfigName("coderockit") // name of config file (without extension)
	if configDir != "" {
		viper.AddConfigPath(configDir) // optional configDir passed in from the command line
	}
	viper.AddConfigPath(".")                  // looking for config in the working directory
	viper.AddConfigPath(GetHomeCRDirectory()) // call multiple times to add many search paths
	viper.AddConfigPath(
		filepath.Join(GetCurrentDirectoryRootPath(), "etc", "coderockit")) // path to look for the config file in
	err := viper.ReadInConfig()
	return err
}

func LoadConfiguration(c *cli.Context, configDir string) bool {
	ready := true

	commandName := c.Command.FullName()
	//fmt.Printf("Command: %s\n", commandName)
	initErr := InitViper(configDir)
	if initErr != nil && commandName == "init" {
		if strings.Contains(fmt.Sprintf("%s", initErr), "Not Found") {
			_, statErr := os.Stat("./coderockit.json")
			if statErr == nil {
				// file ./coderockit.json ALREADY exists, something must have gone wrong
				panic(fmt.Errorf("Fatal error reading coderockit.json config file: %s \n", initErr))
			} else {
				// create a new default ./coderockit.json file
				defaultCodeRockItJson := "{\n" +
					"	\"logger.config\": {\n" +
					"		\"root\": \"ERROR\",\n" +
					"		\"coderockit.cli.config\": \"INFO\",\n" +
					"		\"coderockit.cli.hash\": \"INFO\",\n" +
					"		\"coderockit.cli.cmds\": \"INFO\",\n" +
					"		\"coderockit.cli.fileio\": \"INFO\",\n" +
					"		\"coderockit.cli.crcli\": \"INFO\"\n" +
					"	},\n" +
					"	\"apiAllowInsecure\": false,\n" +
					"	\"registerUser\": \"https://coderockit.io/ui/v1/auth/realms/coderockit/account\",\n" +
					"	\"apiURLs\": [\"https://coderockit.io/api/v1\"]\n" +
					"}"
				codeRockItJsonFile := filepath.Join(".", "coderockit.json")
				fmt.Printf("Creating default coderockit.json file: %s\n", AbsPath(codeRockItJsonFile))
				err := ioutil.WriteFile(codeRockItJsonFile, []byte(defaultCodeRockItJson), 0644)
				if err == nil {
					err := InitViper(configDir)
					if err != nil {
						panic(fmt.Errorf("Fatal error initializing coderockit configuration: %s \n", err))
					}
				} else {
					panic(fmt.Errorf("Fatal error creating default coderockit.json config file: %s \n", err))
				}
			}
		} else {
			panic(fmt.Errorf("Fatal error reading coderockit.json config file: %s \n", initErr))
		}
	} else if initErr != nil && commandName != "init" {
		fmt.Printf("You must first run the init command: 'cr init'\nExample: cr init\n")
		ready = false
	}

	if initErr == nil || commandName == "init" {
		// "<root>=ERROR; coderockit.cli.main=DEBUG; coderockit.cli.config=DEBUG; coderockit.cli.hash=DEBUG; coderockit.cli.cmds=DEBUG; coderockit.cli.fileio=DEBUG; coderockit.cli.crcli=DEBUG"
		loggerConfigMap := viper.GetStringMap("logger.config")
		loggerConfig := ""
		for loggerName := range loggerConfigMap {
			if loggerName == "root" {
				loggerConfig += "<root>=" + loggerConfigMap[loggerName].(string) + "; "
			} else {
				loggerConfig += loggerName + "=" + loggerConfigMap[loggerName].(string) + "; "
			}
		}
		loggo.ConfigureLoggers(loggerConfig)
		ConfigLogger = loggo.GetLogger("coderockit.cli.config")
		CmdsLogger = loggo.GetLogger("coderockit.cli.cmds")
		CrcliLogger = loggo.GetLogger("coderockit.cli.crcli")
		FileioLogger = loggo.GetLogger("coderockit.cli.fileio")
		HashLogger = loggo.GetLogger("coderockit.cli.hash")

		if configDir != "" {
			ConfigLogger.Debugf("CodeRockIt --config directory: %q\n", configDir)
		}
		filename, configErr := GetConfigFilename()
		if configErr == nil {
			fmt.Printf("Using config file: %s\n", filename)
		} else {
			ConfigLogger.Debugf("Error trying to get config filename: %s\n", configErr)
		}
	}

	if commandName == "init" {
		// Create the .coderockit directory in the current directory if it does not exist
		dotcr := GetCRDirectory()
		if !PathExists(dotcr) {
			fmt.Printf("Creating directory: %s\n", AbsPath(dotcr))
			if err := os.MkdirAll(dotcr, os.ModePerm); err != nil {
				ConfigLogger.Debugf("Cannot create the %s directory: %s", dotcr, err)
			}
		}

		// Create the .coderockit/apply directory if it does not exist
		dotcrApply := GetApplyDirectory()
		if !PathExists(dotcrApply) {
			fmt.Printf("Creating directory: %s\n", AbsPath(dotcrApply))
			if err := os.MkdirAll(dotcrApply, os.ModePerm); err != nil {
				ConfigLogger.Debugf("Cannot create the %s directory: %s", dotcrApply, err)
			}
		}

		// Create the $HOME/.coderockit directory if it does not exist
		homeDotcr := GetHomeCRDirectory()
		if homeDotcr != "" {
			if !PathExists(homeDotcr) {
				fmt.Printf("Creating directory: %s\n", AbsPath(homeDotcr))
				if err := os.MkdirAll(homeDotcr, os.ModePerm); err != nil {
					ConfigLogger.Debugf("Cannot create the %s directory: %s", homeDotcr, err)
				} else {
					dotcrCache := GetHomeCacheDirectory()
					if !PathExists(dotcrCache) {
						fmt.Printf("Creating directory: %s\n", AbsPath(dotcrCache))
						if err := os.MkdirAll(dotcrCache, os.ModePerm); err != nil {
							ConfigLogger.Debugf("Cannot create the %s directory: %s", dotcrCache, err)
						}
					}
				}
			}
		}
	}

	InitApiAccessTokenMap()

	return ready
}

func GetHomeCRDirectory() string {

	user, err := user.Current()
	if err == nil {
		homeCRDir := filepath.Join(user.HomeDir, codeRockItDirName)
		//ConfigLogger.Debugf("Home .coderockit Dir: %s", homeCRDir)
		return homeCRDir
	} else {
		ConfigLogger.Criticalf("Error: %s", err)
	}
	return ""
}

func GetHomeCacheDirectory() string {
	return filepath.Join(GetHomeCRDirectory(), "cache")
}

func GetHomeConfigFile() string {
	return filepath.Join(GetHomeCRDirectory(), "config.json")
}

func CurrentTimestampMillis() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func StoreApiAccessToken(tokIndex int, accessToken map[string]interface{}) {
	accessToken["cr-timestamp"] = strconv.FormatInt(CurrentTimestampMillis(), 10)

	if GetApiAccessTokenMap(tokIndex) == nil {
		apiAccessTokens = append(apiAccessTokens, accessToken)
	} else {
		// we must already have an access token array, find out where to put this new one
		// and then store it off to the config.json file
		ConfigLogger.Debugf("We already have an access token array!\n")
		apiAccessTokens[tokIndex] = accessToken
	}

	var configJson map[string]interface{} = make(map[string]interface{})
	configJson["apiAccessTokens"] = apiAccessTokens
	jsonString, err := json.Marshal(configJson)
	if err == nil {
		configJsonFile := filepath.Join(GetHomeCRDirectory(), "config.json")
		ConfigLogger.Debugf("Writing access token to file: %s\n", configJsonFile)
		err := ioutil.WriteFile(configJsonFile, []byte(jsonString), 0644)
		if err != nil {
			ConfigLogger.Debugf("Fatal error creating '%s' file: %s \n", configJsonFile, err)
		}
	} else {
		ConfigLogger.Debugf("Error marshalling object to JSON: %s", err)
	}
}

func InitApiAccessTokenMap() {
	if apiAccessTokens == nil {
		homeConfig := viper.New()
		homeConfig.SetConfigType("json")
		homeConfig.SetConfigName("config")             // name of config file (without extension)
		homeConfig.AddConfigPath(GetHomeCRDirectory()) // call multiple times to add many search paths
		homeConfig.AddConfigPath(
			filepath.Join(GetCurrentDirectoryRootPath(), "etc", "coderockit")) // path to look for the config file in
		err := homeConfig.ReadInConfig()
		if err != nil {
			ConfigLogger.Debugf("Fatal error reading config.json config file: %s \n", err)
		}

		if homeConfig.IsSet("apiAccessTokens") {
			apiAccessTokens = homeConfig.Get("apiAccessTokens").([]interface{})
		}
	}
}

func GetApiAccessTokenMap(tokIndex int) map[string]interface{} {
	if len(apiAccessTokens) > 0 {
		return apiAccessTokens[tokIndex].(map[string]interface{})
	} else {
		return nil
	}
}

func GetValidApiAccessToken(tokIndex int) string {
	username, tokenProblem := CheckToken(tokIndex)
	if tokenProblem || username == "" {
		GetNewAccessToken()
	}

	return GetApiAccessToken(tokIndex)
}

func GetApiAccessToken(tokIndex int) string {
	accessTokenMap := GetApiAccessTokenMap(tokIndex)
	if accessTokenMap == nil {
		return ""
	} else {
		return accessTokenMap["access_token"].(string)
	}
}

func ApiAccessTokenIsTooOld(tokIndex int) bool {
	//compare (cr-timestamp) to (expires_in * 1000) to see if more than CurrentTimestampMillis
	//and if it is then the access token is too old and we need to get a new one
	accessTokenMap := GetApiAccessTokenMap(tokIndex)
	if accessTokenMap == nil {
		return true
	} else {
		// make sure to give 3000 milliseconds to hedge a slow server side token verification
		crTimestamp, _ := strconv.ParseInt(accessTokenMap["cr-timestamp"].(string), 10, 64)
		return crTimestamp-3000+(int64(accessTokenMap["expires_in"].(float64))*1000) <
			CurrentTimestampMillis()
	}
}

func GetCurrentDirectoryRootPath() string {
	currVolPath := GetCurrentDirectory()
	var volName string = ""
	if strings.HasPrefix(currVolPath, "/") {
		volName = "/"
	} else {
		colonIndex := strings.Index(currVolPath, ":")
		if colonIndex == -1 {
			volName = "C:\\"
		} else {
			volName = currVolPath[0:colonIndex+1] + "\\"
		}
	}

	//fmt.Printf("The volume name is: %s\n", volName)
	return volName
}

func GetCurrentDirectory() string {
	currentDir, err := filepath.Abs(".")
	if err != nil {
		ConfigLogger.Debugf("Could not get current working directory!!")
	}
	return currentDir
}

func GetCRDirectory() string {
	return filepath.Join(".", codeRockItDirName)
}

func GetApplyDirectory() string {
	return filepath.Join(".", codeRockItDirName, "apply")
}

func GetPinsToApplyFile() string {
	return filepath.Join(".", codeRockItDirName, "pinsToApply.json")
}

func GetConfigFilename() (string, error) {
	return viper.GetConfigFile()
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
