package core

type Taggable interface {
	GetContent() (Content, error)
}

type Tagger interface {
	Tag(entity Taggable) []EntityTag
}

