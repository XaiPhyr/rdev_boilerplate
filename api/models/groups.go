package models

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func (m Group) Upsert(ctx *gin.Context, item Group) (int, Group, error) {
	var oldData *Group
	httpStatus, action := 201, "POST"

	var setClauseColumns = []string{"name", "description", "updated_at"}

	if item.UUID != "" {
		var tmp Group
		if err := db.NewSelect().Model(&tmp).Where("uuid = ?", item.UUID).Scan(ctx); err == nil {
			httpStatus, action, oldData = 200, "PUT", &tmp
		}
	}

	setClause := parseSetClause(setClauseColumns)
	err := executeTransaction(ctx, func(trx *bun.Tx) error {
		_, err := trx.NewInsert().Model(&item).On("CONFLICT (uuid) DO UPDATE").Set(setClause).Exec(ctx)
		return err
	})

	go auditLog(ctx, oldData, item, item.ID, "group", action, err)
	return httpStatus, item, err
}

func (m Group) Read(qp QueryParams) (res GroupResults, err error) {
	var coalesceCols = []string{"name"}
	var allowedSortFields = map[string]bool{"name": true}

	q := db.NewSelect()

	if qp.UUID != "all" {
		return res, q.Model(&res.Group).Where("uuid = ?", qp.UUID).Scan(qp.Ctx)
	}

	q = sanitizeQuery(q.Model(&res.Groups), qp, coalesceCols, allowedSortFields)
	res.Count, err = q.ScanAndCount(qp.Ctx)
	return res, err
}

func (m Group) Delete(ctx *gin.Context, uuid string) (deletedAt time.Time, msg string, err error) {
	id, deletedAt, msg, err := softDelete(ctx, "groups", uuid)

	go auditLog(ctx, nil, map[string]string{"deleted_at": deletedAt.String()}, id, "group", "DELETE", err)
	return
}

func (m Group) UpdateStatus(ctx *gin.Context, uuid string) (status, msg string, err error) {
	id, status, msg, err := updateStatus(ctx, "groups", uuid)

	go auditLog(ctx, nil, map[string]string{"status": status}, id, "group", "PATCH", err)
	return
}
