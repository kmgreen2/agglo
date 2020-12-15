package core

type MetadataEntity []byte
type MetadataQualifier []byte
type EntityTag string

type MetadataKey interface {
	Entity() MetadataEntity
	Qualifier() MetadataQualifier
}

type EntityTags interface {
	Entity() MetadataEntity
	Tags() []EntityTag
}

// Metadata is assumed to be organized around an entity, where there are one or more qualifiers with values
type MetadataEntry interface {
	MetadataKey
	Value() QualifierValue
}

type QualifierOperator int
const (
	EqualOp = iota
	LessThanOp
	GreaterThanOp
	InOp
)

type QualifierValueType int
const (
	IntQualifier = iota
	StringQualifier
	IntListQualifier
	StringListQualifier
	ByteArrayQualifier
)

type QualifierValue interface {
	Get() interface{}
	Type() QualifierValueType
}

type QuerySpec struct {
	qualifiers []MetadataQualifier
	qualifierValues []QualifierValue
	qualifierOperator []QualifierOperator
	tags []EntityTag
	targetQualifier QualifierValue
}

type QueryBuilder struct {
	spec QuerySpec
}

func (builder *QueryBuilder) And(qualifier MetadataQualifier, op QualifierOperator,
	value QualifierValue) *QueryBuilder {

	builder.spec.qualifiers = append(builder.spec.qualifiers, qualifier)
	builder.spec.qualifierValues = append(builder.spec.qualifierValues, value)
	builder.spec.qualifierOperator = append(builder.spec.qualifierOperator, op)

	return builder
}

func (builder *QueryBuilder) With(tags []EntityTag) *QueryBuilder {
	builder.spec.tags = append(builder.spec.tags, tags...)
	return builder
}

func (builder *QueryBuilder) Get() QuerySpec {
	return builder.spec
}

type MetadataStore interface {
	Put(md MetadataEntry) error
	PutTags(tags []EntityTag) error
	Get(key MetadataKey) (MetadataEntry, error)
	Qualifiers(entity MetadataEntity) ([]MetadataQualifier, error)
	Tags(entity MetadataEntity) ([]EntityTag, error)
	Query(query QuerySpec) ([]MetadataEntry, error)
}

