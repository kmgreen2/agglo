package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	var jsonBody map[string]interface{}
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(&jsonBody)
	if err != nil {
		fmt.Printf("Error decoding json: %s\n", err.Error())
	}
	fmt.Printf("Body: %v\n\n", jsonBody)
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

