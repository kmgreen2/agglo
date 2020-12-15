package core

type Spawnable interface {
	GetPersistedContent() (PersistedContent, error)
}

type Spawner interface {
	Spawn(spawnable Spawnable) error
}