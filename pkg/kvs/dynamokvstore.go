package kvs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/kmgreen2/agglo/pkg/util"
	"strconv"
	"strings"
)

type DynamoKVStore struct {
	tableName string
	endpoint string
	region string
	prefixLength int
	client *dynamodb.DynamoDB
}

func keyPrefix(key string, length int) string {
	prefixLength := length
	if len(key) < length {
		prefixLength = len(key)
	}
	return key[:prefixLength]
}

// NewDynamoKVStoreFromConnectionString
// Format: endpoint=%s,region=%s,tableName=%s,prefixLength=%d
func NewDynamoKVStoreFromConnectionString(connectionString string) (*DynamoKVStore, error) {
	var err error
	var endpoint, region, tableName string
	var prefixLength int
	connectionStringAry := strings.Split(connectionString, ",")

	for _, entry := range connectionStringAry {
		entryAry := strings.Split(entry, "=")
		if len(entryAry) != 2 {
			return nil, util.NewInvalidError(fmt.Sprintf("invalid entry in connection string: %s", entry))
		}
		switch entryAry[0] {
		case "endpoint":
			endpoint = entryAry[1]
		case "region":
			region = entryAry[1]
		case "tableName":
			tableName = entryAry[1]
		case "prefixLength":
			prefixLength, err = strconv.Atoi(entryAry[1])
			if err != nil {
				return nil, err
			}
		}
	}

	missingEntries := ""

	if len(endpoint) == 0 {
		missingEntries += "endpoint "
	}
	if len(region) == 0 {
		missingEntries += "region "
	}
	if len(tableName) == 0 {
		missingEntries += "tableName "
	}
	if prefixLength == 0 {
		missingEntries += "prefixLength "
	}

	if len(missingEntries) > 0 {
		return nil, util.NewInvalidError(fmt.Sprintf("missing entries in connection string: %s", missingEntries))
	}

	return NewDynamoKVStore(endpoint, region, tableName, prefixLength), nil
}

func NewDynamoKVStore(endpoint, region, tableName string, prefixLength int) *DynamoKVStore {
	kvStore := &DynamoKVStore{
		tableName: tableName,
		endpoint: endpoint,
		region: region,
		prefixLength: prefixLength,
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config {
			Endpoint: &endpoint,
			Region: &region,
		},
	}))

	kvStore.client = dynamodb.New(sess)

	return kvStore
}

func (kvStore *DynamoKVStore) AtomicPut(ctx context.Context, key string, prev, value []byte) error {
	if prev == nil {
		return kvStore.Put(ctx, key, value)
	}
	expr, err := expression.NewBuilder().WithCondition(expression.Equal(expression.Name("Value"),
		expression.Value(prev))).Build()
	if err != nil {
		return err
	}
	if len(key) < kvStore.prefixLength {
		return util.NewInvalidError(fmt.Sprintf("key length must be >= prefix length: %d < %d",
			len(key), kvStore.prefixLength))
	}
	prefix := keyPrefix(key, kvStore.prefixLength)
	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			{
				Put: &dynamodb.Put{
					Item: map[string]*dynamodb.AttributeValue{
						"KeyPrefix": {
							S: &prefix,
						},
						"ValueKey": {
							S: &key,
						},
						"Value": {
							B: value,
						},
					},
					TableName:           &kvStore.tableName,
					ConditionExpression: expr.Condition(),
					ExpressionAttributeNames: expr.Names(),
					ExpressionAttributeValues: expr.Values(),
				},
			},
		},
	}

	_, err = kvStore.client.TransactWriteItems(input)
	if err != nil {
		return err
	}
	return nil
}

func (kvStore *DynamoKVStore)  AtomicDelete(ctx context.Context, key string, prev []byte) error {
	if prev == nil {
		return kvStore.Delete(ctx, key)
	}
	expr, err := expression.NewBuilder().WithCondition(expression.Equal(expression.Name("Value"),
		expression.Value(prev))).Build()
	if err != nil {
		return err
	}
	if len(key) < kvStore.prefixLength {
		return util.NewInvalidError(fmt.Sprintf("key length must be >= prefix length: %d < %d",
			len(key), kvStore.prefixLength))
	}
	prefix := keyPrefix(key, kvStore.prefixLength)
	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			{
				Delete: &dynamodb.Delete{
					Key: map[string]*dynamodb.AttributeValue{
						"KeyPrefix": {
							S: &prefix,
						},
						"ValueKey": {
							S: &key,
						},
					},
					TableName:           &kvStore.tableName,
					ConditionExpression: expr.Condition(),
					ExpressionAttributeNames: expr.Names(),
					ExpressionAttributeValues: expr.Values(),
				},
			},
		},
	}

	_, err = kvStore.client.TransactWriteItems(input)
	if err != nil {
		return err
	}
	return nil
}

func (kvStore *DynamoKVStore)  Put(ctx context.Context, key string, value []byte) error {
	if len(key) < kvStore.prefixLength {
		return util.NewInvalidError(fmt.Sprintf("key length must be >= prefix length: %d < %d",
			len(key), kvStore.prefixLength))
	}
	prefix := keyPrefix(key, kvStore.prefixLength)
	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"KeyPrefix": {
				S: &prefix,
			},
			"ValueKey": {
				S: &key,
			},
			"Value": {
				B: value,
			},
		},
		TableName: &kvStore.tableName,
	}
	_, err := kvStore.client.PutItem(input)
	if err != nil {
		return err
	}
	return nil

}

