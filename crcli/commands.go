package crcli

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	//	"github.com/PuerkitoBio/goquery"
	"github.com/dgrijalva/jwt-go"
	"github.com/fatih/color"
	"gopkg.in/headzoo/surf.v1"
	"gopkg.in/headzoo/surf.v1/agent"
	"gopkg.in/resty.v1"
	"gopkg.in/urfave/cli.v1"
)

func Init(args cli.Args) {
	//CmdsLogger.Debugf("init new project: %s", "need to implement")

	//How to use keycloak to create a new user via a web service /coderockit/users
	//https://gist.github.com/thomasdarimont/c4e739c5a319cf78a4cff3b87173a84b

	var username string = ""
	var tokenProblem bool = false

	apiURLs := ConfStringSlice("apiURLs", defaultApiUrls)
	for tokIndex, apiURL := range apiURLs {
		accessToken := GetApiAccessToken(tokIndex)
		CmdsLogger.Debugf("The url is %s and the token is %s\n", apiURL, accessToken)
		if accessToken != "" {
			token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
				//CmdsLogger.Debugf("token is valid: %s\ntoken headers: %s\ntoken claims: %s\n", token.Valid, token.Header, token.Claims)
				tokenKey, err := base64.StdEncoding.DecodeString("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAgA3A/R2TL8aLpLCLF8t+t3IaYTR/xX4UATaaicA8+7AK+K/eGqrdjPxW8HJBJurzGgxqsVY5n/FE08af+baWcm06yLmvVk83V8/90vUu5EFNm4DcUvQXIwoLy9rnjUJRLQ/+F+3hzKaOwxhVyCJQ4cpnOjdfawnDB5m3cVA7t+Y1d3yC9QBGyn8+JohGMRpQt3EEX8PypVzc+sB85j8oBvZxTvHM9uDj6Pj1uY46xt2pHr2Nk8vqJYCmqSwH/ctDyNFADfgdlc9JpsqV8zjgn0B7t/bjMpFkFnIf7f6dUtygcj7Eqp5uVCfKK52hGv/FBCljeex/27iJ920GfJrNRQIDAQAB")
				if err == nil {
					return tokenKey, nil
				} else {
					return nil, err
				}
			})

			if token.Claims != nil {
				username = token.Claims.(jwt.MapClaims)["preferred_username"].(string)
			}

			if token.Valid {
				CmdsLogger.Debugf("Valid access token found")
			} else if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					tokenProblem = true
					CmdsLogger.Debugf("[1] Incorrect access token: ", err)
				} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
					// Token is either expired or not active yet
					tokenProblem = true
					CmdsLogger.Debugf("[2] Access token has expired or is not yet active: ", err)
				} else {
					CmdsLogger.Debugf("[3] Couldn't handle this token:", err)
				}
			} else {
				tokenProblem = true
				CmdsLogger.Debugf("[4] Incorrect access token: ", err)
			}
		}
	}

	CmdsLogger.Debugf("Existing username is: '%s'\n", username)
	if username == "" {
		wantToRegisterNewUser := UserInput("New User or Existing User [n/e]: ")
		if wantToRegisterNewUser == "n" {
			bow := surf.NewBrowser()
			bow.SetUserAgent(agent.Chrome())
			if ConfBool("apiAllowInsecure", false) {
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				bow.SetTransport(tr)
			}

			registerUserUrl := ConfString("registerUser", "https://coderockit.io/ui/v1/auth/realms/coderockit/account")
			err := bow.Open(registerUserUrl)
			if err == nil {

				err := bow.Click("#kc-registration a")
				if err == nil {
					//fmt.Println(bow.Title())
					doneRegisteringNewUser := true
					for {
						fm, err := bow.Form("#kc-register-form")
						if err == nil {
							fm.Input("firstName", UserInput("Enter First Name: "))
							fm.Input("lastName", UserInput("Enter Last Name: "))
							fm.Input("email", UserInput("Enter Email Address: "))
							fm.Input("username", UserInput("Enter Username: "))
							fm.Input("password", UserInput("Enter Password: "))
							fm.Input("password-confirm", UserInput("Enter The Same Password Again: "))
							//fmt.Println("The form is: %s", fm)
							err = fm.Submit()
							if err == nil {
								registerFeedback := bow.Find("span.kc-feedback-text")
								if registerFeedback != nil {
									fmt.Println(registerFeedback.Text())
								}
								registerFailed := bow.Find("#kc-register-form")
								if registerFailed.Length() != 0 {
									doneRegisteringNewUser = false
									//fmt.Println("Found form so NOT done: %s", registerFailed)
								}
							} else {
								fmt.Printf("Could not register new user due to error: %s\n", err)
							}
						} else {
							fmt.Printf("Could not register new user due to error: %s\n", err)
						}

						if doneRegisteringNewUser {
							break
						}
					}
				} else {
					fmt.Printf("Could not register new user due to error: %s\n", err)
				}
			} else {
				fmt.Printf("Could not register new user due to error: %s\n", err)
			}
		} else {
			// get a new token
			GetNewAccessToken()
		}
	} else {
		fmt.Printf("Initialized with user: %s\n", username)
		if tokenProblem {
			// get a new token
			fmt.Printf("A new access token is needed!\n")
			GetNewAccessToken()
		}
	}
}

