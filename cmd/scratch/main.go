package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/core"
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
	//http.HandleFunc("/", handler)
	//log.Fatal(http.ListenAndServe(":8080", nil))
	testJson := `
{
	"a": 1,
	"b": {
		"c": 2,
		"d": [3,4,5]
	},
	"e": [6],
	"f": {
		"g": {
			"h": 7
		}
	}
}
`

	var jsonMap map[string]interface{}
	decoder := json.NewDecoder(bytes.NewBuffer([]byte(testJson)))
	err := decoder.Decode(&jsonMap)
	if err != nil {
		panic(err.Error())
	}

	foldFunc := func(acc, v interface{}) (interface{}, error) {
		if acc == nil {
			acc = 0
		}

		if accVal, err := core.GetNumeric(acc); err != nil {
			return nil, err
		} else if vVal, err := core.GetNumeric(v); err != nil {
			return nil, err
		} else {
			return accVal + vVal, nil

		}
	}

	transformer := core.NewTransformer(nil, ".", ":")
	transformer.AddSpec("foo.bar", core.NewTransformation("a", core.CopyTransformer{}))
	transformer.AddSpec("foo.baz", core.NewTransformation("b.d", core.LeftFoldTransformer{foldFunc}))

	transformedMap, err := transformer.Transform(jsonMap)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(transformedMap)

}

