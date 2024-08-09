package user

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/anousonefs/golang-htmx-template/internal/config"
	"github.com/anousonefs/golang-htmx-template/internal/middleware"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	casbinpgadapter "github.com/cychiuae/casbin-pg-adapter"
)

type Repo struct {
	db      *sql.DB
	authz   *middleware.CasbinMiddleware
	adapter *casbinpgadapter.Adapter
	model   string
}

func NewRepo(db *sql.DB, model string, adapter *casbinpgadapter.Adapter, authz *middleware.CasbinMiddleware) *Repo {
	return &Repo{
		db:      db,
		adapter: adapter,
		model:   model,
		authz:   authz,
	}
}

func (r Repo) listUsers(ctx context.Context, filter FilterUser) ([]UserList, error) {
	query, args := config.Psql().
		Select(
			"u.id",
			"u.first_name",
			"u.last_name",
			"u.role_id",
			"u.updated_at",
			"u.department_id",
			"u.position_id",
			"u.gender",
			"u.is_signer",
			"u.email",
			"u.phone",
			"u.status",
			"u.signature",
			"u.avatar",
			"u.cif",
			"u.created_at",
			"u.created_by",
			"u.updated_by",
		).From("users u").
		LeftJoin("roles r ON r.id = u.role_id").
		Where(filter).
		MustSql()
	fmt.Printf("query: %v\n", query)

	rows, err := r.db.QueryContext(ctx, query, args...)
	defer rows.Close()
	res := []UserList{}
	for rows.Next() {
		var i UserList
		if err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.RoleID,
			&i.UpdatedAt,
			&i.DepartmentID,
			&i.PositionID,
			&i.Gender,
			&i.IsSigner,
			&i.Email,
			&i.Phone,
			&i.Status,
			&i.Signature,
			&i.Avatar,
			&i.Cif,
			&i.CreatedAt,
			&i.CreatedBy,
			&i.UpdatedBy,
		); err != nil {
			return []UserList{}, err
		}
		res = append(res, i)
	}
	if err != nil {
		return []UserList{}, err
	}
	return res, nil
}

func (r Repo) createUser(ctx context.Context, req User) error {
	query, args, err := config.Psql().
		Insert("users").
		Columns(
			"role_id",
			"first_name",
			"last_name",
			"status",
			"password",
			"gender",
			"phone",
			"email",
			"department_id",
			"position_id",
			"created_by",
			"updated_by",
		).
		Values(
			req.RoleID,
			req.FirstName,
			req.LastName,
			req.Status,
			req.Password,
			req.Gender,
			req.Phone,
			req.Email,
			req.DepartmentID,
			req.PositionID,
			req.CreatedBy,
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

func (r Repo) getUser(ctx context.Context, filter FilterUser) (res *UserDetail, err error) {
	query, args := config.Psql().
		Select(
			"u.id",
			"r.id",
			"r.code",
			"u.first_name",
			"u.last_name",
			"u.gender",
			"u.email",
			"u.phone",
			"u.status",
			"u.password",
			"u.created_at",
			"u.created_by",
			"u.updated_at",
			"u.updated_by",
		).
		From("users u").
		LeftJoin("roles r ON r.id = u.role_id").
		Where(filter).MustSql()
	fmt.Printf("query: %v\n", query)
	var i UserDetail
	row := r.db.QueryRowContext(ctx, query, args...)
	if err := row.Scan(
		&i.ID,
		&i.Role.ID,
		&i.Role.Name,
		&i.FirstName,
		&i.LastName,
		&i.Gender,
		&i.Email,
		&i.Phone,
		&i.Status,
		&i.Password,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	); err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *Repo) createPermission(_ context.Context, req Permission) error {
	var role string = req.RoleID
	m, _ := model.NewModelFromString(r.model)
	e, err := casbin.NewEnforcer(m, r.adapter)
	if err != nil {
		return err
	}
	if _, err := e.RemovePolicy(role); err != nil {
		return err
	}
	if err = e.LoadPolicy(); err != nil {
		return err
	}
	if !e.HasPolicy(role) {
		for i := 0; i < len(req.User); i++ {
			if _, err := e.AddPolicy(role, "user", req.User[i]); err != nil {
				return err
			}
		}
	}

	r.authz.ReloadEnforcer(r.model, r.adapter)

	return nil
}

func (r *Repo) getRole(ctx context.Context, filter FilterRole) (Role, error) {
	query, args := config.Psql().
		Select(
			"id",
			"name",
			"status",
			"created_at",
		).
		From("roles").
		Where(filter).
		MustSql()
	row := r.db.QueryRowContext(ctx, query, args...)
	var i Role
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Status,
		&i.CreatedAt,
	)
	return i, err
}

func (r *Repo) listRoles(ctx context.Context) ([]Role, error) {
	query, args, err := config.Psql().
		Select(
			"id",
			"name",
			"status",
			"created_at",
		).From("roles").
		ToSql()
	if err != nil {
		return []Role{}, err
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Role{}
	for rows.Next() {
		var i Role
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Status,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repo) listPermissions(ctx context.Context, roleID string) ([]ListPermission, error) {
	query, args, err := config.Psql().
		Select(
			"v1",
			"v2",
		).From("permissions").
		Where("v0 = ?", roleID).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListPermission{}
	for rows.Next() {
		var i ListPermission
		if err := rows.Scan(
			&i.Domain,
			&i.Action,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repo) getPermissions(ctx context.Context, filter FilterPermission) (ListPermission, error) {
	query, args := config.Psql().
		Select(
			"v1",
			"v2",
		).From("permissions").
		Where(filter).MustSql()
	row := r.db.QueryRowContext(ctx, query, args...)
	var i ListPermission
	if err := row.Scan(
		&i.Domain,
		&i.Action,
	); err != nil {
		return ListPermission{}, err
	}
	return i, nil
}

func (r *Repo) listAllPermissions(ctx context.Context) ([]AllPermission, error) {
	query, args, err := config.Psql().
		Select(
			"resource",
			"action",
			"created_at",
		).From("all_permissions").
		ToSql()
	if err != nil {
		return []AllPermission{}, err
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []AllPermission{}
	for rows.Next() {
		var i AllPermission
		if err := rows.Scan(
			&i.Resource,
			&i.Action,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}
