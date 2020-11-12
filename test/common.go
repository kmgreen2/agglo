package test

import (
	"bytes"
	gocrypto "crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/crypto"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/storage"
	"time"
)

// GetSignerAuthenticator will return a signer/authenticator pair for a
// given hash algorithm
func GetSignerAuthenticator(hashAlgorithm gocrypto.Hash) (*crypto.RSASigner,
	*crypto.RSAAuthenticator, *rsa.PublicKey, error) {
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)

	if err != nil {
		return nil, nil, nil, err
	}

	signer := crypto.NewRSASigner(key, hashAlgorithm)
	authenticator := crypto.NewRSAAuthenticator(&key.PublicKey, hashAlgorithm)

	return signer, authenticator, &key.PublicKey, nil
}

func GetTestObjectStoreWithObjects(backendType storage.BackendType,
	objects map[string]*bytes.Buffer, defaultObjectStore bool) (storage.ObjectStore,
	storage.ObjectStoreBackendParams, error) {
	objectStoreInstance := "default"
	if !defaultObjectStore {
		objectStoreInstance = gUuid.New().String()
	}
	storageParams, err := storage.NewMemObjectStoreBackendParams(backendType, objectStoreInstance)
	if err != nil {
		return nil, nil, err
	}
	objectStore, err := storage.NewMemObjectStore(storageParams)
	if err != nil {
		return nil, nil, err
	}

	for key, reader := range objects {
		err = objectStore.Put(key, reader)
		if err != nil {
			return nil, nil, err
		}
	}

	return objectStore, storageParams, nil
}

func GetTestObjects(objectStore storage.ObjectStore, numObjects int) (map[string]*bytes.Buffer, error) {
	objectMap := make(map[string]*bytes.Buffer)
	for i := 0; i < numObjects; i++ {
		key := fmt.Sprintf("%d", i)
		objectMap[key] = bytes.NewBuffer([]byte(key))
		err := objectStore.Put(key, objectMap[key])
		if err != nil {
			return nil, err
		}
	}
	return objectMap, nil
}

func GetSubStream(subStreamID entwine.SubStreamID, numMessages int, defaultObjectStore bool,
	prevMessage *entwine.StreamImmutableMessage) ([]*entwine.StreamImmutableMessage, crypto.Authenticator,
	storage.ObjectStore, error) {
	messages := make([]*entwine.StreamImmutableMessage, numMessages)
	signer, authenticator, _, err := GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		return nil, nil, nil, err
	}

	anchorUuid := gUuid.New()
	i := 0
	if prevMessage == nil {
		messages[0], err = entwine.NewStreamGenesisMessage(subStreamID, common.SHA1, signer, anchorUuid)
		if err != nil {
			return nil, nil, nil, err
		}
		prevMessage = messages[0]
		i++
	}

	objectStoreInstance := "default"
	if !defaultObjectStore {
		objectStoreInstance = gUuid.New().String()
	}

	objectStoreParams, err := storage.NewMemObjectStoreBackendParams(storage.MemObjectStoreBackend, objectStoreInstance)
	if err != nil {
		return nil, nil, nil, err
	}

	objectStore, err := storage.NewObjectStore(objectStoreParams)
	if err != nil {
		return nil, nil, nil, err
	}

	objectMap, err := GetTestObjects(objectStore, numMessages - i)
	if err != nil {
		return nil, nil, nil, err
	}

	for key, _ := range objectMap {
		objectDescriptor := storage.NewObjectDescriptor(objectStoreParams, key)
		messages[i], err = entwine.NewStreamImmutableMessage(subStreamID, objectDescriptor, key, []string{},
			common.SHA1, signer, time.Now().Unix(), prevMessage, anchorUuid)
		prevMessage = messages[i]
		i++
	}

	return messages, authenticator, objectStore, nil
}

func GetTickerStream(numMessages int) ([]*entwine.TickerImmutableMessage, crypto.Authenticator, error) {
	messages := make([]*entwine.TickerImmutableMessage, numMessages)
	signer, authenticator, _, err := GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		return nil, nil, err
	}

	var prevMessage *entwine.TickerImmutableMessage = nil
	for i := 0; i < numMessages; i++ {
		messages[i], err = entwine.NewTickerImmutableMessage(common.SHA1, signer, int64(i), prevMessage)
		if err != nil {
			return nil, nil, err
		}
		prevMessage = messages[i]
	}
	return messages, authenticator, nil
}

