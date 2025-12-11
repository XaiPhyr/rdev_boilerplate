package models

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type (
	UserResults struct {
		User  User
		Users []User
		Count int
	}

	User struct {
		bun.BaseModel `bun:"table:users,alias:u"`

		ID         int64           `bun:"id,pk,autoincrement" json:"id"`
		Username   string          `bun:"username,unique" json:"username"`
		Password   string          `bun:"password" json:"-"`
		Email      string          `bun:"email,unique" json:"email"`
		FirstName  *string         `bun:"first_name,nullzero,default:null" json:"first_name"`
		MiddleName *string         `bun:"middle_name,nullzero,default:null" json:"middle_name"`
		LastName   *string         `bun:"last_name,nullzero,default:null" json:"last_name"`
		Mobile     *string         `bun:"mobile,unique,nullzero,default:null" json:"mobile"`
		Address    *map[string]any `bun:"address,type:jsonb" json:"address"`
		Optin      bool            `bun:"optin,default:false" json:"optin"`
		LastLogin  time.Time       `bun:"last_login,nullzero,default:null" json:"last_login,omitzero"`
		Metadata   *map[string]any `bun:"metadata,type:jsonb,default:null" json:"metadata"`
		IsAdmin    bool            `bun:"is_admin" json:"is_admin"`
		IsOnline   bool            `bun:"is_online" json:"is_online"`

		Permissions []string `bun:"-" json:"permissions,omitempty"`
		AppModel
	}
)

func (m User) Upsert(ctx *gin.Context, item User) (int, User, error) {
	var oldData *User
	httpStatus, action := 201, "POST"

	var setClauseColumns = []string{
		"username",
		"email",
		"first_name",
		"middle_name",
		"last_name",
		"mobile",
		"address",
		"optin",
		"metadata",
		"is_admin",
		"updated_at",
	}

	if item.UUID != "" {
		var tmp User
		if err := db.NewSelect().Model(&tmp).Where("uuid = ?", item.UUID).Scan(ctx); err == nil {
			httpStatus, action, oldData = 200, "PUT", &tmp
		}
	}

	setClause := parseSetClause(setClauseColumns)
	err := executeTransaction(ctx, func(trx *bun.Tx) error {
		_, err := trx.NewInsert().Model(&item).On("CONFLICT (uuid) DO UPDATE").Set(setClause).Exec(ctx)
		return err
	})

	go auditLog(ctx, oldData, item, item.ID, "user", action, err)
	return httpStatus, item, err
}

func (m User) Read(qp QueryParams) (res UserResults, err error) {
	var coalesceCols = []string{"username", "first_name", "middle_name", "last_name"}
	var allowedSortFields = map[string]bool{"username": true, "email": true}

	q := db.NewSelect()

	if qp.UUID != "all" {
		err = q.Model(&res.User).Where("u.uuid = ?", qp.UUID).Scan(qp.Ctx)

		if permissions, err := getPermissions(qp.Ctx, qp.UUID); err == nil {
			res.User.Permissions = permissions
		}

		return res, err
	}

	q = sanitizeQuery(q.Model(&res.Users), qp, coalesceCols, allowedSortFields)
	res.Count, err = q.ScanAndCount(qp.Ctx)
	return res, err
}

func (m User) Delete(ctx *gin.Context, uuid string) (deletedAt time.Time, msg string, err error) {
	id, deletedAt, _, msg, err := setStatus(ctx, "users", uuid, "deleted_at")

	go auditLog(ctx, nil, map[string]string{"deleted_at": deletedAt.String()}, id, "user", "DELETE", err)
	return
}

func (m User) UpdateStatus(ctx *gin.Context, uuid string) (status, msg string, err error) {
	id, _, status, msg, err := setStatus(ctx, "users", uuid, "status")

	go auditLog(ctx, nil, map[string]string{"status": status}, id, "user", "PATCH", err)
	return
}
