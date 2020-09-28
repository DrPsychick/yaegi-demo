///usr/bin/env yaegi run "$0" "$@"; exit
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	var token = os.Getenv("GITHUB_TOKEN")
	var user = os.Getenv("GITHUB_USER")

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/users/%s/repos", user), nil)
	if err != nil {
		panic(err)
	}

	req.SetBasicAuth(token, "x-oauth-basic")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var prettyJson bytes.Buffer
	json.Indent(&prettyJson, body, "", "\t")
	fmt.Println(string(prettyJson.Bytes()))
}
