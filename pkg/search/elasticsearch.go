package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/cenkalti/backoff/v4"
	"strconv"
	"strings"
	"time"
)

type ElasticsearchIndex struct {
	client *elasticsearch.Client
	index string
	useBulkIndexer bool
	bulkIndexer esutil.BulkIndexer
	bulkThreads int
	bulkThresholdBytes int
	connectionString string
}

func NewElasticsearchIndexFromConnectionString(connectionString string) (*ElasticsearchIndex, error) {
	var err error
	esIndex := &ElasticsearchIndex{}
	connectionStringAry := strings.Split(connectionString, ",")
	if len(connectionStringAry) < 3 {
		return nil, util.NewInvalidError(
			fmt.Sprintf("invalid connection string: %s", connectionString))
	}

	// ToDo(KMG): Can make this configurable
	retryBackoff := backoff.NewExponentialBackOff()

	config := &elasticsearch.Config{
		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},
		MaxRetries: 5,
	}

	// Set connection string
	esIndex.connectionString = connectionString

	// Set defaults
	esIndex.bulkThreads = 4
	esIndex.bulkThresholdBytes = 5e+6

	hasAuth := false;

	for _, entry := range connectionStringAry {
		entryAry := strings.Split(entry, "=")
		if len(entryAry) != 2 && (len(entryAry) == 1 && strings.Compare(entryAry[0], "bulk") != 0) {
			return nil, util.NewInvalidError(fmt.Sprintf("invalid entry in connection string: %s", entry))
		}
		switch entryAry[0] {
		case "bulk":
			esIndex.useBulkIndexer = true
		case "bulkThreads":
			esIndex.bulkThreads, err = strconv.Atoi(entryAry[1])
			if err != nil {
				return nil, err
			}
		case "bulkThresholdBytes":
			esIndex.bulkThresholdBytes, err = strconv.Atoi(entryAry[1])
			if err != nil {
				return nil, err
			}
		case "index":
			esIndex.index = entryAry[1]
		case "nodes":
			config.Addresses = strings.Split(entryAry[1], "=")
		case "authString":
			authStringAry := strings.Split(entryAry[1], ";")
			if len(authStringAry) != 3 {
				return nil, util.NewInvalidError(fmt.Sprintf("invalid authString in connection string: %s", entryAry[1]))
			}
			if strings.Compare(authStringAry[0], "basic") == 0 {
				config.Username = authStringAry[1]
				config.Password = authStringAry[2]
			} else if strings.Compare(authStringAry[0], "cloud") == 0 {
				config.CloudID = authStringAry[1]
				config.APIKey = authStringAry[2]
			} else {
				return nil, util.NewInvalidError(fmt.Sprintf("invalid authType in connection string: %s", authStringAry[0]))
			}
			hasAuth = true
		}
	}

	if esIndex.index == "" {
		return nil, util.NewInvalidError(fmt.Sprintf("Must specify an index"))
	}

	if len(config.Addresses) == 0 {
		return nil, util.NewInvalidError(fmt.Sprintf("Must specify hosts"))
	}

	if !hasAuth {
		return nil, util.NewInvalidError(fmt.Sprintf("Must specify auth string"))
	}

	esIndex.client, err = elasticsearch.NewClient(*config)
	if err != nil {
		return nil, err
	}

	if esIndex.useBulkIndexer {
		esIndex.bulkIndexer, err = esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
			Index: esIndex.index,
			Client: esIndex.client,
			NumWorkers: esIndex.bulkThreads,
			FlushBytes: esIndex.bulkThresholdBytes,
			FlushInterval: 30 * time.Second,
		})
		if err != nil {
			return nil, err
		}
	}

	return esIndex, nil
}

const (
	ElasticNumeric string = "numeric"
	ElasticKeyword = "keywords"
	ElasticFreeText = "text"
	ElasticDate = "date"
	ElasticCreated = "created"
	ElasticBlob = "blob"
)

