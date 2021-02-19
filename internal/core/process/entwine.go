package process

import (
	"bytes"
	"context"
	gocrypto "crypto"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/internal/common"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/pkg/crypto"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/kmgreen2/agglo/pkg/util"
)

var EntwineMetadataKey string = string(common.EntwineMetadataKey)

type Entwine struct {
	name string
	appender *entwine.SubStreamAppender
	objectStore storage.ObjectStore
	signer crypto.Signer
	subStreamID entwine.SubStreamID
	condition *core.Condition
	tickerURL string
	tickerInterval int
	currTickerUUID gUuid.UUID
	numMessages int
}

func NewEntwine(name, subStreamID, pem string, kvStore kvs.KVStore, objectStore storage.ObjectStore, tickerURL string,
	tickerInterval int,
	condition *core.Condition) (*Entwine,
	error) {
	entwiner := &Entwine{
		name: name,
		subStreamID: entwine.SubStreamID(subStreamID),
		tickerURL: tickerURL,
		tickerInterval: tickerInterval,
		objectStore: objectStore,
		condition: condition,
	}

	privateKey, err := crypto.ParseRSAPrivateKeyFromPEM(pem)
	if err != nil {
		return nil, err
	}

	entwiner.signer = crypto.NewRSASigner(privateKey, gocrypto.SHA256)

	streamStore := entwine.NewKVStreamStore(kvStore, util.SHA256)
	entwiner.appender = entwine.NewSubStreamAppender(streamStore, entwine.SubStreamID(subStreamID))

	return entwiner, nil
}

func (e Entwine) Name() string {
	return e.name
}

func (e Entwine) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
	shouldEntwine, err := e.condition.Evaluate(in)
	if err != nil {
		return in, PipelineProcessError(e, err, "evaluating condition")
	}

	if !shouldEntwine {
		return in, nil
	}

	uuid, err := gUuid.NewRandom()
	if err != nil {
		return nil, PipelineProcessError(e, err, "generating UUID")
	}

	out := util.CopyableMap(in).DeepCopy()
	if _, ok := out[EntwineMetadataKey]; !ok {
		out[EntwineMetadataKey] = make([]map[string]interface{}, 0)
	}

	mapBytes, err := util.MapToJson(in)
	if err != nil {
		return nil, PipelineProcessError(e, err, "serializing map to JSON")
	}

	err = e.objectStore.Put(ctx, uuid.String(), bytes.NewBuffer(mapBytes))
	if err != nil {
		return nil, PipelineProcessError(e, err, "storing map to object store")
	}

	// Anchor with ticker store if necessary.  If there is no ticker UUID, get it from the ticker store
	if e.numMessages % e.tickerInterval == 0 || e.currTickerUUID == gUuid.Nil {
	}

	// Need to get ObjectStoreBackendParams from objectStore
	// Add to the interface for ObjectStore

	// desc := storage.NewObjectDescriptor(params, uuid.String())

	// message := entwine.NewUncommittedMessage(desc, uuid.String(), []string{}, e.signer)

	 // e.appender.Append(message, e.currTickerUUID)
	 return out, nil
}