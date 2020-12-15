package core

type Extractable interface {
	GetContent() (Content, error)
}

type ExtractedValue interface {
	QualifierValue
	Label() MetadataQualifier
}

type Extractor interface {
	Extract(extractable Extractable) ([]ExtractedValue, error)
}