package serialization

import (
	"fmt"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestPipelinesBasic(t *testing.T) {
	fp, err := os.Open("../../test/config/basic_pipeline.json")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	configBytes, err := ioutil.ReadAll(fp)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	pipelines, err := PipelinesFromJson(configBytes)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.NotNil(t, pipelines)
	assert.Equal(t, 1, len(pipelines))

	testMaps := test.PipelineTestMapsOne()

	for _, m := range testMaps {
		out, err := pipelines[0].RunSync(m)
		fmt.Println(out)
		assert.Nil(t, err)
	}
}