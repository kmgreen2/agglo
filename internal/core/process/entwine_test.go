package process_test

import (
	"context"
	gocrypto "crypto"
	"crypto/rand"
	"crypto/rsa"
	"github.com/golang/mock/gomock"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/internal/core/process"
	"github.com/kmgreen2/agglo/pkg/client"
	"github.com/kmgreen2/agglo/pkg/crypto"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/test"
	mocks "github.com/kmgreen2/agglo/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func doEntwineMessages(maps []map[string]interface{}, subStreamID entwine.SubStreamID, privatePem string,
	kvStore kvs.KVStore, objectStore storage.ObjectStore, tickFunc func(),
	tickerClient client.TickerClient, tickerInterval int, condition *core.Condition) ([]map[string]interface{}, error) {

	var outMaps []map[string]interface{}

	entwineProc, err := process.NewEntwine("test", subStreamID, privatePem, kvStore, objectStore, tickerClient,
		tickerInterval, condition)

	if err != nil {
		return nil, err
	}

	for _, m := range maps {
		tickFunc()
		out, err := entwineProc.Process(context.Background(), m)
		if err != nil {
			return nil, err
		}
		outMaps = append(outMaps, out)
	}

	return outMaps, nil
}

func entwineSubStream(t *testing.T, numMaps, anchorInterval int, subStreamID entwine.SubStreamID,
	privateKey *rsa.PrivateKey, kvStore kvs.KVStore, objStore storage.ObjectStore,
	tickerStore entwine.TickerStore, tickFunc func(), startTickerUuid gUuid.UUID) ([]map[string]interface{},
	gUuid.UUID, error) {

	var tickerUuids []gUuid.UUID
	var maps []map[string]interface{}

	ctrl := gomock.NewController(t)
	mockTicker := mocks.NewMockTickerClient(ctrl)

	authenticator := crypto.NewRSAAuthenticator(&privateKey.PublicKey, gocrypto.SHA256)

	privatePem := crypto.ExportRSAPrivateKeyAsPEM(privateKey)

	var mockCalls  []*gomock.Call
	for i := 0; i < numMaps; i++ {
		if i % anchorInterval == 0 {
			if i == 0 {
				if startTickerUuid == gUuid.Nil {
					mockCalls = append(mockCalls, mockTicker.EXPECT().CreateGenesisProof(gomock.Any(),
						gomock.Eq(subStreamID)).
						DoAndReturn(func(ctx context.Context,
							subStreamID entwine.SubStreamID) (gUuid.UUID, error) {
							tickerMessage, err := tickerStore.CreateGenesisProof(subStreamID)
							if err != nil {
								return gUuid.Nil, err
							}
							tickerUuids = append(tickerUuids, tickerMessage.TickerUuid())
							return tickerMessage.TickerUuid(), nil
						}).Times(1))
				} else {
					tickerUuids = append(tickerUuids, startTickerUuid)
				}
			} else {
				mockCalls = append(mockCalls, mockTicker.EXPECT().GetProofStartUuid(gomock.Any(),
					gomock.Eq(subStreamID)).DoAndReturn(func (ctx context.Context,
					subStreamID entwine.SubStreamID) (gUuid.UUID, error) {
					return tickerStore.GetProofStartUuid(subStreamID)
				}).Times(1))
				mockCalls = append(mockCalls, mockTicker.EXPECT().Anchor(gomock.Any(), gomock.Any(), gomock.Eq(subStreamID)).
					DoAndReturn(func(ctx context.Context, proof []*entwine.StreamImmutableMessage,
						subStreamID entwine.SubStreamID) (*entwine.TickerImmutableMessage, error){
						tickerMessage, err := tickerStore.Anchor(proof, subStreamID, authenticator)
						if err != nil {
							return nil, err
						}
						tickerUuids = append(tickerUuids, tickerMessage.Uuid())
						return tickerMessage, nil
					}).Times(1))
			}
		}
		m, _ := test.GenRandomMap(2, 24)
		maps = append(maps, m)
	}

	gomock.InOrder(mockCalls...)

	outMaps, err := doEntwineMessages(maps, subStreamID, privatePem, kvStore, objStore, tickFunc, mockTicker,
		anchorInterval,
		core.TrueCondition)

	if err != nil {
		return nil, gUuid.Nil, err
	}

	for i, m := range outMaps {
		entwineMetadata := m[process.EntwineMetadataKey].([]map[string]interface{})[0]
		assert.Equal(t, tickerUuids[i / anchorInterval].String(), entwineMetadata["tickerUuid"].(string))
	}

	return outMaps, tickerUuids[len(tickerUuids)-1], nil
}

