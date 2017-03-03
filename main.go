package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"io/ioutil"
	"flag"
	"os"
	"os/exec"
	"sync"
	"strings"

	. "github.com/ivantjh/flow/constants"
	. "github.com/ivantjh/flow/models"
 	logger "github.com/ivantjh/flow/log"
)

var configData []Config
var secretKey string

func exec_cmd(cmd string, wg *sync.WaitGroup) {
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		errStr := fmt.Sprintf("SCRIPT: %v", err)
		logger.Log(errStr, ERROR)
	}

	wg.Done()
}

func process(dl DeployLog) {
	// Find location of repo
	var config Config
	for _, elem := range configData {
		if elem.RepoName == dl.RepoName {
			config = elem
			break;
		}
	}

	if config.RepoName == "" {
		logger.Log(fmt.Sprintf("No such repository found in config file: %s", dl.RepoName), ERROR)
	} else {

		// Go into dir of repo and run deploy.sh
		if err := os.Chdir(config.Location); err != nil {
			logger.Log(fmt.Sprintf("Invalid directory: %v", err), ERROR)
		}

		cmds := [2]string{"chmod +x deploy.sh",
											"/bin/sh deploy.sh"}

		var wg sync.WaitGroup
		wg.Add(2)

		for _, cmd := range cmds {
			exec_cmd(cmd, &wg)
		}

		wg.Wait()
		logger.Log(fmt.Sprintf("Deployed %s", dl.RepoName), INFO)
	}
}

func handler(rw http.ResponseWriter, req *http.Request) {
	if event, ok := req.Header["X-Github-Event"]; ok {
		if event[0] == "push" {

			body, err := ioutil.ReadAll(req.Body)

			if secretKey != "" {
				sig := req.Header.Get("X-Hub-Signature")

				if sig == "" {
					http.Error(rw, "403 Forbidden - Missing X-Hub-Signature required for HMAC verification", http.StatusForbidden)
					return
				}

				mac := hmac.New(sha1.New, []byte(secretKey))
				mac.Write(body)
				expectedMAC := mac.Sum(nil)
				expectedSig := "sha1=" + hex.EncodeToString(expectedMAC)
				if !hmac.Equal([]byte(expectedSig), []byte(sig)) {
					http.Error(rw, "403 Forbidden - HMAC verification failed", http.StatusForbidden)
					return
				}
			}

			var data map[string]interface{}
			if err = json.Unmarshal(body, &data); err != nil {
				logger.Log(fmt.Sprintf("Invalid webhook payload %s", err), ERROR)
			}

			repo := data["repository"].(map[string]interface{})
			headCommit := data["head_commit"].(map[string]interface{})

			var dl DeployLog
			dl.Id = repo["id"].(float64)
			dl.RepoName = repo["name"].(string)
			dl.TimeStamp = headCommit["timestamp"].(string)

			process(dl)
			rw.Header().Set("Content-Type", "text/plain")
			rw.Write([]byte("Processing commit to master\n"))
			return
		}
	}

	http.Error(rw, "500 Bad Request", http.StatusBadRequest)
}

func parseConfig(fileLocation string) {
	file, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		fmt.Printf("Unable to read %s\n%v\n", fileLocation, err)
		os.Exit(1)
	}

	err = json.Unmarshal(file, &configData)
	if err != nil {
		fmt.Printf("Invalid json from %s\n%v\n", fileLocation, err)
		os.Exit(1)
	}
}

func parseFlags() (configLocation string) {
	logsLocaPtr := flag.String("logs", "/var/log/", "Location of flow logs")
	configLocaPtr := flag.String("config", "", "Location of config file")
	secretKeyPtr := flag.String("secret", "", "Secret key, needs to be added on github too")

	flag.Usage = func() {
		fmt.Println("Flow runs a deploy.sh script once commits on master are detected.")
		fmt.Println("Add locations of repositories to be tracked in config.json.")
		fmt.Println("Usage:")
		fmt.Println("flow -config config.json\n")

		fmt.Println("Parameters")
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(*configLocaPtr) == 0 {
		fmt.Println("No config file specified")
		os.Exit(1)
	}

	if *secretKeyPtr != "" {
		secretKey = *secretKeyPtr
	}

	logger.LogsPath = fmt.Sprintf("%sflow.log", *logsLocaPtr)
	return *configLocaPtr
}

func main() {
	fileLocation := parseFlags()
	parseConfig(fileLocation)

	fmt.Println("Starting server on port 8080")
	http.HandleFunc("/payload", handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("%v", err)
	}
}
