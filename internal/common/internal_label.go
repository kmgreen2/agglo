package common

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
	"strings"
)

type InternalKey string
type InternalKeyPrefix string

const (
	PartitionIDKey InternalKey = "internal:partitionID"
	ResourceNameKey InternalKey = "internal:name"
	AccumulatorKey InternalKey = "internal:acc"
	CheckpointIndexKey InternalKey = "internal:checkpoint:idx"
	CheckpointDataKey InternalKey = "internal:checkpoint:data"
	TeeMetadataKey InternalKey = "internal:tee:output"
	SpawnMetadataKey InternalKey = "internal:spawn:output"
	EntwineMetadataKey InternalKey = "internal:entwine:output"
	MessageIDKey InternalKey = "internal:messageID"
	NTPTimeKey InternalKey = "internal:ntp:time"
)

func (k InternalKey) String()string {
	return string(k)
}

const (
	AggregationDataPrefix  InternalKeyPrefix = "internal:aggregation"
	CompletionStatusPrefix InternalKeyPrefix = "internal:completion"
	CompletionStatePrefix  InternalKeyPrefix = "internal:completion:state"
)

func (k InternalKeyPrefix) String()string {
	return string(k)
}
var reservedKeys = []fmt.Stringer{PartitionIDKey, ResourceNameKey, AccumulatorKey, CheckpointIndexKey,
	CheckpointDataKey, TeeMetadataKey, AggregationDataPrefix, CompletionStatusPrefix, MessageIDKey}

func GetReservedKeys()[]fmt.Stringer {
	return reservedKeys
}

func IsReservedKey(key string) bool {
	for _, s := range reservedKeys {
		if strings.Contains(key, s.String()) {
			return true
		}
	}
	return false
}

func ContainsReservedKey(in map[string]interface{}) bool {
	for k, _ := range in {
		if IsReservedKey(k) {
			return true
		}
	}
	return false
}

func GetFromInternalKey(key InternalKey, in map[string]interface{}) (interface{}, bool) {
	value, ok := in[string(key)]
	return value, ok
}

func MustSetUsingInternalKey(key InternalKey, value interface{}, in map[string]interface{}) {
	in[string(key)] = value
}

func SetUsingInternalKey(key InternalKey, value interface{}, in map[string]interface{}, overwrite bool) error {
	if _, ok := in[string(key)]; !ok || overwrite {
		in[string(key)] = value
		return nil
	}
	return util.NewConflictError(fmt.Sprintf("'%s' already exists in map", key))
}

func GetFromInternalPrefix(prefix InternalKeyPrefix, suffix string, in map[string]interface{}) (interface{}, bool) {
	value, ok := in[string(prefix) + ":" + suffix]
	return value, ok
}

func InternalKeyFromPrefix(prefix InternalKeyPrefix, suffix string) string {
	return fmt.Sprintf("%s:%s", prefix.String(), suffix)
}

func SetUsingInternalPrefix(prefix InternalKeyPrefix, suffix string, value interface{}, in map[string]interface{},
	overwrite bool) error {
	key := string(prefix) + ":" + suffix
	if _, ok := in[key]; !ok || overwrite {
		in[key] = value
		return nil
	}
	return util.NewConflictError(fmt.Sprintf("'%s' already exists in map", key))
}
