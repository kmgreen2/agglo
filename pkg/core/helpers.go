package core

import (
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
)

func GetPartitionID(in map[string]interface{}) (gUuid.UUID, error) {
	if partitionIDRaw, ok := common.GetFromInternalKey(common.PartitionIDKey, in); ok {
		if partitionID, ok := partitionIDRaw.(string); ok {
			return gUuid.Parse(partitionID)
		}
	}
	msg := fmt.Sprintf("could not find valid 'partitionID' in payload")
	return gUuid.Nil, common.NewInvalidError(msg)
}

func GetName(in map[string]interface{}) (string, error) {
	if nameRaw, ok := common.GetFromInternalKey(common.ResourceNameKey, in); ok {
		if name, ok := nameRaw.(string); ok {
			return name, nil
		}
	}
	msg := fmt.Sprintf("could not find valid 'resource name' in payload")
	return "", common.NewInvalidError(msg)
}

