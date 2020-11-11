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
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/kmgreen2/agglo/pkg/ticker"
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
	objects map[string]*bytes.Buffer) (storage.ObjectStore, error) {
	storageParams, err := storage.NewMemObjectStoreBackendParams(backendType, "default")
	if err != nil {
		return nil, err
	}
	objectStore, err := storage.NewMemObjectStore(storageParams)
	if err != nil {
		return nil, err
	}

	for key, reader := range objects {
		err = objectStore.Put(key, reader)
		if err != nil {
			return nil, err
		}
	}

	return objectStore, nil
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

func GetSubStream(subStreamID ticker.SubStreamID, numMessages int,
	prevMessage *ticker.StreamImmutableMessage) ([]*ticker.StreamImmutableMessage, crypto.Authenticator,
	storage.ObjectStore, error) {
	messages := make([]*ticker.StreamImmutableMessage, numMessages)
	signer, authenticator, _, err := GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		return nil, nil, nil, err
	}

	anchorUuid := gUuid.New()
	i := 0
	if prevMessage == nil {
		messages[0], err = ticker.NewGenesisMessage(subStreamID, common.SHA1, signer, anchorUuid)
		if err != nil {
			return nil, nil, nil, err
		}
		prevMessage = messages[0]
		i++
	}

	objectStoreParams, err := storage.NewMemObjectStoreBackendParams(storage.MemObjectStoreBackend, "default")
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
		messages[i], err = ticker.NewStreamImmutableMessage(subStreamID, objectDescriptor, key, []string{},
			common.SHA1, signer, time.Now().Unix(), prevMessage, anchorUuid)
		prevMessage = messages[i]
		i++
	}

	return messages, authenticator, objectStore, nil
}
