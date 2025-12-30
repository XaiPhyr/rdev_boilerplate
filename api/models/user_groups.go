package models

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type (
	UserGroup struct {
		bun.BaseModel `bun:"table:user_groups,alias:ug"`

		ID      int64 `bun:"id,pk,autoincrement" json:"id"`
		UserID  int64 `bun:"user_id" json:"user_id"`
		GroupID int64 `bun:"group_id" json:"group_id"`

		AppModel
	}
)

func (m UserGroup) Upsert(ctx *gin.Context, item UserGroup) (int, UserGroup, error) {
	var oldData *UserGroup
	httpStatus, action := 201, "POST"

	var setClauseColumns = []string{"user_id", "group_id", "updated_at"}

	if item.UUID != "" {
		var tmp UserGroup
		if err := db.NewSelect().Model(&tmp).Where("uuid = ?", item.UUID).Scan(ctx); err == nil {
			httpStatus, action, oldData = 200, "PUT", &tmp
		}
	}

	setClause := parseSetClause(setClauseColumns)
	err := executeTransaction(ctx, func(trx *bun.Tx) error {
		_, err := trx.NewInsert().Model(&item).On("CONFLICT (uuid) DO UPDATE").Set(setClause).Exec(ctx)
		return err
	})

	go auditLog(ctx, oldData, item, item.ID, "user_group", action, err)
	return httpStatus, item, err
}

func (m UserGroup) Read(qp QueryParams) (res Results, err error) {
	var coalesceCols = []string{}
	var allowedSortFields = map[string]bool{}

	q := db.NewSelect()

	if qp.UUID != "all" {
		var data UserGroup
		err = q.Model(&data).Where("uuid = ?", qp.UUID).Scan(qp.Ctx)

		res.Item = data

		return res, err
	}

	var data []UserGroup
	q = sanitizeQuery(q.Model(&data), qp, coalesceCols, allowedSortFields)
	res.Count, err = q.ScanAndCount(qp.Ctx)

	for _, item := range data {
		res.Items = append(res.Items, item)
	}

	return res, err
}

func (m UserGroup) Delete(ctx *gin.Context, uuid string) (deletedAt time.Time, msg string, err error) {
	id, deletedAt, _, msg, err := setStatus(ctx, "user_groups", uuid, "deleted_at")

	go auditLog(ctx, nil, map[string]string{"deleted_at": deletedAt.String()}, id, "user_group", "DELETE", err)
	return
}

func (m UserGroup) UpdateStatus(ctx *gin.Context, uuid string) (status, msg string, err error) {
	id, _, status, msg, err := setStatus(ctx, "user_groups", uuid, "status")

	go auditLog(ctx, nil, map[string]string{"status": status}, id, "user_group", "PATCH", err)
	return
}
