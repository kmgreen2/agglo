package util

import "context"

type Completable interface {
	Success(ctx context.Context, result interface{}) error
	Fail(ctx context.Context, err error) error
	Cancel(ctx context.Context) error
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

func (p *promise) Success(ctx context.Context, result interface{}) error {
	return p.future.success(ctx, result)
}

func (p *promise) Fail(ctx context.Context, err error) error {
	return p.future.fail(ctx, err)
}

func (p *promise) Cancel(ctx context.Context) error {
	return p.future.Cancel(ctx)
}

func (p *promise) Future() Future {
	return p.future
}

func (p *promise) Close() {
	p.future.close()
}
