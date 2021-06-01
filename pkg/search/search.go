package search

import (
	"context"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
	"strings"
)

type IndexResult struct {}

type ItemIndexType int

const (
	IndexItemNumeric ItemIndexType = 0x01
	IndexItemKeyword = 0x02
	IndexItemFreeText = 0x04
	IndexItemDate = 0x08
	IndexItemBlob = 0x10
)

type IndexItem struct {
	item interface{}
	itemType ItemIndexType
}

type IndexValue interface {
	Values() map[string]*IndexItem
	Id() string
}

type QuerySpec interface {
	Render()  ([]byte, error)
}

type Index interface {
	Put(ctx context.Context, documentID string, value IndexValue) error
	Delete(ctx context.Context, documentID string) error
	Query(ctx context.Context, query QuerySpec) (map[string]interface{}, error)
	Close(ctx context.Context) error
}

func NewIndexFromConnectionString(connectionString string) (Index, error) {
	connectionStringAry := strings.Split(connectionString, ":")
	if len(connectionStringAry) < 2 {
		return nil, util.NewInvalidError(fmt.Sprintf("invalid connection string, expected <type>:<connStr> got: %s",
			connectionString))
	}
	switch connectionStringAry[0] {
	case "elastic":
		return NewElasticsearchIndexFromConnectionString(strings.Join(connectionStringAry[1:], ":"))
	}
	return nil, util.NewInvalidError(fmt.Sprintf("invalid index type: %s", connectionStringAry[0]))
}