func GetNewAccessToken() {
	username := UserInput("Enter Username: ")
	password := UserInput("Enter Password: ")
	//fmt.Printf("Username: %s\nPassword: %s\n", username, password)

	if ConfBool("apiAllowInsecure", false) {
		resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	apiURLs := ConfStringSlice("apiURLs", defaultApiUrls)
	for tokenIndex, apiURL := range apiURLs {

		tokenURL := apiURL + "/token"

		CmdsLogger.Debugf("Getting access token using URL: %s", tokenURL)

		resp, err := resty.R().
			SetHeader("Accept", "application/json").
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetBody("username=" + username + "&password=" + password).
			Post(tokenURL)

		if err == nil {
			if resp.StatusCode() == 200 {
				respBody := resp.Body()
				var respObj interface{}
				err := json.Unmarshal(respBody, &respObj)
				if err == nil {
					respMap := respObj.(map[string]interface{})
					//respVersions := respMap["versions"].([]interface{})
					//for _, respVer := range respVersions {
					//	matchingVersions = append(matchingVersions, fmt.Sprintf("%s", respVer))
					//}

					CmdsLogger.Debugf("Response from access token CORRECT object: %s", respMap)
					StoreApiAccessToken(tokenIndex, respMap)

					//if matchingVersions != nil {
					//	break
					//}
				} else {
					CmdsLogger.Debugf("Error INCORRECT json in response %s: %s", respBody, err)
					//pin.ApiMsg = fmt.Sprintf("Fatal: could not parse response JSON: %s :: %s", respBody, err)
				}
			} else {
				//myerr = errors.New(string(resp.StatusCode()))
				respBody := string(resp.Body())
				CmdsLogger.Debugf("Fatal %d: getting access token failed with error: %s", resp.StatusCode(), respBody)
			}
		} else {
			CmdsLogger.Criticalf("Error: %s", err)
		}
	}
}

func UserInput(msg string) string {
	fmt.Print(msg)
	var input string
	fmt.Scanln(&input)
	//fmt.Printf("You entered: %s\n", input)
	return input
}

func AddPaths(args cli.Args) {
	//CmdsLogger.Debugf("added file: %s", args.First())

	pinsToApply := ReadInPinsToApply()
	CmdsLogger.Debugf("Found existing pins to apply: %s", pinsToApply)

	// loop over all of args scanning for pin directives
	for index, addPath := range args {
		abs, err := filepath.Abs(addPath)
		if err == nil {
			CmdsLogger.Debugf("Adding path %s at index %d\n", abs, index)
			pinsToApply = AddPathToPins(abs, pinsToApply)
			//CmdsLogger.Debugf("^^^^^ pinsToApply is: %s\n", pinsToApply)
		} else {
			CmdsLogger.Debugf("Could not add path %s due to error: %s", addPath, err)
		}
	}

	pinsToApply = VerifyGetPinsAgainstLocalPutPins(pinsToApply)
	SavePinsToApply(pinsToApply)
}

func ApplyPins(args cli.Args) {
	//CmdsLogger.Debugf("added file: %s", args.First())

	pinsToApply := ReadInPinsToApply()
	//CmdsLogger.Debugf("Found existing pins to apply: %s", pinsToApply)

	onlyDoPuts := true
	for {
		for pinFile := range pinsToApply {
			//CmdsLogger.Debugf("Applying pins in file: %s\n", pinFile)
			pins := pinsToApply[pinFile]
			//var appliedPins []Pin
			for pinIndex, pin := range pins {
				if (onlyDoPuts && pin.IsPut()) || (!onlyDoPuts && pin.IsGet()) {

					pinContent := ""
					//pinContentFile := ""
					if pin.IsPut() {
						_, pinContent = ReadPinContentToPut(pin)
					} else if pin.IsGet() {
						_, pinContent = ReadPinContentToGet(pin)
					}
					pins[pinIndex] = verifyPin(pin, pinContent, true)

					CmdsLogger.Debugf("pin.ApiMsg: %s", pin.ApiMsg)
					if pin.ApiSuccess() {
						//appliedPins = append(appliedPins, pin)
						//if pin.IsPut() {
						//	FinishApplyingPut(pin)
						//} else if pin.IsGet() {
						//	FinishApplyingGet(pinFile, pin)
						//}
					} else {
						//appliedPins = append(appliedPins, pin)
					}
				} else {
					//appliedPins = append(appliedPins, pin)
				}
			}

			//pinsToApply[pinFile] = appliedPins
			if !onlyDoPuts {
				//InsertGetsIntoFile(pinFile, appliedPins)
				InsertGetsIntoFile(pinFile, pins)
			}
			//if len(failedPins) > 0 {
			//	pinsToApply[pinFile] = failedPins
			//} else {
			//	delete(pinsToApply, pinFile)
			//}
		}

		if !onlyDoPuts {
			break
		} else {
			onlyDoPuts = false
		}
	}

	SavePinsToApply(pinsToApply)
}

func RemovePaths(args cli.Args) {

	pinsToApply := ReadInPinsToApply()
	//CmdsLogger.Debugf("Found existing pins to apply: %s", pinsToApply)

	for index, removePath := range args {
		abs, err := filepath.Abs(removePath)
		if err == nil {
			CmdsLogger.Debugf("Removing path %s at index %d\n", abs, index)
			pinsToApply = RemovePathFromPins(abs, pinsToApply)
			//CmdsLogger.Debugf("^^^^^ pinsToApply is: %s\n", pinsToApply)
		} else {
			CmdsLogger.Debugf("Could not remove path %s due to error: %s", removePath, err)
		}
	}

	SavePinsToApply(pinsToApply)

}

func EmptyPinsToApply(args cli.Args) {
	DeleteFileOrDir(GetPinsToApplyFile())
	DeleteDirectoryRecursively(GetApplyDirectory())
}

//GET        /pin/xxx/xxxxx/4.5.4       (1.0.1) src/sdf/sdf/sdf.c
//PUTPRIVATE /pin/xxxxx/xxxxx/4.5.4     (1.0.1) dfd/fdfdfsdf/sdfdfsdf.java
//PUT        /pin/xxxxmmmmx/xxxxx/4.5.4 (1.0.1) poiup/ppipoi/rwwerwr/hkhjjk.js
func ShowStatus(args cli.Args) {
	pinsToApply := ReadInPinsToApply()
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	white := color.New(color.FgHiWhite).SprintFunc()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, white(fmt.Sprintf("CMD\tPin\tVersion\tFile\tStatus")))
	fmt.Fprintln(w, white(fmt.Sprintf("---\t---\t-------\t----\t------")))
	for pinFile := range pinsToApply {
		filePath := GetEndingPath(pinFile, GetCurrentDirectory())
		pins := pinsToApply[pinFile]
		for _, pin := range pins {
			//CmdsLogger.Debugf("Pin: %s", pin)
			//fmt.Printf("      >> Api message ==> %s\n", pin.ApiMsg)
			if pin.ApiSuccess() {
				fmt.Fprintln(w, green(fmt.Sprintf("%s\t%s/%s/%s\t(%s)\t%s\t%s",
					pin.Verb, pin.GroupName, pin.Name,
					pin.Version, pin.ApplyVersion, filePath, pin.ApiMsg[:7])))
			} else {
				fmt.Fprintln(w, red(fmt.Sprintf("%s\t%s/%s/%s\t(%s)\t%s\t%s >>>",
					pin.Verb, pin.GroupName, pin.Name,
					pin.Version, pin.ApplyVersion, filePath, pin.ApiMsg[:14])))
			}
		}
	}
	w.Flush()
}

