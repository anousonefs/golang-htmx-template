package user

import (
	"time"

	"github.com/Masterminds/squirrel"
)

const partSize = 10 * 1024 * 1024 // 10 mb

type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInActive UserStatus = "INACTIVE"
)

type User struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	RoleID       string     `json:"roleID"`
	TenantID     string     `json:"tenantID"`
	Status       UserStatus `json:"status"`
	FirstName    string     `json:"firstname"`
	LastName     string     `json:"lastname"`
	Gender       string     `json:"gender"`
	Avatar       string     `json:"avatar"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	DepartmentID string     `json:"departmentID"`
	PositionID   string     `json:"positionID"`
	Password     string     `json:"password,omitempty"`
	CreatedBy    string     `json:"createdBy"`
}

type Genders string

const (
	GendersM Genders = "M"
	GendersF Genders = "F"
	GendersO Genders = "O"
)

type UserList struct {
	ID           string     `json:"id"`
	FirstName    string     `json:"firstName"`
	LastName     string     `json:"lastName"`
	RoleID       string     `json:"roleID"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DepartmentID string     `json:"departmentID"`
	PositionID   string     `json:"positionID"`
	Gender       Genders    `json:"gender"`
	IsSigner     bool       `json:"isSigner"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	Status       UserStatus `json:"status"`
	Signature    string     `json:"signature"`
	Avatar       string     `json:"avatar"`
	Cif          string     `json:"cif"`
	CreatedAt    time.Time  `json:"createdAt"`
	CreatedBy    string     `json:"createdBy"`
	UpdatedBy    string     `json:"updatedBy"`
}

type UserDetail struct {
	ID           string `json:"id"`
	DepartmentID string `json:"departmentID"`
	Role         struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	Status    UserStatus `json:"status"`
	FirstName string     `json:"firstname"`
	LastName  string     `json:"lastname"`
	Gender    string     `json:"gender"`
	Avatar    string     `json:"avatar"`
	Email     string     `json:"email"`
	Phone     string     `json:"phone"`
	Password  string     `json:"-"`
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy string     `json:"-"`
	UpdatedAt time.Time  `json:"updatedAt"`
	UpdatedBy string     `json:"updatedBy"`
}

func (f User) Validate() error {
	if f.Username == "" || f.FirstName == "" || f.LastName == "" || f.Password == "" || f.Phone == "" || f.Email == "" {
		return ErrBadRequest
	}
	return nil
}

type FilterUser struct {
	ID       string
	Username string
	Email    string
	Phone    string
}

func (f FilterUser) ToSql() (string, []interface{}, error) {
	eq := squirrel.Eq{}
	if f.ID != "" {
		eq["u.id"] = f.ID
	}
	if f.Username != "" {
		eq["u.username"] = f.Username
	}
	if f.Phone != "" {
		eq["u.phone"] = f.Phone
	}
	if f.Email != "" {
		eq["u.email"] = f.Email
	}
	return eq.ToSql()
}

type Role struct {
	ID        *string   `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type RoleObj struct {
	ID     *string `json:"id"`
	Name   *string `json:"name"`
	Status string  `json:"status"`
}

type ListPermission struct {
	Domain string `json:"domain"`
	Action string `json:"action"`
}

type PositionObj struct {
	ID     *string `json:"id"`
	Name   *string `json:"name"`
	Status string  `json:"status"`
}

type Permission struct {
	RoleID string   `json:"roleID"`
	User   []string `json:"user"`
}

type AllPermission struct {
	Resource  string    `json:"domain"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"createdAt"`
}

type FilterPermission struct {
	RoleID   string
	Resource string
	Action   string
}

func (f FilterPermission) ToSql() (string, []interface{}, error) {
	eq := squirrel.Eq{}
	if f.RoleID != "" {
		eq["v0"] = f.RoleID
	}
	if f.Resource != "" {
		eq["v1"] = f.Resource
	}
	if f.Action != "" {
		eq["v2"] = f.Action
	}
	return eq.ToSql()
}

type FilterRole struct {
	ID   string
	Name string
}

func (f FilterRole) ToSql() (string, []interface{}, error) {
	eq := squirrel.Eq{}
	if f.ID != "" {
		eq["id"] = f.ID
	}
	if f.Name != "" {
		eq["name"] = f.Name
	}
	return eq.ToSql()
}
