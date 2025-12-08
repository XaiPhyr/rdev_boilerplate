package models

import (
	"api/utils"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

var db = utils.InitDB()

func sanitizeQuery(q *bun.SelectQuery, qp QueryParams, cols []string, allowedSortFields map[string]bool) *bun.SelectQuery {
	if len(cols) > 0 {
		var cls []string
		for _, col := range cols {
			cls = append(cls, "coalesce("+col+",'')")
		}

		qp.Filter = strings.ToLower(qp.Filter)
		cl := "(" + strings.Join(cls, " || ") + ")"
		if qp.Filter != "" {
			q = q.Where(cl+" ~* ?", qp.Filter)
		}
	}

	if qp.Sort != "" {
		for sortField := range strings.SplitSeq(qp.Sort, ",") {
			if validateField(allowedSortFields, sortField) {
				if after, ok := strings.CutPrefix(sortField, "-"); ok {
					q = q.Order(after + " DESC")
				} else {
					q = q.Order(sortField + " ASC")
				}
			}
		}
	} else {
		q = q.Order("id ASC")
	}

	if qp.Status != "" {
		q = q.Where("status = ?", qp.Status)
	}

	if qp.Limit > 0 {
		q = q.Limit(qp.Limit)

		if qp.Page > 0 {
			offset := (qp.Page - 1) * qp.Limit
			q = q.Offset(offset)
		}
	}

	if qp.FilterExt != "" {
		filters := strings.Split(qp.FilterExt, ",")
		for _, f := range filters {
			if len(strings.Split(f, "=")) == 2 {
				q = buildFilterExtQuery(q, filters, qp.FilterExtOp, strings.Split(f, "="))
			}
		}
	}

	return q
}

func buildFilterExtQuery(q *bun.SelectQuery, filters []string, filterExtOp string, kv []string) *bun.SelectQuery {
	operator := "where"
	if len(filters) > 1 && filterExtOp == "OR" {
		operator = "whereor"
	}

	identifier := kv[0] + " = ?"
	var literal any = kv[1]

	if strings.Contains(kv[1], "||") {
		identifier = kv[0] + " IN (?)"
		literal = bun.In(strings.Split(kv[1], "||"))
	}

	if operator == "where" {
		return q.Where(identifier, literal)
	}

	return q.WhereOr(identifier, literal)
}

func executeTransaction(ctx context.Context, trxFunc func(*bun.Tx) error) error {
	trx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	if err := trxFunc(&trx); err != nil {
		trx.Rollback()
		return err
	}

	return trx.Commit()
}

func auditLog(ctx *gin.Context, beforeDataChange, afterDataChange any, moduleId int64, module, action string, err error) {
	v, _ := ctx.Get("userId")

	var userID int64
	id, ok := v.(int)
	if ok {
		userID = int64(id)
	}

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	auditLog := &AuditLog{
		UserID:           userID,
		Token:            "",
		Path:             ctx.FullPath(),
		Action:           action,
		ResponseStatus:   ctx.Writer.Status(),
		ModuleID:         moduleId,
		Module:           module,
		BeforeDataChange: beforeDataChange,
		AfterDataChange:  afterDataChange,
		Description:      errMsg,
		IPAddress:        ctx.ClientIP(),
		UserAgent:        ctx.Request.UserAgent(),
	}

	_, dbErr := db.NewInsert().Model(auditLog).Exec(ctx)

	if dbErr != nil {
		fmt.Println("AUDIT LOG ERR: ", dbErr)
	}
}

func softDelete(ctx *gin.Context, tableName, uuid string) (id int64, deletedAt time.Time, msg string, err error) {
	var temp struct {
		ID        int64     `bun:"id"`
		DeletedAt time.Time `bun:"deleted_at" json:"deleted_at"`
	}

	err = executeTransaction(ctx, func(trx *bun.Tx) error {
		_, err := trx.NewUpdate().
			Table(tableName).
			Where("uuid = ?", uuid).
			Set("deleted_at = CASE WHEN deleted_at IS NULL THEN NOW() ELSE NULL END").
			Returning("deleted_at, id").
			Exec(ctx, &temp)
		return err
	})

	id = temp.ID
	msg = "restored successfully"
	if !temp.DeletedAt.IsZero() {
		deletedAt = temp.DeletedAt
		msg = "deleted successfully"
	}

	return
}

func parseSetClause(cols []string) string {
	setClauses := make([]string, 0, len(cols))
	for _, col := range cols {
		setClauses = append(setClauses, col+" = EXCLUDED."+col)
	}

	return strings.Join(setClauses, ", ")
}

func validateField(allowedSortFields map[string]bool, sortField string) bool {
	after, _ := strings.CutPrefix(sortField, "-")
	return allowedSortFields[after]
}
