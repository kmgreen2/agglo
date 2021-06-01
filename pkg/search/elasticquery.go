package search

import (
	"bytes"
	"encoding/json"
	"time"
)

type ElasticQuery interface {
	Render() ([]byte, error)
}

type ElasticBuilder interface {
	Build() ElasticQuery
}

type SimpleElasticQuery struct {
	text string
	keywords map[string]string
	startTime int64
	endTime int64
	sortBy []string
}

func (query *SimpleElasticQuery) Render() ([]byte, error) {
	var terms []map[string]interface{}
	var timeRange map[string]interface{}

	for k, v := range(query.keywords) {
		terms = append(terms, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]map[string]string{
					{"term": {"keywords.id": k}},
					{"term": {"keywords.value": v}},
				},
				"minimum_should_match": 2,
			},
		})
	}

	if query.startTime > 0 {
		t := time.Unix(query.startTime, 0).UTC()
		timeRange = map[string]interface{}{
			"created": map[string]string{
				"gte": t.Format("2006-01-02 15:04:05"),
			},
		}
	}

	if query.endTime > 0 {
		t := time.Unix(query.endTime, 0).UTC()
		if timeDict, timeDictOk := timeRange["created"]; timeDictOk {
			if createdEntry, createdOk := timeDict.(map[string]string); createdOk {
				createdEntry["lte"] = t.Format("2006-01-02 15:04:05")
			}
		} else {
			timeRange = map[string]interface{}{
				"created": map[string]string{
					"lte": t.Format("2006-01-02 15:04:05"),
				},
			}
		}
	}

	queryMap := map[string]interface{}{
		"query": map[string]map[string]interface{}{
			"bool": make(map[string]interface{}),
		},
	}

	if q, okQ := queryMap["query"]; okQ {
		if qMap, okQMap := q.(map[string]map[string]interface{}); okQMap {
			if b, okB := qMap["bool"]; okB {
				if len(terms) > 0 {
					b["must"] = map[string]map[string]interface{}{
						"nested" :{
							"path": "keywords",
							"query": terms,
						},
					}
				}
				if len(timeRange) > 0 {
					b["filter"] = map[string]interface{}{
						"range": timeRange,
					}
				}
				if len(query.text) > 0 {
					b["should"] = map[string]interface{}{
						"match": map[string]interface{}{
							"fulltext": map[string]string{
								"query": query.text,
							},
						},
					}
				}
			}
		}
	}

	queryBytes := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(queryBytes)
	err := encoder.Encode(queryMap)

	if err != nil {
		return nil, err
	}

	return queryBytes.Bytes(), nil
}

type SimpleElasticQueryBuilder struct {
	query *SimpleElasticQuery
}

func NewSimpleElasticQueryBuilder() (builder *SimpleElasticQueryBuilder) {
	return &SimpleElasticQueryBuilder{
		query: &SimpleElasticQuery{
			keywords: make(map[string]string),
		},
	}
}

func (builder *SimpleElasticQueryBuilder) AddTerm(key, value string) *SimpleElasticQueryBuilder {
	builder.query.keywords[key] = value
	return builder
}

func (builder *SimpleElasticQueryBuilder) AddText(value string) *SimpleElasticQueryBuilder {
	if len(builder.query.text) > 0 {
		builder.query.text += " " + value
	} else {
		builder.query.text = value
	}
	return builder
}

func (builder *SimpleElasticQueryBuilder) SetStartTime(ts int64) *SimpleElasticQueryBuilder {
	builder.query.startTime = ts
	return builder
}

func (builder *SimpleElasticQueryBuilder) SetEndTime(ts int64) *SimpleElasticQueryBuilder {
	builder.query.endTime = ts
	return builder
}

func (builder *SimpleElasticQueryBuilder) Build() ElasticQuery {
	return builder.query
}