func ShowDiffs(args cli.Args) {
	pinsToApply := ReadInPinsToApply()
	if len(args) == 0 {
		ShowAllDiffs(pinsToApply, args)
	} else {
		for _, pinOrFile := range args {
			ShowOneDiff(pinsToApply, pinOrFile)
		}
	}
}

func ShowOneDiff(pinsToApply Pinmap, pinOrFile string) {
	for pinFile := range pinsToApply {
		pins := pinsToApply[pinFile]
		for _, pin := range pins {
			beginIndex := strings.Index(pinFile, pinOrFile)
			if beginIndex == -1 {
				pinPath := fmt.Sprintf("%s/%s/%s", pin.GroupName, pin.Name, pin.Version)
				beginIndex = strings.Index(pinPath, pinOrFile)
				if beginIndex != -1 {
					ShowPinDiff(pin, pinFile)
				}
			} else {
				ShowPinDiff(pin, pinFile)
			}
		}
	}
}

func ShowAllDiffs(pinsToApply Pinmap, args cli.Args) {

	white := color.New(color.FgHiWhite).SprintFunc()
	fmt.Println(white("======================================================" +
		"======================================================" +
		"============================================================"))
	for pinFile := range pinsToApply {
		fmt.Printf(white("** %s\n"), pinFile)
		pins := pinsToApply[pinFile]
		for _, pin := range pins {
			ShowPinDiff(pin, "")
		}
		fmt.Println(white("======================================================" +
			"======================================================" +
			"============================================================"))
	}
}