func GetKVStreamStore(numMessages int, subStreamID entwine.SubStreamID, signer crypto.Signer,
	anchorTickerUuid gUuid.UUID, newAnchorStride int) (*entwine.KVStreamStore, kvs.KVStore, storage.ObjectStore,
	map[string]string, error) {
	currAnchorUuid := anchorTickerUuid
	objects := make(map[string]*bytes.Buffer)
	uuidToName := make(map[string]string)
	kvStore := kvs.NewMemKVStore()
	kvStreamStore := entwine.NewKVStreamStore(kvStore, common.SHA1)

	err := kvStreamStore.Create(subStreamID, common.SHA1, signer, currAnchorUuid)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	for i := 0; i < numMessages; i++ {
		objKey := fmt.Sprintf("%d", i)
		objects[objKey] = bytes.NewBuffer([]byte(objKey))
	}
	objectStore, storageParams, err := GetTestObjectStoreWithObjects(storage.MemObjectStoreBackend, objects,
		false)
	if err != nil {
		return nil, nil, nil, nil, err
	}


	for i := 0; i < numMessages; i++ {
		objKey := fmt.Sprintf("%d", i)
		desc := storage.NewObjectDescriptor(storageParams, objKey)
		message := entwine.NewUncommittedMessage(desc, objKey, []string{objKey}, signer)

		if newAnchorStride > 0 && i % newAnchorStride == 0 {
			currAnchorUuid = gUuid.New()
		}
		uuid, err := kvStreamStore.Append(message, subStreamID, currAnchorUuid)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		uuidToName[uuid.String()] = objKey
	}
	return kvStreamStore, kvStore, objectStore, uuidToName, nil
}

func GetTickerStore(numMessages int) (*entwine.KVTickerStore, crypto.Signer, error) {
	kvStore := kvs.NewMemKVStore()
	kvTickerStore := entwine.NewKVTickerStore(kvStore, common.SHA1)
	signer, _, _, err := GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		return nil, nil, err
	}

	for i := 0; i < numMessages; i++ {
		err = kvTickerStore.Append(signer)
		if err != nil {
			return nil, nil, err
		}
	}
	return kvTickerStore, signer, nil
}

func GetProofStream(startNumTicks, tickStride, messageStride, numEntanglements int,
	subStreamID entwine.SubStreamID) (*entwine.
	KVTickerStore, *entwine.KVStreamStore, error) {

	var err error

	numMessages := messageStride * numEntanglements
	tickerStore, tickerSigner, err := GetTickerStore(startNumTicks)
	if err != nil {
		return nil, nil, err
	}

	genesisProof, err := tickerStore.CreateGenesisProof(subStreamID)
	if err != nil {
		return nil, nil, err
	}

	messageSigner, messageAuthenticator, _, err := GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		return nil, nil, err
	}


	kvStreamStore, _, _, _, err := GetKVStreamStore(0, subStreamID, messageSigner,
		genesisProof.TickerUuid(), 0)
	if err != nil {
		return nil, nil, err
	}


	objects := make(map[string]*bytes.Buffer)
	for i := 0; i < numMessages; i++ {
		objKey := fmt.Sprintf("%d", i)
		objects[objKey] = bytes.NewBuffer([]byte(objKey))
	}

	_, storageParams, err := GetTestObjectStoreWithObjects(storage.MemObjectStoreBackend, objects,
		false)
	if err != nil {
		return nil, nil, err
	}

	k := 0
	anchorUuid := genesisProof.TickerUuid()
	head, err  := kvStreamStore.Head(subStreamID)
	if err != nil {
		return nil, nil, err
	}
	endUuid := head.Uuid()
	for i := 0; i < numEntanglements; i++ {
		for j := 0; j < messageStride; j++ {
			objKey := fmt.Sprintf("%d", k)
			desc := storage.NewObjectDescriptor(storageParams, objKey)
			message := entwine.NewUncommittedMessage(desc, objKey, []string{}, messageSigner)
			endUuid, err = kvStreamStore.Append(message, subStreamID, anchorUuid)
			if err != nil {
				return nil, nil, err
			}
			k++
			if k % tickStride == 0 {
				err = tickerStore.Append(tickerSigner)
				if err != nil {
					return nil, nil, err
				}
			}
		}
		startUuid, err := tickerStore.GetProofStartUuid(subStreamID)
		if err != nil {
			return nil, nil, err
		}
		messages, err := kvStreamStore.GetHistory(startUuid, endUuid)
		if err != nil {
			return nil, nil, err
		}

		valid, err := entwine.ValidateStreamMessages(messages, messageAuthenticator)
		if err != nil {
			return nil, nil, err
		}
		if !valid {
			return nil, nil, fmt.Errorf("Generated invalid message stream furing %d-th entanglement", i)
		}
		anchorMessage, err := tickerStore.Anchor(messages, subStreamID, messageAuthenticator)
		if err != nil {
			return nil, nil, err
		}
		anchorUuid = anchorMessage.Uuid()
	}
	return tickerStore, kvStreamStore, nil
}
