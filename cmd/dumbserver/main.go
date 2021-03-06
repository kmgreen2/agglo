package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

type DumbServerArgs struct {
	port int
}

func parseArgs() *DumbServerArgs {
	args := &DumbServerArgs{}

	portPtr := flag.Int("port", 8080, "listening port")
	flag.Parse()

	args.port = *portPtr

	return args
}

func main() {
	args := parseArgs()
	http.HandleFunc("/webhook", PrintBody)
	err := http.ListenAndServe(fmt.Sprintf(":%d", args.port), nil)
	if err != nil {
		panic(err)
	}
}

func PrintBody(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Error reading the body: " + err.Error())
		}
		fmt.Println(string(bodyBytes))
		_, _ = w.Write(bodyBytes)
	}
}

