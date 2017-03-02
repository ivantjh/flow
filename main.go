package main

import (
	"encoding/json"
	"fmt"
	// "strconv"
	"net/http"
	"io/ioutil"
	"flag"
	"os"
	"log"

	. "github.com/ivantjh/flow/models"
)

var configData []Config

func handler(rw http.ResponseWriter, req *http.Request) {
	// fmt.Println(req.Header)
	body, err := ioutil.ReadAll(req.Body)

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	repo := data["repository"].(map[string]interface{})
	headCommit := data["head_commit"].(map[string]interface{})

	var dl DeployLog
	dl.Id = (repo["id"].(float64))
	dl.RepoName = repo["name"].(string)
	dl.TimeStamp = headCommit["timestamp"].(string)

	// fmt.Println(repo)


	// dl := DeployLog{repo["id"], repo.(string)["name"], headCommit.(string)["timeStamp"]}
	fmt.Println(dl)
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

	fmt.Println(configData)
}

func parseFlags() (fileLocation string){
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


	// fmt.Println("Starting server")
	// http.HandleFunc("/payload", handler)
	// http.ListenAndServe(":8080", nil)

}
