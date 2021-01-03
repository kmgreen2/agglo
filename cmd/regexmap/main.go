package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/core"
	"io/ioutil"
	"os"
	"regexp"
)

func main() {
	var inMap, outMap map[string]interface{}
	var keys []string
	var regex, replace string

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

	outMap = core.CopyableMap(inMap).DeepCopy()

	if len(os.Args) < 4 {
		panic(fmt.Sprintf("expected at least 4 args: <regex> <replace> [k1, k2, ...], got %d", len(os.Args)))
	}

	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		if i == 1 {
			regex = arg
			continue
		}
		if i == 2 {
			replace = arg
			continue
		}
		keys = append(keys, arg)
	}

	re, err := regexp.Compile(regex)
	if err != nil {
		panic(err)
	}

	for _, key := range keys {
		if _, keyOk := inMap[key]; keyOk {
			if val, ok := inMap[key].(string); ok {
				if re.Match([]byte(val)) {
					outMap[key] = replace
				}
			}
		}
	}

	encodeBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(encodeBuffer)
	err = encoder.Encode(outMap)
	if err != nil {
		panic(err)
	}
	fmt.Print(encodeBuffer.String())
}
