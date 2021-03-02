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
	"net/http"
	"net/url"
	"testing"
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
		Host: "127.0.0.1",
		Port: 8081,
	})
}

func LocalGCSObjectStoreParams() (*storage.GCSObjectStoreBackendParams, error) {
	return storage.NewGCSObjectStoreBackendParams(storage.GCSObjectStoreBackend,
		"localtest")
}

// Using the golang SDK with a GCS emulator is a PITA.
// Workaround: https://github.com/googleapis/google-cloud-go/issues/2476
type roundTripper url.URL
func (rt roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Host = rt.Host
	req.URL.Host = rt.Host
	req.URL.Scheme = rt.Scheme
	return http.DefaultTransport.RoundTrip(req)
}

func GetTestHttpClient(endpoint string) *http.Client {
	u, _ := url.Parse(endpoint)
	return &http.Client{Transport: roundTripper(*u)}
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
	objStore, err := storage.NewGCSObjectStore(params, option.WithoutAuthentication(),
		option.WithHTTPClient(GetTestHttpClient("http://127.0.0.1:8081")))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	randomReader := NewRandomReader(fileSize)

	key := gUuid.New().String()

	err = objStore.Put(context.Background(), key, randomReader)
	assert.Nil(t, err)

	objHash := randomReader.Hash()

	err = objStore.Head(context.Background(), key)
	assert.Nil(t, err)

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