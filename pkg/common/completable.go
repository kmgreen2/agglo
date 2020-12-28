package common

type Completable interface {
	Success(result interface{}) error
	Fail(err error) error
	Cancel() error
	Future() Future
	Close()
}

type promise struct {
	future *future
}

func NewCompletable() Completable {
	return &promise{
		future: newFuture(),
	}
}

func (p *promise) Success(result interface{}) error {
	return p.future.success(result)
}

func (p *promise) Fail(err error) error {
	return p.future.fail(err)
}

func (p *promise) Cancel() error {
	return p.future.Cancel()
}

func (p *promise) Future() Future {
	return p.future
}

func (p *promise) Close() {
	p.future.close()
}
