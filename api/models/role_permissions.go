package models

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type (
	RolePermission struct {
		bun.BaseModel `bun:"table:role_permissions,alias:rp"`

		ID           int64 `bun:"id,pk,autoincrement" json:"id"`
		RoleID       int64 `bun:"role_id" json:"role_id"`
		PermissionID int64 `bun:"permission_id" json:"permission_id"`

		AppModel
	}
)

func (m RolePermission) Upsert(ctx *gin.Context, item RolePermission) (int, RolePermission, error) {
	var oldData *RolePermission
	httpStatus, action := 201, "POST"

	var setClauseColumns = []string{"role_id", "permission_id", "updated_at"}

	if item.UUID != "" {
		var tmp RolePermission
		if err := db.NewSelect().Model(&tmp).Where("uuid = ?", item.UUID).Scan(ctx); err == nil {
			httpStatus, action, oldData = 200, "PUT", &tmp
		}
	}

	setClause := parseSetClause(setClauseColumns)
	err := executeTransaction(ctx, func(trx *bun.Tx) error {
		_, err := trx.NewInsert().Model(&item).On("CONFLICT (uuid) DO UPDATE").Set(setClause).Exec(ctx)
		return err
	})

	go auditLog(ctx, oldData, item, item.ID, "role_permission", action, err)
	return httpStatus, item, err
}

func (m RolePermission) Read(qp QueryParams) (res Results, err error) {
	var coalesceCols = []string{}
	var allowedSortFields = map[string]bool{}

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

func (m RolePermission) Delete(ctx *gin.Context, uuid string) (deletedAt time.Time, msg string, err error) {
	id, deletedAt, _, msg, err := setStatus(ctx, "role_permissions", uuid, "deleted_at")

	go auditLog(ctx, nil, map[string]string{"deleted_at": deletedAt.String()}, id, "role_permission", "DELETE", err)
	return
}

func (m RolePermission) UpdateStatus(ctx *gin.Context, uuid string) (status, msg string, err error) {
	id, _, status, msg, err := setStatus(ctx, "role_permissions", uuid, "status")

	go auditLog(ctx, nil, map[string]string{"status": status}, id, "role_permission", "PATCH", err)
	return
}