func generatePrivateKey() (*rsa.PrivateKey, error) {
	reader := rand.Reader
	bitSize := 2048

	privateKey, err := rsa.GenerateKey(reader, bitSize)

	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func TestEntwineBasicHappyPath(t *testing.T) {
	numMaps := 20
	anchorInterval := 4
	tickerStore := entwine.NewKVTickerStore(kvs.NewMemKVStore(), util.SHA256)

	tickerPrivateKey, err := generatePrivateKey()
	tickerSigner := crypto.NewRSASigner(tickerPrivateKey, gocrypto.SHA256)

	tickFunc := func () {
		_ = tickerStore.Append(tickerSigner)
	}
	// Do first tick
	tickFunc()

	subStreamPrivateKey, err := generatePrivateKey()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	kvStore := kvs.NewMemKVStore()
	objStore, err := storage.NewObjectStoreFromConnectionString("mem:testing")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, _, err = entwineSubStream(t, numMaps, anchorInterval, entwine.SubStreamID(gUuid.New().String()),
		subStreamPrivateKey, kvStore, objStore, tickerStore, tickFunc, gUuid.Nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}

func TestEntwineMultipleHappyPath(t *testing.T) {
	numMaps := 20
	anchorInterval := 4
	tickerStore := entwine.NewKVTickerStore(kvs.NewMemKVStore(), util.SHA256)

	tickerPrivateKey, err := generatePrivateKey()
	tickerSigner := crypto.NewRSASigner(tickerPrivateKey, gocrypto.SHA256)

	tickFunc := func () {
		_ = tickerStore.Append(tickerSigner)
	}
	// Do first tick
	tickFunc()

	kvStore1 := kvs.NewMemKVStore()
	kvStore2 := kvs.NewMemKVStore()

	objStore1, err := storage.NewObjectStoreFromConnectionString("mem:testing1")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	objStore2, err := storage.NewObjectStoreFromConnectionString("mem:testing2")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	subStreamPrivateKey1, err := generatePrivateKey()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	subStreamPrivateKey2, err := generatePrivateKey()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	subStreamID1 := entwine.SubStreamID(gUuid.New().String())
	subStreamID2 := entwine.SubStreamID(gUuid.New().String())

	outMaps1, currTickerUuid1, err := entwineSubStream(t, numMaps, anchorInterval, subStreamID1,
		subStreamPrivateKey1, kvStore1, objStore1, tickerStore, tickFunc, gUuid.Nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	outMaps2, currTickerUuid2, err := entwineSubStream(t, numMaps, anchorInterval, subStreamID2,
		subStreamPrivateKey2, kvStore2, objStore2, tickerStore, tickFunc, gUuid.Nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	outMaps3, _, err := entwineSubStream(t, numMaps, anchorInterval, subStreamID1,
		subStreamPrivateKey1, kvStore1, objStore1, tickerStore, tickFunc, currTickerUuid1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	outMaps4, _, err := entwineSubStream(t, numMaps, anchorInterval, subStreamID2,
		subStreamPrivateKey2, kvStore2, objStore2, tickerStore, tickFunc, currTickerUuid2)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	// Validate ordering of all of the messages between the two streams
	streamStore1 := entwine.NewKVStreamStore(kvStore1, util.SHA256)
	streamStore2 := entwine.NewKVStreamStore(kvStore2, util.SHA256)

	var finalMessages []*entwine.StreamImmutableMessage

	for i, outMaps := range [][]map[string]interface{}{outMaps1, outMaps2, outMaps3, outMaps4} {
		streamStore := streamStore1
		if i % 2 == 1 {
			streamStore = streamStore2
		}
		for _, m := range outMaps {
			entwineMetadata := m[process.EntwineMetadataKey].([]map[string]interface{})[0]
			finalMessage, err := streamStore.GetMessageByUUID(gUuid.MustParse(entwineMetadata["entwineUuid"].(string)))
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			finalMessages = append(finalMessages, finalMessage)
		}
	}

	// Ensure that m[i] NOT happenedBefore m[i-1]
	for i, _ := range finalMessages {
		if i > 0 {
			ok, err := finalMessages[i].HappenedBefore(finalMessages[i-1], tickerStore)
			assert.Nil(t, err)
			assert.False(t, ok)
		}
	}
}