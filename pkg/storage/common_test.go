package storage_test

import (
	"context"
	"crypto/sha1"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/pkg/errors"
	"hash"
	"io"
	"math/rand"
)

type RandomReader struct {
	len int
	offset int
	digest hash.Hash
}

func NewRandomReader(len int) *RandomReader {
	return &RandomReader{
		len: len,
		offset: 0,
		digest: sha1.New(),
	}
}

func (reader *RandomReader) Read(b []byte) (int, error) {
	if reader.offset >= reader.len {
		return 0, io.EOF
	}
	numLeft := reader.len - reader.offset
	bufRef := b
	if len(b) > numLeft {
		bufRef = b[:numLeft]
	}
	numRead, err := rand.Read(bufRef)
	if err != nil && !errors.Is(err, io.EOF) {
		return -1, err
	}
	_, err = reader.digest.Write(bufRef)
	if err != nil {
		return -1, err
	}
	reader.offset += numRead
	return numRead, nil
}

func (reader *RandomReader) Hash() []byte {
	return reader.digest.Sum(nil)
}

type BadReader struct {
}

func (reader *BadReader) Read(b []byte) (int ,error) {
	return -1, &util.InternalError{}
}

func putObjects(objStore storage.ObjectStore, prefix string, numWithPrefix int, numWithoutPrefix int) error {
	for i := 0; i < numWithPrefix; i++ {
		randomReader := NewRandomReader(1024)
		err := objStore.Put(context.Background(), fmt.Sprintf("%s%d", prefix, i), randomReader)
		if err != nil {
			return err
		}
	}
	for i := 0; i < numWithoutPrefix; i++ {
		randomReader := NewRandomReader(1024)
		err := objStore.Put(context.Background(), fmt.Sprintf("%d%s%d", i, prefix, i), randomReader)
		if err != nil {
			return err
		}
	}
	return nil
}

