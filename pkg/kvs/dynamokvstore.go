package kvs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/kmgreen2/agglo/pkg/util"
)

type DynamoKVStore struct {
	tableName string
	endpoint string
	region string
	client *dynamodb.DynamoDB
}

func NewDynamoKVStore(endpoint, region, tableName string) *DynamoKVStore {
	kvStore := &DynamoKVStore{
		tableName: tableName,
		endpoint: endpoint,
		region: region,
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
	expr, err := expression.NewBuilder().WithCondition(expression.Equal(expression.Name("Value"),
		expression.Value(prev))).Build()
	if err != nil {
		return err
	}
	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			{
				Put: &dynamodb.Put{
					Item: map[string]*dynamodb.AttributeValue{
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
	expr, err := expression.NewBuilder().WithCondition(expression.Equal(expression.Name("Value"),
		expression.Value(prev))).Build()
	if err != nil {
		return err
	}
	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			{
				Delete: &dynamodb.Delete{
					Key: map[string]*dynamodb.AttributeValue{
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
	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
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
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ValueKey": {
				S: &key,
			},
		},
		TableName: &kvStore.tableName,
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
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ValueKey": {
				S: &key,
			},
		},
		TableName: &kvStore.tableName,
	}
	_, err := kvStore.client.GetItem(input)
	if err != nil {
		return err
	}
	return nil
}

func (kvStore *DynamoKVStore)  Delete(ctx context.Context, key string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ValueKey": {
				S: &key,
			},
		},
		TableName: &kvStore.tableName,
	}
	_, err := kvStore.client.DeleteItem(input)
	if err != nil {
		return err
	}
	return nil
}

func (kvStore *DynamoKVStore) List(ctx context.Context, prefix string) ([]string, error) {
	var retItems []string
	var exclusiveStartKey map[string]*dynamodb.AttributeValue
	consistentRead := true
	expr, err := expression.NewBuilder().WithKeyCondition(expression.KeyBeginsWith(expression.Key("ValueKey"),
		prefix)).Build()
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
	return kvStore.tableName

}

func (kvStore *DynamoKVStore)  Close() error {
	return nil
}