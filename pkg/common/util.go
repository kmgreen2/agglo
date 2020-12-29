package common

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"hash"
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

func WaitAll(futures []Future, timeout time.Duration) {
	timedOut := false
	done := make(chan bool, 1)
	completedFutures := make(map[int]Future)

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
			timedOut = true
		case <-done:
			return
		}
	}()


	for len(completedFutures) != len(futures) && !timedOut {
		for i, future := range futures {
			if future.IsCompleted() || future.IsCancelled() {
				completedFutures[i] = futures[i]
			}
		}
	}

	done <- true
}

