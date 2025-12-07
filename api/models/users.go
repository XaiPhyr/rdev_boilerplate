package models

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type (
	User struct {
		bun.BaseModel `bun:"table:users,alias:u"`
		AppModel

		Username   string          `bun:"username" json:"username"`
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
	}

	UserResults struct {
		User  User   `json:"user,omitempty"`
		Users []User `json:"users,omitempty"`
		Count int    `json:"count,omitempty"`
	}
)

func (m User) Upsert(ctx *gin.Context, user User) (int, User, error) {
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

	if user.UUID != "" {
		var tmp User
		if err := db.NewSelect().Model(&tmp).Where("uuid = ?", user.UUID).Scan(ctx); err == nil {
			httpStatus, action, oldData = 200, "PUT", &tmp
		}
	}

	setClause := parseSetClause(setClauseColumns)
	err := executeTransaction(ctx, func(trx *bun.Tx) error {
		_, err := trx.NewInsert().
			Model(&user).
			On("CONFLICT (uuid) DO UPDATE").
			Set(setClause).
			Exec(ctx)
		return err
	})

	go auditLog(ctx, oldData, user, user.ID, "user", action, err)
	return httpStatus, user, err
}

func (m User) Read(qp QueryParams) (uuid string, res UserResults, err error) {
	var coalesceColumns = []string{}

	var allowedSortFields = map[string]bool{
		"id":         true,
		"username":   true,
		"email":      true,
		"created_at": true,
	}

	q := db.NewSelect()

	if qp.UUID != "all" {
		return qp.UUID, res, q.Model(&res.User).Where("uuid = ?", qp.UUID).Scan(qp.Ctx, &res.User)
	}

	q = q.Model(&res.Users)
	q = sanitizeQuery(q, coalesceColumns, qp.Filter, qp.Sort, qp.Status, qp.Limit, qp.Page, allowedSortFields)
	q = applyGlobalFilterExt(q, qp.FilterExtOp, qp.FilterExt)

	res.Count, err = q.ScanAndCount(qp.Ctx)
	return qp.UUID, res, err
}

func (m User) Delete(ctx *gin.Context, uuid string) (deletedAt time.Time, msg string, err error) {
	id, deletedAt, msg, err := softDelete(ctx, "users", uuid)

	go auditLog(ctx, nil, map[string]string{"deleted_at": deletedAt.String()}, id, "user", "DELETE", err)
	return
}
