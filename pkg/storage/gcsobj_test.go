package storage_test

import (
	"context"
	"crypto/sha1"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func SetupGCSLocal() (*fakestorage.Server, error){
	return fakestorage.NewServerWithOptions(fakestorage.Options{
		InitialObjects: []fakestorage.Object{
			{
				BucketName: "localtest",
				Name:       "some/object/foo",
				Content:    []byte("bar"),
			},
		},
		Scheme: "http",
		Host: "localhost",
		Port: 8081,
		Writer: os.Stdout,
		PublicHost: "localtest.localhost",
	})
}

// This is due to a "bug" in the fake GCS server that expects <bucketname>.<host> URL scheme.  We need to hijack
// DNS resolution to resolve localtest.localhost -> localhost
func GetFakeTransport() *http.Transport {
	transport := &http.Transport{
	}
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if addr == "localtest.localhost:8081" {
			addr = "localhost:8081"
		}
		return dialer.DialContext(ctx, network, addr)
	}

	return transport
}

func LocalGCSObjectStoreParams() (*storage.GCSObjectStoreBackendParams, error) {
	return storage.NewGCSObjectStoreBackendParams(storage.GCSObjectStoreBackend,
		"localtest")
}

func TestGCSHappyPath(t *testing.T) {
	fileSize := 1024
	fakeGCS, err := SetupGCSLocal()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer fakeGCS.Stop()
	params, err := LocalGCSObjectStoreParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	os.Setenv("STORAGE_EMULATOR_HOST", "localhost:8081")
	objStore, err := storage.NewGCSObjectStore(params, option.WithoutAuthentication(),
		option.WithHTTPClient(&http.Client{Transport: GetFakeTransport()}),
		option.WithEndpoint("http://localtest.localhost:8081/storage/v1"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	randomReader := NewRandomReader(fileSize)

	key := gUuid.New().String()

	err = objStore.Put(context.Background(), key, randomReader)
	assert.Nil(t, err)

	objHash := randomReader.Hash()

	err = objStore.Head(context.Background(), key)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	reader, err := objStore.Get(context.Background(), key)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	readBytes := make([]byte, 1024)
	readDigest := sha1.New()
	isDone := false
	for {
		if isDone {
			break
		}
		numRead, err := reader.Read(readBytes)
		if errors.Is(err, io.EOF) {
			isDone = true
		} else if err != nil {
			break
		}
		_, err = readDigest.Write(readBytes[:numRead])
		if err != nil {
			break
		}
	}
	assert.Equal(t, objHash, readDigest.Sum(nil))

	err = objStore.Delete(context.Background(), key)
	assert.Nil(t, err)
}