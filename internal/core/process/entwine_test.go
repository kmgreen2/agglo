package process_test

import (
	"context"
	gocrypto "crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
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
	kvStore kvs.KVStore, objectStore storage.ObjectStore, tickerClient client.TickerClient, tickerInterval int,
	condition *core.Condition) ([]map[string]interface{}, error) {

	var outMaps []map[string]interface{}

	entwineProc, err := process.NewEntwine("test", subStreamID, privatePem, kvStore, objectStore, tickerClient,
		tickerInterval, condition)

	if err != nil {
		return nil, err
	}

	for _, m := range maps {
		out, err := entwineProc.Process(context.Background(), m)
		if err != nil {
			return nil, err
		}
		outMaps = append(outMaps, out)
	}

	return outMaps, nil
}

func TestEntwineHappyPath(t *testing.T) {
	numMaps := 20
	var maps []map[string]interface{}
	tickerInterval := 4
	subStreamID := entwine.SubStreamID(gUuid.New().String())
	kvStore := kvs.NewMemKVStore()
	objStore, err := storage.NewObjectStoreFromConnectionString("mem:testing")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	tickerStore := entwine.NewKVTickerStore(kvStore, util.SHA256)

	ctrl := gomock.NewController(t)
	mockTicker := mocks.NewMockTickerClient(ctrl)

	reader := rand.Reader
	bitSize := 2048

	privateKey, err := rsa.GenerateKey(reader, bitSize)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	authenticator := crypto.NewRSAAuthenticator(&privateKey.PublicKey, gocrypto.SHA256)

	// Use same signer for ticker (in reality this should be using a different key pair)
	signer := crypto.NewRSASigner(privateKey, gocrypto.SHA256)

	privatePem := crypto.ExportRSAPrivateKeyAsPEM(privateKey)

	var mockCalls  []*gomock.Call
	for i := 0; i < numMaps; i++ {
		// Make sure we Tick the Ticker
		err = tickerStore.Append(signer)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		if i % tickerInterval == 0 {
			if i == 0 {
				mockCalls = append(mockCalls, mockTicker.EXPECT().CreateGenesisProof(gomock.Any(),
					gomock.Eq(subStreamID)).
					DoAndReturn(func (ctx context.Context,
					subStreamID entwine.SubStreamID) (gUuid.UUID, error) {
						tickerMessage, err := tickerStore.CreateGenesisProof(subStreamID)
						if err != nil {
							return gUuid.Nil, err
						}
						return tickerMessage.TickerUuid(), nil
					}).Times(1))
			} else {
				mockCalls = append(mockCalls, mockTicker.EXPECT().GetProofStartUuid(gomock.Any(),
					gomock.Eq(subStreamID)).DoAndReturn(func (ctx context.Context,
					subStreamID entwine.SubStreamID) (gUuid.UUID, error) {
						return tickerStore.GetProofStartUuid(subStreamID)
					}).Times(1))
				mockCalls = append(mockCalls, mockTicker.EXPECT().Anchor(gomock.Any(), gomock.Any(), gomock.Eq(subStreamID)).
					DoAndReturn(func(ctx context.Context, proof []*entwine.StreamImmutableMessage,
						subStreamID entwine.SubStreamID) (*entwine.TickerImmutableMessage, error){
						fmt.Printf("%v", proof)
						return tickerStore.Anchor(proof, subStreamID, authenticator)
					}).Times(1))
			}
		}
		m, _ := test.GenRandomMap(2, 24)
		maps = append(maps, m)
	}

	gomock.InOrder(mockCalls...)

	outMaps, err := doEntwineMessages(maps, subStreamID, privatePem, kvStore, objStore, mockTicker, tickerInterval,
		core.TrueCondition)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	for _, m := range outMaps {
		jsonMap, _ := util.MapToJson(m[process.EntwineMetadataKey].([]map[string]interface{})[0])
		fmt.Println(string(jsonMap))
	}
}