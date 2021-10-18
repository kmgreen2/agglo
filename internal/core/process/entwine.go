package process

import (
	"bytes"
	"context"
	gocrypto "crypto"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/internal/common"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/pkg/client"
	"github.com/kmgreen2/agglo/pkg/crypto"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/pkg/errors"
	"reflect"
)

var EntwineMetadataKey string = string(common.EntwineMetadataKey)

type Entwine struct {
	name string
	appender *entwine.SubStreamAppender
	objectStore storage.ObjectStore
	signer crypto.Signer
	subStreamID entwine.SubStreamID
	streamStore entwine.StreamStore
	condition *core.Condition
	ticker client.TickerClient
	tickerInterval int
	currTickerUUID gUuid.UUID
	numMessages int
}

func NewEntwine(name string, subStreamID entwine.SubStreamID, pem string, kvStore kvs.KVStore, objectStore storage.ObjectStore,
	tickerClient client.TickerClient, tickerInterval int, condition *core.Condition) (*Entwine,error) {
	entwiner := &Entwine{
		name: name,
		subStreamID: entwine.SubStreamID(subStreamID),
		tickerInterval: tickerInterval,
		objectStore: objectStore,
		condition: condition,
	}

	privateKey, err := crypto.ParseRSAPrivateKeyFromPEM(pem)
	if err != nil {
		return nil, err
	}

	entwiner.signer = crypto.NewRSASigner(privateKey, gocrypto.SHA256)

	entwiner.streamStore = entwine.NewKVStreamStore(kvStore, util.SHA256)
	entwiner.appender = entwine.NewSubStreamAppender(entwiner.streamStore, subStreamID)

	if tickerClient != nil {
		entwiner.ticker = tickerClient

		tickerUuid, err := entwiner.appender.GetAnchorUuid()
		if err != nil && errors.Is(err, &util.NotFoundError{}) {
			tickerUuid, err = entwiner.ticker.CreateGenesisProof(context.Background(), entwiner.subStreamID)
			if err != nil {
				return nil, err
			}
			// Be sure to set the genesis anchor
			err = entwiner.appender.SetAnchorUuid(tickerUuid)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		entwiner.currTickerUUID = tickerUuid
	}

	_, err = entwiner.streamStore.Head(entwiner.subStreamID)
	if err != nil && errors.Is(&util.NotFoundError{}, err) {
		err = entwiner.streamStore.Create(entwiner.subStreamID, util.SHA256, entwiner.signer, entwiner.currTickerUUID)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return entwiner, nil
}

func (e Entwine) Name() string {
	return e.name
}

func (e *Entwine) Process(ctx context.Context, in map[string]interface{}) (out map[string]interface{}, err error) {

	defer func() {
		if r := recover(); r != nil {
			if rError, ok := r.(error); ok {
				err = errors.Wrap(rError, "")
			}
		}
	}()

	shouldEntwine, err := e.condition.Evaluate(in)
	if err != nil {
		return in, PipelineProcessError(e, err, "evaluating condition")
	}

	if !shouldEntwine {
		return in, nil
	}

	objectUuid, err := gUuid.NewRandom()
	if err != nil {
		return nil, PipelineProcessError(e, err, "generating UUID")
	}

	out = util.CopyableMap(in).DeepCopy()
	if _, ok := out[EntwineMetadataKey]; !ok {
		out[EntwineMetadataKey] = make([]map[string]interface{}, 0)
	}

	mapBytes, err := util.MapToJson(in)
	if err != nil {
		return nil, PipelineProcessError(e, err, "serializing map to JSON")
	}

	err = e.objectStore.Put(ctx, objectUuid.String(), bytes.NewBuffer(mapBytes))
	if err != nil {
		return nil, PipelineProcessError(e, err, "storing map to object store")
	}

	params := e.objectStore.ObjectStoreBackendParams()

	desc := storage.NewObjectDescriptor(params, objectUuid.String())

	message := entwine.NewUncommittedMessage(desc, objectUuid.String(), []string{}, e.signer)

	// WriteLock/WriteUnlock will ensure the anchor operation is synchronized with the append operations and
	// other anchor operations for the same substream
	appendLockCtx, err := e.appender.WriteLock()
	if err != nil {
		return nil, PipelineProcessError(e, err, "write lock for subStream")
	}

	defer func() {
		_ = e.appender.WriteUnlock(appendLockCtx)
	}()

	// There is a race with the curr ticker UUID, where the wrong UUID is used because
	// an anchor operation could happen concurrently
	// ToDo(KMG): Need to handle ticker UUID internally => get anchor UUID inside of append (under lock)
	entwineUuid, err := e.appender.Append(appendLockCtx, message)
	if err != nil {
		return nil, PipelineProcessError(e, err, "appending")
	}

	// Anchor with ticker store if necessary.
	if e.ticker != nil && (e.numMessages % e.tickerInterval == 0) && e.numMessages != 0 {
		// Anchor from last known message with ticker to *this* latest appended message
		startUuid, err := e.ticker.GetProofStartUuid(ctx, e.subStreamID)
		if err != nil {
			return nil, PipelineProcessError(e, err, "getting proof UUID")
		}

		messages, err := e.appender.GetHistory(startUuid, entwineUuid)
		if err != nil {
			return nil, PipelineProcessError(e, err, "getting history")
		}

		var anchorUuids []string
		for _, m := range messages {
			anchorUuids = append(anchorUuids, m.Uuid().String()[:4])
		}

		if len(messages) > 0 {
			anchor, err := e.ticker.Anchor(ctx, messages, e.subStreamID)
			if err != nil {
				return nil, PipelineProcessError(e, err, "anchoring")
			}
			err = e.appender.SetAnchorUuid(anchor.Uuid())
			if err != nil {
				return nil, PipelineProcessError(e, err, "storing anchor UUID")
			}
			e.currTickerUUID = anchor.Uuid()
		}
	}

	committedMessage, err := e.streamStore.GetMessageByUUID(entwineUuid)
	if err != nil {
		return nil, PipelineProcessError(e, err, "error getting committed message")
	}

	switch outVal := out[EntwineMetadataKey].(type) {
	case []map[string]interface{}:
		outputMap := map[string]interface{}{
			"objectDescriptor": map[string]interface{} {
				"objectKey": objectUuid.String(),
				"objectStoreConnectionString": e.objectStore.ConnectionString(),
			},
			"entwineUuid": entwineUuid.String(),
			"tickerUuid": e.currTickerUUID.String(),
			"subStreamID": e.subStreamID,
			"subStreamIndex": committedMessage.Index(),
		}
		out[EntwineMetadataKey] = append(outVal, outputMap)
	default:
		msg := fmt.Sprintf("detected corrupted %s in map when entwining.  expected []map[string]string, got %v",
			EntwineMetadataKey, reflect.TypeOf(outVal))
		return nil, PipelineProcessError(e, util.NewInternalError(msg), "setting output map")
	}

	e.numMessages++

	return out, nil
}