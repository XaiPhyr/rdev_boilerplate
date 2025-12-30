package models

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type (
	GroupPermission struct {
		bun.BaseModel `bun:"table:group_permissions,alias:gp"`

		ID           int64 `bun:"id,pk,autoincrement" json:"id"`
		GroupID      int64 `bun:"group_id" json:"group_id"`
		PermissionID int64 `bun:"permission_id" json:"permission_id"`

		AppModel
	}
)

func (m GroupPermission) Upsert(ctx *gin.Context, item GroupPermission) (int, GroupPermission, error) {
	var oldData *GroupPermission
	httpStatus, action := 201, "POST"

	var setClauseColumns = []string{"group_id", "permission_id", "updated_at"}

	if item.UUID != "" {
		var tmp GroupPermission
		if err := db.NewSelect().Model(&tmp).Where("uuid = ?", item.UUID).Scan(ctx); err == nil {
			httpStatus, action, oldData = 200, "PUT", &tmp
		}
	}

	setClause := parseSetClause(setClauseColumns)
	err := executeTransaction(ctx, func(trx *bun.Tx) error {
		_, err := trx.NewInsert().Model(&item).On("CONFLICT (uuid) DO UPDATE").Set(setClause).Exec(ctx)
		return err
	})

	go auditLog(ctx, oldData, item, item.ID, "group_permission", action, err)
	return httpStatus, item, err
}

func (m GroupPermission) Read(qp QueryParams) (res Results, err error) {
	var coalesceCols = []string{}
	var allowedSortFields = map[string]bool{}

	q := db.NewSelect()

	if qp.UUID != "all" {
		var data Group
		err = q.Model(&data).Where("uuid = ?", qp.UUID).Scan(qp.Ctx)

		res.Item = data

		return res, err
	}

	var data []Group
	q = sanitizeQuery(q.Model(&data), qp, coalesceCols, allowedSortFields)
	res.Count, err = q.ScanAndCount(qp.Ctx)

	for _, item := range data {
		res.Items = append(res.Items, item)
	}

	return res, err
}

func (m GroupPermission) Delete(ctx *gin.Context, uuid string) (deletedAt time.Time, msg string, err error) {
	id, deletedAt, _, msg, err := setStatus(ctx, "group_permissions", uuid, "deleted_at")

	go auditLog(ctx, nil, map[string]string{"deleted_at": deletedAt.String()}, id, "group_permission", "DELETE", err)
	return
}

func (m GroupPermission) UpdateStatus(ctx *gin.Context, uuid string) (status, msg string, err error) {
	id, _, status, msg, err := setStatus(ctx, "group_permissions", uuid, "status")

	go auditLog(ctx, nil, map[string]string{"status": status}, id, "group_permission", "PATCH", err)
	return
}
