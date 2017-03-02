package main

import (
	"encoding/json"
	"fmt"
	// "strconv"
	"net/http"
	"io/ioutil"
)

type DeployLog struct {
	id float64;
	repoName string;
	timeStamp string;
}

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
	dl.id = (repo["id"].(float64))
	dl.repoName = repo["name"].(string)
	dl.timeStamp = headCommit["timestamp"].(string)

	// fmt.Println(repo)


	// dl := DeployLog{repo["id"], repo.(string)["name"], headCommit.(string)["timeStamp"]}
	fmt.Println(dl)
}

func main() {
	fmt.Println("Starting server")
	http.HandleFunc("/payload", handler)
	http.ListenAndServe(":8080", nil)

}
