package process

import (
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestPipelinesBasic(t *testing.T) {
	tmpFile := "/tmp/testPipelinesBasic"
	_ = os.Remove(tmpFile)
	fp, err := os.Open("../../../test/config/basic_pipeline.json")
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
	assert.Equal(t, 1, len(pipelines.Underlying()))


	testMaps := test.PipelineTestMapsOne()

	for _, m := range testMaps {
		out, err := pipelines.Underlying()[0].RunSync(m)
		if _, ok := out["author"]; ok {
			assert.Equal(t, "git-dev", out["version-control"].(string))
		}
		if deployer, ok := out["deploy"]; ok {
			assert.Equal(t, "circleci-dev", deployer.(string))
			assert.Equal(t, "deploybot", out["user"].(string))
		}
		assert.Nil(t, err)
	}

	file, err := os.OpenFile(tmpFile, os.O_RDONLY, 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer func() {
		_ = file.Close()
	}()

	result, err := ioutil.ReadAll(file)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, "abcd\nbeef\n", (string(result)))

	_ = os.Remove(tmpFile)
}