func (kvStore *DynamoKVStore) Get(ctx context.Context, key string) ([]byte, error) {
	if len(key) < kvStore.prefixLength {
		return nil, util.NewInvalidError(fmt.Sprintf("key length must be >= prefix length: %d < %d",
			len(key), kvStore.prefixLength))
	}
	expr, err := expression.NewBuilder().WithProjection(expression.AddNames(expression.NamesList(expression.Name(
		"Value")))).Build()
	if err != nil {
		return nil, err
	}
	prefix := keyPrefix(key, kvStore.prefixLength)
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"KeyPrefix": {
				S: &prefix,
			},
			"ValueKey": {
				S: &key,
			},
		},
		TableName: &kvStore.tableName,
		ProjectionExpression: expr.Projection(),
		ExpressionAttributeNames: expr.Names(),
	}
	result, err := kvStore.client.GetItem(input)
	if err != nil {
		return nil, err
	}

	if value, ok := result.Item["Value"]; ok {
		return value.B, nil
	}
	return nil, util.NewNotFoundError(fmt.Sprintf("cannot find item with key: %s", key))

}

func (kvStore *DynamoKVStore)  Head(ctx context.Context, key string) error {
	if len(key) < kvStore.prefixLength {
		return util.NewInvalidError(fmt.Sprintf("key length must be >= prefix length: %d < %d",
			len(key), kvStore.prefixLength))
	}
	expr, err := expression.NewBuilder().WithProjection(expression.AddNames(expression.NamesList(expression.Name(
		"Value")))).Build()
	if err != nil {
		return err
	}
	prefix := keyPrefix(key, kvStore.prefixLength)
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"KeyPrefix": {
				S: &prefix,
			},
			"ValueKey": {
				S: &key,
			},
		},
		TableName: &kvStore.tableName,
		ProjectionExpression: expr.Projection(),
		ExpressionAttributeNames: expr.Names(),
	}
	result, err := kvStore.client.GetItem(input)
	if err != nil {
		return err
	}

	if _, ok := result.Item["Value"]; ok {
		return nil
	}
	return util.NewNotFoundError(fmt.Sprintf("cannot find item with key: %s", key))
}

func (kvStore *DynamoKVStore)  Delete(ctx context.Context, key string) error {
	returnAllOld := "ALL_OLD"
	if len(key) < kvStore.prefixLength {
		return util.NewInvalidError(fmt.Sprintf("key length must be >= prefix length: %d < %d",
			len(key), kvStore.prefixLength))
	}
	prefix := keyPrefix(key, kvStore.prefixLength)
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"KeyPrefix": {
				S: &prefix,
			},
			"ValueKey": {
				S: &key,
			},
		},
		TableName: &kvStore.tableName,
		ReturnValues: &returnAllOld,
	}
	result, err := kvStore.client.DeleteItem(input)
	if err != nil {
		return err
	}

	if _, ok := result.Attributes["Value"]; ok {
		return nil
	}
	return util.NewNotFoundError(fmt.Sprintf("cannot find item with key: %s", key))
}

func (kvStore *DynamoKVStore) List(ctx context.Context, prefix string) ([]string, error) {
	var retItems []string
	var exclusiveStartKey map[string]*dynamodb.AttributeValue
	consistentRead := true
	if len(prefix) < kvStore.prefixLength {
		return nil, util.NewInvalidError(fmt.Sprintf("prefix length must be >= kvStore prefix length: %d < %d",
			len(prefix), kvStore.prefixLength))
	}
	lhs := expression.KeyEqual(expression.Key("KeyPrefix"),
		expression.Value(keyPrefix(prefix,kvStore.prefixLength)))
	rhs := expression.KeyBeginsWith(expression.Key("ValueKey"), prefix)

	expr, err := expression.NewBuilder().WithKeyCondition(expression.KeyAnd(lhs, rhs)).Build()
	if err != nil {
		return nil, err
	}

	for {
		input := &dynamodb.QueryInput{
			ConsistentRead: &consistentRead,
			TableName: &kvStore.tableName,
			ExpressionAttributeNames: expr.Names(),
			ExpressionAttributeValues: expr.Values(),
			KeyConditionExpression: expr.KeyCondition(),
			ExclusiveStartKey: exclusiveStartKey,
		}
		result, err := kvStore.client.Query(input)
		if err != nil {
			return nil, err
		}

		if len(result.Items) == 0 {
			break
		}
		for _, item := range result.Items {
			if key, ok := item["ValueKey"]; ok {
				retItems = append(retItems, key.GoString())
			}
		}
		if len(result.LastEvaluatedKey) == 0 {
			break
		}
		exclusiveStartKey = result.LastEvaluatedKey
	}
	return retItems, nil

}

func (kvStore *DynamoKVStore)  ConnectionString() string {
	return fmt.Sprintf("endpoint=%s,region=%s,tableName=%s,prefixLength=%d", kvStore.endpoint, kvStore.region,
		kvStore.tableName, kvStore.prefixLength)
}

func (kvStore *DynamoKVStore)  Close() error {
	return nil
}