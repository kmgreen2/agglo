package core

import (
	"github.com/kmgreen2/agglo/pkg/kvs"
	"time"
)

type Property struct {
	in string
	out string
}

type Join struct {
	key string
	properties []Property
}

type Completion struct {
	name string
	joins []Join
	timeout time.Duration
	notifyIfPartial bool
	kvStore kvs.KVStore
}

