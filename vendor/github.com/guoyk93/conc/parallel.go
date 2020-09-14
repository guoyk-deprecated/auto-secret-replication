package conc

import (
	"context"
	"sync"
)

// Parallel, create a Task run all tasks in parallel, one failed, others will be cancelled via context
// returns Errs indicating which task failed, which not failed
func Parallel(tasks ...Task) Task {
	return TaskFunc(func(ctx context.Context) error {
		errs := make(Errs, len(tasks), len(tasks))
		ctx2, ctx2Cancel := context.WithCancel(ctx)
		wg := &sync.WaitGroup{}
		for _i, _task := range tasks {
			i, task := _i, _task
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := task.Do(ctx2)
				if err != nil {
					ctx2Cancel()
				}
				errs[i] = err
			}()
		}
		wg.Wait()
		return errs.Sanitize()
	})
}

// ParallelWithLimit, create a Task run all tasks in parallel with limit, one failed, others will be cancelled via context
// returns Errs indicating which task failed, which not failed
func ParallelWithLimit(limit int, tasks ...Task) Task {
	return TaskFunc(func(ctx context.Context) error {
		errs := make(Errs, len(tasks), len(tasks))
		lm := NewFixedLimiter(limit)
		ctx2, ctx2Cancel := context.WithCancel(ctx)
		wg := &sync.WaitGroup{}
		for _i, _task := range tasks {
			i, task := _i, _task
			wg.Add(1)
			go func() {
				defer wg.Done()
				lm.Take()
				defer lm.Done()
				if ctx2.Err() != nil {
					errs[i] = ctx2.Err()
					return
				}
				if err := task.Do(ctx2); err != nil {
					ctx2Cancel()
					errs[i] = err
				}
			}()
		}
		wg.Wait()
		return errs.Sanitize()
	})
}

// ParallelFailSafe, create a Task run all tasks in parallel, one failed, others will still Run
// returns Errs indicating which task failed, which not failed
func ParallelFailSafe(tasks ...Task) Task {
	return TaskFunc(func(ctx context.Context) error {
		errs := make(Errs, len(tasks), len(tasks))
		wg := &sync.WaitGroup{}
		for _i, _task := range tasks {
			i, task := _i, _task
			wg.Add(1)
			go func() {
				defer wg.Done()
				errs[i] = task.Do(ctx)
			}()
		}
		wg.Wait()
		return errs.Sanitize()
	})
}

// ParallelFailSafeWithLimit, create a Task run all tasks in parallel with limit
// using token based concurrency control
// returns Errs indicating which task failed, which not failed
func ParallelFailSafeWithLimit(limit int, tasks ...Task) Task {
	return TaskFunc(func(ctx context.Context) error {
		errs := make(Errs, len(tasks), len(tasks))
		lm := NewFixedLimiter(limit)
		wg := &sync.WaitGroup{}
		for _i, _task := range tasks {
			i, task := _i, _task
			wg.Add(1)
			go func() {
				defer wg.Done()
				lm.Take()
				defer lm.Done()
				if ctx.Err() != nil {
					errs[i] = ctx.Err()
					return
				}
				errs[i] = task.Do(ctx)
			}()
		}
		wg.Wait()
		return errs.Sanitize()
	})
}
