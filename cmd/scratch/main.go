package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/core/pipeline"
	"net/http"
	"strings"
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
		"c": "hello",
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

	intMapFunc := func(v interface{}) (interface{}, error) {
		vVal, err := core.GetNumeric(v)
		if err != nil {
			return nil, err
		}
		return vVal * 2, nil
	}

	mapFunc := func(v interface{}) (interface{}, error) {
		return strings.ToUpper(v.(string)), nil
	}

	transformer := pipeline.NewTransformer(nil, ".", ":")
	transformer.AddSpec("a", "foo.bar", core.NewTransformation([]core.FieldTransformation{&core.CopyTransformation{}},
	nil))
	transformer.AddSpec("b.c", "foo.bizz.bar", core.NewTransformation(
		[]core.FieldTransformation{&core.MapTransformation{mapFunc}}, nil))
	cond, err := core.NewCondition(core.NewComparatorExpression(core.Variable("b.d.0"), 2, core.Equal))
	if err != nil {
		panic(err.Error())
	}
	transformer.AddSpec("b.d", "foo.baz", core.NewTransformation(
		[]core.FieldTransformation{&core.MapTransformation{intMapFunc}, &core.LeftFoldTransformation{foldFunc}}, cond))

	transformedMap, err := transformer.Process(jsonMap)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(transformedMap)

	fmt.Println(core.Flatten(jsonMap))

}

