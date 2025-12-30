package models

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type (
	UserRole struct {
		bun.BaseModel `bun:"table:user_roles,alias:ur"`

		ID     int64 `bun:"id,pk,autoincrement" json:"id"`
		UserID int64 `bun:"user_id" json:"user_id"`
		RoleID int64 `bun:"role_id" json:"role_id"`

		AppModel
	}
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

func (m UserRole) Read(qp QueryParams) (res Results, err error) {
	var coalesceCols = []string{}
	var allowedSortFields = map[string]bool{}

	q := db.NewSelect()

	if qp.UUID != "all" {
		var data UserRole
		err = q.Model(&data).Where("uuid = ?", qp.UUID).Scan(qp.Ctx)

		res.Item = data

		return res, err
	}

	var data []UserRole
	q = sanitizeQuery(q.Model(&data), qp, coalesceCols, allowedSortFields)
	res.Count, err = q.ScanAndCount(qp.Ctx)

	for _, item := range data {
		res.Items = append(res.Items, item)
	}

	return res, err
}

func (m UserRole) Delete(ctx *gin.Context, uuid string) (deletedAt time.Time, msg string, err error) {
	id, deletedAt, _, msg, err := setStatus(ctx, "user_roles", uuid, "deleted_at")

	go auditLog(ctx, nil, map[string]string{"deleted_at": deletedAt.String()}, id, "user_role", "DELETE", err)
	return
}

func (m UserRole) UpdateStatus(ctx *gin.Context, uuid string) (status, msg string, err error) {
	id, _, status, msg, err := setStatus(ctx, "user_roles", uuid, "status")

	go auditLog(ctx, nil, map[string]string{"status": status}, id, "user_role", "PATCH", err)
	return
}
