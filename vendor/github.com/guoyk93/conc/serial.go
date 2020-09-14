package conc

import "context"

// Serial run tasks in serial, return on first failed
func Serial(tasks ...Task) Task {
	return TaskFunc(func(ctx context.Context) error {
		errs := make(Errs, len(tasks), len(tasks))
		for i, task := range tasks {
			if err := task.Do(ctx); err != nil {
				errs[i] = err
				break
			}
		}
		return errs.Sanitize()
	})
}

// SerialFailSafe run tasks in serial, one failed will not interfere others
func SerialFailSafe(tasks ...Task) Task {
	return TaskFunc(func(ctx context.Context) error {
		errs := make(Errs, len(tasks), len(tasks))
		for i, task := range tasks {
			errs[i] = task.Do(ctx)
		}
		return errs.Sanitize()
	})
}
