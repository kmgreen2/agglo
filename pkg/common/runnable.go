package common

type Runnable interface {
	Run() (interface{}, error)
}

type PartialRunnable interface {
	Runnable
	SetArgs(args ...interface{}) error
}

