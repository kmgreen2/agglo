package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/core"
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

	transformer := core.NewTransformer(nil, ".", ":")
	transformer.AddSpec("foo.bar", core.NewTransformation("a", []core.FieldTransformer{&core.CopyTransformer{}}, nil))
	transformer.AddSpec("foo.bizz.bar", core.NewTransformation("b.c",
		[]core.FieldTransformer{&core.MapTransformer{mapFunc}}, nil))
	cond, err := core.NewCondition(core.NewComparatorExpression(core.Variable("b.d.0"), core.Numeric(3), core.Equal))
	if err != nil {
		panic(err.Error())
	}
	transformer.AddSpec("foo.baz", core.NewTransformation("b.d",
		[]core.FieldTransformer{&core.MapTransformer{intMapFunc}, &core.LeftFoldTransformer{foldFunc}}, cond))

	transformedMap, err := transformer.Transform(jsonMap)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(transformedMap)

	fmt.Println(core.Flatten(jsonMap))

}

