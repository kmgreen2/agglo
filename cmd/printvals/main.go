package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	var inMap map[string]interface{}
	var keys []string
	var outfile string

	outStr := ""

	inBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	decodeBuffer := bytes.NewBuffer(inBytes)
	decoder := json.NewDecoder(decodeBuffer)
	err = decoder.Decode(&inMap)
	if err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		panic(fmt.Sprintf("expected at least 2 args: <outfile> [k1, k2, ...], got %d", len(os.Args)))
	}

	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		if i == 1 {
			outfile = arg
			continue
		}
		keys = append(keys, arg)
	}

	for _, key := range keys {
		if _, inOk := inMap[key]; inOk {
			outStr += fmt.Sprintf("%s\n", inMap[key])
		}
	}

	fp, err := os.OpenFile(outfile, os.O_APPEND | os.O_CREATE | os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	_, err = io.WriteString(fp, outStr)
	if err != nil {
		panic(err)
	}
}
