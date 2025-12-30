package models

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type (
	Permission struct {
		bun.BaseModel `bun:"table:permissions,alias:p"`

		ID          int64  `bun:"id,pk,autoincrement" json:"id"`
		Name        string `bun:"name" json:"name"`
		Description string `bun:"description" json:"description"`

		AppModel
	}
)

func (m Permission) Upsert(ctx *gin.Context, item Permission) (int, Permission, error) {
	var oldData *Permission
	httpStatus, action := 201, "POST"

	var setClauseColumns = []string{"name", "description", "updated_at"}

	if item.UUID != "" {
		var tmp Permission
		if err := db.NewSelect().Model(&tmp).Where("uuid = ?", item.UUID).Scan(ctx); err == nil {
			httpStatus, action, oldData = 200, "PUT", &tmp
		}
	}

	setClause := parseSetClause(setClauseColumns)
	err := executeTransaction(ctx, func(trx *bun.Tx) error {
		_, err := trx.NewInsert().Model(&item).On("CONFLICT (uuid) DO UPDATE").Set(setClause).Exec(ctx)
		return err
	})

	go auditLog(ctx, oldData, item, item.ID, "permission", action, err)
	return httpStatus, item, err
}

func (m Permission) Read(qp QueryParams) (res Results, err error) {
	var coalesceCols = []string{"name"}
	var allowedSortFields = map[string]bool{"name": true}

	q := db.NewSelect()

	if qp.UUID != "all" {
		var data Role
		err = q.Model(&data).Where("uuid = ?", qp.UUID).Scan(qp.Ctx)

		res.Item = data

		return res, err
	}

	var data []Role
	q = sanitizeQuery(q.Model(&data), qp, coalesceCols, allowedSortFields)
	res.Count, err = q.ScanAndCount(qp.Ctx)

	for _, item := range data {
		res.Items = append(res.Items, item)
	}

	return res, err
}

func (m Permission) Delete(ctx *gin.Context, uuid string) (deletedAt time.Time, msg string, err error) {
	id, deletedAt, _, msg, err := setStatus(ctx, "permissions", uuid, "deleted_at")

	go auditLog(ctx, nil, map[string]string{"deleted_at": deletedAt.String()}, id, "permission", "DELETE", err)
	return
}

func (m Permission) UpdateStatus(ctx *gin.Context, uuid string) (status, msg string, err error) {
	id, _, status, msg, err := setStatus(ctx, "permissions", uuid, "status")

	go auditLog(ctx, nil, map[string]string{"status": status}, id, "permission", "PATCH", err)
	return
}
