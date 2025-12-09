package models

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func (m UserRole) Upsert(ctx *gin.Context, item UserRole) (int, UserRole, error) {
	var oldData *UserRole
	httpStatus, action := 201, "POST"

	var setClauseColumns = []string{"user_id", "role_id", "updated_at"}

	if item.UUID != "" {
		var tmp UserRole
		if err := db.NewSelect().Model(&tmp).Where("uuid = ?", item.UUID).Scan(ctx); err == nil {
			httpStatus, action, oldData = 200, "PUT", &tmp
		}
	}

	setClause := parseSetClause(setClauseColumns)
	err := executeTransaction(ctx, func(trx *bun.Tx) error {
		_, err := trx.NewInsert().Model(&item).On("CONFLICT (uuid) DO UPDATE").Set(setClause).Exec(ctx)
		return err
	})

	go auditLog(ctx, oldData, item, item.ID, "user_role", action, err)
	return httpStatus, item, err
}

func (m UserRole) Read(qp QueryParams) (res UserRoleResults, err error) {
	var coalesceCols = []string{}
	var allowedSortFields = map[string]bool{}

	q := db.NewSelect()

	if qp.UUID != "all" {
		return res, q.Model(&res.UserRole).Where("uuid = ?", qp.UUID).Scan(qp.Ctx)
	}

	q = sanitizeQuery(q.Model(&res.UserRoles), qp, coalesceCols, allowedSortFields)
	res.Count, err = q.ScanAndCount(qp.Ctx)
	return res, err
}

func (m UserRole) Delete(ctx *gin.Context, uuid string) (deletedAt time.Time, msg string, err error) {
	id, deletedAt, msg, err := softDelete(ctx, "user_roles", uuid)

	go auditLog(ctx, nil, map[string]string{"deleted_at": deletedAt.String()}, id, "user_role", "DELETE", err)
	return
}
