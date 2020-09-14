package conc

import "context"

type Task interface {
	Do(ctx context.Context) error
}

type TaskFunc func(ctx context.Context) error

func (t TaskFunc) Do(ctx context.Context) error {
	return t(ctx)
}
