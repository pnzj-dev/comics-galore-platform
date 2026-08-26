package comics

import (
	"context"

	myjobs "comics-galore/backend/jobs"
)

// runTrackedJob records a job run in the jobs service around fn. Recording is
// best-effort: a failure to record must never block the underlying job.
func runTrackedJob(ctx context.Context, name, ref string, fn func() error) error {
	run, _ := myjobs.RecordJobStart(ctx, &myjobs.RecordJobStartParams{JobName: name, Ref: ref})
	err := fn()
	if run != nil {
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		_ = myjobs.RecordJobFinish(ctx, &myjobs.RecordJobFinishParams{ID: run.ID, Error: errStr})
	}
	return err
}
