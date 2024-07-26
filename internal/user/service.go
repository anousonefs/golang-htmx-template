package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/anousonefs/golang-htmx-template/internal/activity"
	"github.com/anousonefs/golang-htmx-template/internal/utils"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Service struct {
	repo     *Repo
	activity *activity.Service
}

func NewService(repo *Repo, activity *activity.Service) Service {
	return Service{repo, activity}
}

func (u *Service) CreateUser(ctx context.Context, req User, act activity.Activity) (err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("user.CreateUser(): %v\n", err)
		}
	}()
	req.Status = UserStatusActive
	utils.PrettyPrint(req)
	if err = u.repo.createUser(ctx, req); err != nil {
		fmt.Printf("err: %v\n", err)
		pgErr, isPGErr := err.(*pq.Error)
		if isPGErr && pgErr.Code == "23505" {
			return ErrDuplicateKey
		}
		return err
	}

	act.Title = "Create User"
	act.Resource = "user"
	act.Action = "create"
	act.ResData = []byte(fmt.Sprintf("%+v", req))
	act.ResData = []byte(fmt.Sprintf("%+v", req))
	if err := u.activity.CreateActivity(ctx, act); err != nil {
		return err
	}

	return nil
}

func (u *Service) ListUsers(ctx context.Context, filter FilterUser) (res []UserList, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("user.ListUsers(): %v\n", err)
		}
	}()
	res, err = u.repo.listUsers(ctx, filter)
	if err != nil {
		return []UserList{}, err
	}
	return res, nil
}

func (u *Service) GetUser(ctx context.Context, filter FilterUser) (res *UserDetail, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("user.GetUser(%#v): %v\n", filter, err)
		}
	}()
	res, err = u.repo.getUser(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStatusNotFound
		}
		return nil, err
	}
	return res, err
}

func (u *Service) CreatePermission(ctx context.Context, req Permission) (res []ListPermission, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("u.CreatePermission(): %v\n", err)
		}
	}()
	utils.PrettyPrint(req)
	if _, err = u.GetRole(ctx, FilterRole{ID: req.RoleID}); err != nil {
		return nil, err
	}
	println("pass")
	if err = u.repo.createPermission(ctx, req); err != nil {
		return nil, err
	}
	return u.ListPermissions(ctx, req.RoleID)
}

func (u *Service) ListRoles(ctx context.Context) (res []Role, err error) {
	res, err = u.repo.listRoles(ctx)
	if err != nil {
		logrus.Errorf("u.ListRoles(): %v\n", err)
		return []Role{}, err
	}
	return res, nil
}

func (u *Service) GetRole(ctx context.Context, filter FilterRole) (res Role, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("u.GetRoles(): %v\n", err)
		}
	}()
	utils.PrettyPrint(filter)
	res, err = u.repo.getRole(ctx, filter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Role{}, ErrStatusNotFound
		}
		return Role{}, err
	}
	return res, nil
}

func (u *Service) ListPermissions(ctx context.Context, roleID string) (res []ListPermission, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("u.ListPermissions(): %v\n", err)
		}
	}()
	res, err = u.repo.listPermissions(ctx, roleID)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, err
		}
		return nil, err
	}
	return res, nil
}

func (u *Service) GetPermissions(ctx context.Context, filter FilterPermission) (res ListPermission, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("u.GetPermissions(): %v\n", err)
		}
	}()
	res, err = u.repo.getPermissions(ctx, filter)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return ListPermission{}, ErrStatusNotFound
		}
		return ListPermission{}, err
	}
	return res, nil
}

func (u *Service) ListAllPermissions(ctx context.Context) (res []AllPermission, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("u.ListPermissions(): %v\n", err)
		}
	}()
	res, err = u.repo.listAllPermissions(ctx)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, ErrStatusNotFound
		}
		return nil, err
	}
	return res, nil
}
