package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/webhook", PrintBody)
	http.ListenAndServe(":80", nil)
}

func PrintBody(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Error reading the body: " + err.Error())
		}
		fmt.Println(string(bodyBytes))
	}
}