func ShowPinDiff(pin Pin, filePath string) {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	white := color.New(color.FgHiWhite).SprintFunc()
	if pin.ApiSuccess() {
		fmt.Printf("%s\n", green(pin))
		if filePath != "" {
			fmt.Printf("%s\n", white(filePath))
		}
		fmt.Printf("   -- Ready to apply version: '%s'\n", pin.ApplyVersion)
		//fmt.Printf("      ==> %s\n", pin)
	} else {
		fmt.Printf("%s\n", red(pin))
		if filePath != "" {
			fmt.Printf("%s\n", white(filePath))
		}
		fmt.Printf("   -- Cannot apply version: '%s'\n", pin.ApplyVersion)
		//fmt.Printf("      ==> %s\n", pin)
	}

	fmt.Printf("      >> Api message ==> %s\n", pin.ApiMsg)

	if pin.IsPut() {
		pinContentFile, pinContent := ReadPinContentToPut(pin)
		fmt.Printf("      >> Content from: %s\n", pinContentFile)
		fmt.Print(pinContent)
	} else if pin.IsGet() {
		pinContentFile, pinContent := ReadPinContentToGet(pin)
		fmt.Printf("      >> Content from: %s\n", pinContentFile)
		fmt.Print(pinContent)
	}
}

func ShowConfig(args cli.Args) {
	// read in the config file and write it out to the console
	filename, _ := GetConfigFilename()
	fmt.Printf("Configuration in file: %s\n", filename)
	config, err := ioutil.ReadFile(filename)
	if err == nil {
		fmt.Printf("%s\n", config)
	} else {
		CmdsLogger.Debugf("Error reading file %s: %s", filename, err)
	}

	filename = GetHomeConfigFile()
	fmt.Printf("Configuration in file: %s\n", filename)
	config, err = ioutil.ReadFile(filename)
	if err == nil {
		fmt.Printf("%s\n", config)
	} else {
		CmdsLogger.Debugf("Error reading file %s: %s", filename, err)
	}
}

func CalculateHash(args cli.Args) {
	if len(args) > 0 {
		fileContents, err := ioutil.ReadFile(args[0])
		if err == nil {
			hash := Hash(string(fileContents))
			fmt.Printf("%s\n", hash)
		}
	}
}

func ApplyPermissions(args cli.Args) {
	CmdsLogger.Debugf("grant/remove/modify perms: %s", "need to implement")
}

func SendMessage(args cli.Args) {
	CmdsLogger.Debugf("send messages: %s", "need to implement")
}
