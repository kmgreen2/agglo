package search_test

import (
	"context"
	"github.com/kmgreen2/agglo/pkg/search"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestElasticSearchConnectionStringHappyPath(t *testing.T) {
	_, err := search.NewIndexFromConnectionString("elastic:index=test-index,nodes=localhost:9200,authString=basic;;")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	_, err = search.NewIndexFromConnectionString("elastic:bulk,bulkThreads=4,bulkThresholdBytes=500000,index=test-index,nodes=localhost:9200,authString=basic;;")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}

func TestElasticSearchConnectionStringInvalid(t *testing.T) {
	_, err := search.NewElasticsearchIndexFromConnectionString("=test-index,nodes=localhost:9200,authString=basic;;")
	assert.Error(t, err)
	_, err = search.NewElasticsearchIndexFromConnectionString("index,nodes=localhost:9200,authString=basic;;")
	assert.Error(t, err)
	_, err = search.NewElasticsearchIndexFromConnectionString("bulk,bulkThreads=4,bulkThresholdBytes=500000,index=test-index,nodes=localhost:9200,authString=foo;bar")
	assert.Error(t, err)
}

func TestElasticSearchSinglePutQueryHappyPath(t *testing.T) {
	docId := "1234"
	esIndex, err := search.NewElasticsearchIndexFromConnectionString("index=test-index,nodes=http://localhost:9200,authString=basic;;")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer func() {
		 _ = esIndex.Delete(context.Background(), docId)
	}()

	builder := search.NewElasticIndexValueBuilder(docId)

	value := builder.AddKeyword("foo", "bar").
			AddNumeric("fizz", 56).
			SetBlob([]byte("blob of stuff")).
			Get()

	err = esIndex.Put(context.Background(), value.Id(), value)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	queryBuilder := search.NewSimpleElasticQueryBuilder()

	query := queryBuilder.AddTerm("foo", "bar").Build()

	results, err := esIndex.Query(context.Background(), query)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	elasticResults, err := search.ResolveElasticResults(results)
	_ = elasticResults
	if hitsResults, ok := results["hits"].(map[string]interface{}); ok {
		if hitsList, ok2 := hitsResults["hits"].([]interface{}); ok2 {
			assert.Equal(t, len(hitsList), 1)
		} else {
			assert.FailNow(t, "Cannot find inner 'hits' entry in results map")
		}
	} else {
		assert.FailNow(t, "Cannot find outer 'hits' entry in results map")
	}
}

func TestElasticSearchSinglePutBadPayload(t *testing.T) {
}

func TestElasticSearchBulkPutQueryHappyPath(t *testing.T) {
}

func TestElasticSearchBulkPutBadPayload(t *testing.T) {
}

func TestElasticSearchQueryBadPayload(t *testing.T) {
}

func TestElasticSearchDeleteHappyPath(t *testing.T) {
}

func TestElasticSearchDeleteInvalid(t *testing.T) {
}
