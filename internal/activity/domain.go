package activity

import "github.com/Masterminds/squirrel"

type Activity struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Resource     string `json:"resource"`
	Action       string `json:"action"`
	ReqData      []byte `json:"reqData"`
	ResData      []byte `json:"resData"`
	DepartmentID string `json:"departmentID"`
	CreatedBy    string `json:"createdBy"`
	CreatedAt    string `json:"createdAt"`
}

type ActivityList []Activity

type FilterActivity struct {
}

func (r FilterActivity) ToSql() (string, []interface{}, error) {
	eq := squirrel.Eq{}
	return eq.ToSql()
}
