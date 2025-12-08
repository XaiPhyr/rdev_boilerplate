package models

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
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
		_, err := trx.NewInsert().Model(&user).On("CONFLICT (uuid) DO UPDATE").Set(setClause).Exec(ctx)
		return err
	})

	go auditLog(ctx, oldData, user, user.ID, "user", action, err)
	return httpStatus, user, err
}

func (m User) Read(qp QueryParams) (res UserResults, err error) {
	var coalesceCols = []string{"username", "first_name", "middle_name", "last_name"}
	var allowedSortFields = map[string]bool{"id": true, "username": true, "email": true, "created_at": true}

	q := db.NewSelect()

	if qp.UUID != "all" {
		return res, q.Model(&res.User).Where("uuid = ?", qp.UUID).Scan(qp.Ctx)
	}

	q = sanitizeQuery(q.Model(&res.Users), qp, coalesceCols, allowedSortFields)
	res.Count, err = q.ScanAndCount(qp.Ctx)
	return res, err
}

func (m User) Delete(ctx *gin.Context, uuid string) (deletedAt time.Time, msg string, err error) {
	id, deletedAt, msg, err := softDelete(ctx, "users", uuid)

	go auditLog(ctx, nil, map[string]string{"deleted_at": deletedAt.String()}, id, "user", "DELETE", err)
	return
}
