package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"io/ioutil"
	"flag"
	"os"
	"os/exec"
	"log"
	"sync"
	"strings"

	. "github.com/ivantjh/flow/models"
)

var configData []Config

func exec_cmd(cmd string, wg *sync.WaitGroup) {
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	_, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Printf("%s", err)
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
		log.Printf("No such repository found in config file: %s", dl.RepoName)
	} else {

		// Go into dir of repo, pull and run deploy.sh
		if err := os.Chdir(config.Location); err != nil {
			log.Printf("Invalid directory: %v", err)
		}

		cmds := [2]string{"chmod +x deploy.sh",
											"/bin/sh deploy.sh"}

		var wg sync.WaitGroup
		wg.Add(2)

		for _, cmd := range cmds {
			exec_cmd(cmd, &wg)
		}

		wg.Wait()
		log.Println("Finish executing")
	}
}

func handler(rw http.ResponseWriter, req *http.Request) {
	// fmt.Println(req.Header)

	// only handle push payload
	body, err := ioutil.ReadAll(req.Body)

	var data map[string]interface{}
	if err = json.Unmarshal(body, &data); err != nil {
		log.Printf("Invalid webhook payload %v", err)
	}

	repo := data["repository"].(map[string]interface{})
	headCommit := data["head_commit"].(map[string]interface{})

	var dl DeployLog
	dl.Id = repo["id"].(float64)
	dl.RepoName = repo["name"].(string)
	dl.TimeStamp = headCommit["timestamp"].(string)

	process(dl)
}

func readfile(fileLocation string) {
	file, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		log.Printf("Unable to read %s\n%v\n", fileLocation, err)
		os.Exit(1)
	}

	err = json.Unmarshal(file, &configData)
	if err != nil {
		log.Printf("Invalid json from %s\n%v\n", fileLocation, err)
		os.Exit(1)
	}
}

func parseFlags() (fileLocation string) {
	flag.Usage = func() {
		fmt.Println("Flow runs a deploy.sh script once commits on master are detected.")
		fmt.Println("Add locations of repositories to be tracked.")
		fmt.Println("Usage of Flow:")
		fmt.Println("flow config.json")
	}

	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("No config file specified")
		os.Exit(1)
	}

	return flag.Args()[0]
}

func main() {
	fileLocation := parseFlags()
	readfile(fileLocation)


	fmt.Println("Starting server")
	http.HandleFunc("/payload", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
