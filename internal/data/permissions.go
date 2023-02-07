package data

import (
	"context"
	"time"
)

type Permissions []string

func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

type PermissionModel struct {
	DB DB
}

func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := m.DB.Get(ctx, "", userID, "permissions", "")
	if err != nil {
		return nil, err
	}

	return res.(Permissions), nil
}

type UserPermissionsSend struct {
	User_Id int64    `json:"user_id"`
	Codes   []string `json:"codes"`
}

func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.Insert(ctx, "", "user_permissions", UserPermissionsSend{User_Id: userID, Codes: codes})

}
