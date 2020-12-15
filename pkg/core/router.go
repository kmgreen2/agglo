package core

type Routable interface {
	GetAttrs() (ContentAttrs, error)
}

type Router interface {
	Route(entity Routable) (Endpoint, error)
}
