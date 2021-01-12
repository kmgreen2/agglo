package common

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"hash"
	"sync"
	"time"
)

// Used to distinguish between different Digest algorithms
type DigestType int

const (
	SHA1 DigestType = iota
	SHA256
	MD5
)

// Construct a hash object using a supported Digest
// type.  If the Digest type is not supported, return
// nil.
func InitHash(digestType DigestType) hash.Hash {
	switch digestType {
	case SHA1:
		return sha1.New()
	case SHA256:
		return sha256.New()
	case MD5:
		return md5.New()
	default:
		return nil
	}
}

// ToDo(KMG): Re-visit this function.  I could not think of a way to
// use WaitGroups without leaking a go routine when the Wait() call
// hangs forever when we set a timeout.  The best I could think of was
// to track the number of waiters and decrement the count when we timeout.
//
// This can probably done with atomic incr/decr and channels
func WaitAll(futures []Future, timeout time.Duration) {
	lock := &sync.Mutex{}
	numFutures := len(futures)
	done := make(chan bool, 1)
	wg := sync.WaitGroup{}
	wg.Add(numFutures)

	go func() {
		var ctx context.Context
		var cancel context.CancelFunc

		ctx = context.Background()
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		select {
		case <-ctx.Done():
			lock.Lock()
			defer lock.Unlock()
			if numFutures == 0 {
				return
			}
			wg.Add(-numFutures)
		case <-done:
			return
		}
	}()

	for _, future := range futures {
		future.OnSuccess(func(ctx context.Context, x interface{}) {
			lock.Lock()
			defer lock.Unlock()
			numFutures--
			wg.Done()
		}).OnFail(func(ctx context.Context, err error) {
			lock.Lock()
			defer lock.Unlock()
			numFutures--
			wg.Done()
		})
	}

	wg.Wait()
	done <- true
}

func MapToJson(in map[string]interface{}) ([]byte, error) {
	byteBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(byteBuffer)
	err := encoder.Encode(in)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func JsonToMap(in []byte) (map[string]interface{}, error) {
	var out map[string]interface{}
	byteBuffer := bytes.NewBuffer(in)
	decoder := json.NewDecoder(byteBuffer)
	err := decoder.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