func toPayload(value IndexValue) ([]byte, error) {
	m := make(map[string][]interface{})
	m[ElasticNumeric] = make([]interface{}, 0)
	m[ElasticKeyword] = make([]interface{}, 0)
	m[ElasticFreeText] = make([]interface{}, 0)
	m[ElasticDate] = make([]interface{}, 0)
	m[ElasticBlob] = make([]interface{}, 0)

	for k, v := range value.Values() {
		switch v.itemType {
		case IndexItemNumeric:
			m[ElasticNumeric] = append(m[ElasticNumeric], map[string]interface{}{
				"id": k,
				"value": v.item,
			})
		case IndexItemFreeText:
			m[ElasticFreeText] = append(m[ElasticFreeText], map[string]interface{}{
				"id": k,
				"value": v.item,
			})
		case IndexItemKeyword:
			m[ElasticKeyword] = append(m[ElasticKeyword], map[string]interface{}{
				"id": k,
				"value": v.item,
			})
		case IndexItemDate:
			m[ElasticDate] = append(m[ElasticDate], map[string]interface{}{
				"id": k,
				"value": v.item,
			})
		case IndexItemCreated:
			if len(m[ElasticCreated]) == 0 {
				m[ElasticCreated] = append(m[ElasticCreated], v.item)
			} else {
				m[ElasticCreated][0] = v.item
			}
		case IndexItemBlob:
			if len(m[ElasticBlob]) == 0 {
				m[ElasticBlob] = append(m[ElasticBlob], v.item)
			} else {
				m[ElasticBlob][0] = v.item
			}
		}
	}

	jsonBytes := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(jsonBytes).Encode(m)
	if err != nil {
		return nil, err
	}
	return jsonBytes.Bytes(), nil
}

func (elastic *ElasticsearchIndex) singlePut(ctx context.Context, documentID string, value IndexValue) error {
	bodyBytes, err := toPayload(value)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index: elastic.index,
		DocumentID: documentID,
		Body: bytes.NewBuffer(bodyBytes),
		Refresh: "true",
	}

	res, err := req.Do(ctx, elastic.client)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.IsError() {
		return util.NewIndexError(res.Status())
	}
	return nil
}

func (elastic *ElasticsearchIndex) bulkPut(ctx context.Context, documentID string, value IndexValue) error {
	bodyBytes, err := toPayload(value)
	if err != nil {
		return err
	}

	return elastic.bulkIndexer.Add(ctx, esutil.BulkIndexerItem{
		Action: "index",
		DocumentID: documentID,
		Body: bytes.NewBuffer(bodyBytes),
	})
}

func (elastic *ElasticsearchIndex) Put(ctx context.Context, documentID string, value IndexValue) error {
	if elastic.useBulkIndexer == true {

	} else {
		return elastic.singlePut(ctx, documentID, value)
	}
	return nil
}

func (elastic *ElasticsearchIndex) Query(ctx context.Context, query QuerySpec) (map[string]interface{}, error) {
	var results map[string]interface{}

	queryBodyBytes, err := query.Render()
	if err != nil {
		return nil, err
	}

	req := esapi.SearchRequest{
		Index: []string{elastic.index},
		Body: bytes.NewBuffer(queryBodyBytes),
	}

	res, err := req.Do(ctx, elastic.client)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.IsError() {
		return nil, util.NewIndexError(res.Status())
	}

	err = json.NewDecoder(res.Body).Decode(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (elastic *ElasticsearchIndex) Delete(ctx context.Context, documentID string) error {
	req := esapi.DeleteRequest {
		Index: elastic.index,
		DocumentID: documentID,
	}

	res, err := req.Do(ctx, elastic.client)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.IsError() {
		return util.NewIndexError(res.Status())
	}
	return nil
}

func (elastic *ElasticsearchIndex) Close(ctx context.Context) error {
	if elastic.useBulkIndexer {
		return elastic.bulkIndexer.Close(ctx)
	}
	return nil
}

func (elastic *ElasticsearchIndex) ConnectionString() string {
	return elastic.connectionString
}
