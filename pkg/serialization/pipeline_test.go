package serialization

import (
	"bytes"
	"github.com/golang/protobuf/jsonpb"
	pipelineapi "github.com/kmgreen2/agglo/pkg/proto"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestPipelinesBasic(t *testing.T) {
	var pipelinesPb pipelineapi.Pipelines
	fp, err := os.Open("test/config/basic_pipeline.json")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	configBytes, err := ioutil.ReadAll(fp)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	byteBuffer := bytes.NewBuffer(configBytes)
	err = jsonpb.Unmarshal(byteBuffer, &pipelinesPb)
}