package activity

import (
	"context"
	"database/sql"

	"github.com/anousonefs/golang-htmx-template/internal/config"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r Repo) createActivity(ctx context.Context, req Activity) error {
	query, args, err := config.Psql().
		Insert("activities").
		Columns(
			"title",
			"resource",
			"action",
			"req_data",
			"res_data",
			"created_by",
		).
		Values(
			req.Title,
			req.Resource,
			req.Action,
			req.ReqData,
			req.ResData,
			req.CreatedBy,
		).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (r Repo) listActivities(ctx context.Context, req FilterActivity) (res ActivityList, err error) {
	query, args := config.Psql().
		Select(
			"title",
			"resource",
			"action",
			"req_data",
			"res_data",
			"department_id",
			"created_by",
			"created_at",
		).From("activities").Where(req).MustSql()
	_, err = r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return res, err
	}
	return res, nil
}
