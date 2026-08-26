// Package jobs provides an app-level observability layer for background work
// (cron jobs and pubsub handlers). Encore's dead-letter queue is internal and
// not queryable from app code, so jobs record their own runs here; the admin
// dashboard reads this table to surface failures and manual re-runs.
package jobs

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	myauth "comics-galore/backend/auth"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("jobsdb", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

type JobRun struct {
	ID         string     `json:"id"`
	JobName    string     `json:"job_name"`
	Ref        string     `json:"ref"`
	Status     string     `json:"status"`
	Error      string     `json:"error"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

type RecordJobStartParams struct {
	JobName string `json:"job_name"`
	Ref     string `json:"ref"`
}

type RecordJobStartResponse struct {
	ID string `json:"id"`
}

// RecordJobStart marks the beginning of a background job run and returns the
// run id, to be passed to RecordJobFinish when the job completes.
//encore:api private method=POST path=/jobs/record-start
func RecordJobStart(ctx context.Context, p *RecordJobStartParams) (*RecordJobStartResponse, error) {
	var id string
	err := db.QueryRow(ctx, `
		INSERT INTO job_runs (job_name, ref, status)
		VALUES ($1, $2, 'running')
		RETURNING id
	`, p.JobName, p.Ref).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &RecordJobStartResponse{ID: id}, nil
}

type RecordJobFinishParams struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

// RecordJobFinish marks a job run as succeeded or failed.
//encore:api private method=POST path=/jobs/record-finish
func RecordJobFinish(ctx context.Context, p *RecordJobFinishParams) error {
	status := "success"
	var errVal interface{}
	if p.Error != "" {
		status = "failed"
		errVal = p.Error
	}
	_, err := db.Exec(ctx, `
		UPDATE job_runs SET status = $1, error_message = $2, finished_at = now()
		WHERE id = $3
	`, status, errVal, p.ID)
	return err
}

type ListJobRunsParams struct {
	JobName string `query:"job_name"`
	Status  string `query:"status"`
	Limit   int    `query:"limit"`
}

type ListJobRunsResponse struct {
	Runs []JobRun `json:"runs"`
}

// ListJobRuns lists recent job runs for the admin dashboard.
//encore:api auth method=GET path=/admin/jobs
func ListJobRuns(ctx context.Context, p *ListJobRunsParams) (*ListJobRunsResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}

	limit := p.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	query := `SELECT id, job_name, ref, status, COALESCE(error_message, ''), started_at, finished_at FROM job_runs`
	var args []interface{}
	argIdx := 1
	if p.JobName != "" || p.Status != "" {
		query += " WHERE"
		if p.JobName != "" {
			query += " job_name = $" + strconv.Itoa(argIdx)
			args = append(args, p.JobName)
			argIdx++
		}
		if p.Status != "" {
			if p.JobName != "" {
				query += " AND"
			}
			query += " status = $" + strconv.Itoa(argIdx)
			args = append(args, p.Status)
			argIdx++
		}
	}
	query += " ORDER BY started_at DESC LIMIT $" + strconv.Itoa(argIdx)
	args = append(args, limit)

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := []JobRun{}
	for rows.Next() {
		var r JobRun
		var finished sql.NullTime
		if err := rows.Scan(&r.ID, &r.JobName, &r.Ref, &r.Status, &r.Error, &r.StartedAt, &finished); err != nil {
			return nil, err
		}
		if finished.Valid {
			r.FinishedAt = &finished.Time
		}
		runs = append(runs, r)
	}
	if runs == nil {
		runs = []JobRun{}
	}
	return &ListJobRunsResponse{Runs: runs}, rows.Err()
}
