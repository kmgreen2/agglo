package core

import "github.com/kmgreen2/agglo/pkg/storage"

type ContentType int

const (
	TEXT = iota
	JSON
	YAML
	PDF
	JPG
	GIF
)

type ContentAttrs map[string]string

type Content struct {
	raw []byte
	contentType ContentType
	contentAttrs ContentAttrs
}

type PersistedContent struct {
	desc storage.ObjectDescriptor
	entity MetadataEntity
	contentType ContentType
}
