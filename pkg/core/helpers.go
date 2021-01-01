package core

import (
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
)

func GetPartitionID(in map[string]interface{}) (gUuid.UUID, error) {
	if partitionIDRaw, ok := in["agglo:internal:partitionID"]; ok {
		if partitionID, ok := partitionIDRaw.(string); ok {
			return gUuid.Parse(partitionID)
		}
	}
	msg := fmt.Sprintf("could not find valid 'agglo:internal:partitionID' in payload")
	return gUuid.Nil, common.NewInvalidError(msg)
}

func GetName(in map[string]interface{}) (string, error) {
	if nameRaw, ok := in["agglo:internal:name"]; ok {
		if name, ok := nameRaw.(string); ok {
			return name, nil
		}
	}
	msg := fmt.Sprintf("could not find valid 'agglo:internal:name' in payload")
	return "", common.NewInvalidError(msg)
}

