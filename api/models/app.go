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

type (
	AppModel struct {
		Active    bool      `bun:"active,default:true" json:"active"`
		Status    string    `bun:"status,default:O" json:"status"`
		Flag      string    `bun:"flag,nullzero,default:null" json:"flag,omitempty"`
		UUID      string    `bun:"uuid,default:gen_random_uuid()" json:"uuid,omitempty"`
		CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at,omitzero"`
		CreatedBy int64     `bun:"created_by,default:0" json:"created_by,omitzero"`
		UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at,omitzero"`
		UpdatedBy int64     `bun:"updated_by,default:0" json:"updated_by,omitzero"`
		DeletedAt time.Time `bun:"deleted_at,nullzero,default:null" json:"deleted_at,omitzero"`
		DeletedBy int64     `bun:"deleted_by,default:0" json:"deleted_by,omitzero"`
	}

	QueryParams struct {
		UUID        string
		Filter      string
		FilterExtOp string
		FilterExt   string
		Sort        string
		Status      string
		Deleted     bool
		Limit       int
		Page        int
		Ctx         *gin.Context
	}

	AuditLog struct {
		bun.BaseModel `bun:"table:audit_logs,alias:au"`

		ID               int64     `bun:"id,pk,autoincrement" json:"id"`
		UserID           int64     `bun:"user_id" json:"user_id"`
		Token            string    `bun:"token,default:null" json:"token,omitempty"`
		Path             string    `bun:"path" json:"path"`
		Action           string    `bun:"action" json:"action"`
		ResponseStatus   int       `bun:"response_status" json:"response_status"`
		ModuleID         int64     `bun:"module_id" json:"module_id"`
		Module           string    `bun:"module" json:"module"`
		BeforeDataChange any       `bun:"before_data_change" json:"before_data_change"`
		AfterDataChange  any       `bun:"after_data_change" json:"after_data_change"`
		Description      string    `bun:"description,default:null" json:"description"`
		IPAddress        string    `bun:"ip_address" json:"ip_address"`
		UserAgent        string    `bun:"user_agent" json:"user_agent"`
		CreatedAt        time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at,omitzero"`
	}

	Results struct {
		Item  any
		Items []any
		Count int
	}
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
		q = q.Order("created_at ASC")
	}

	if qp.Status != "A" {
		if qp.Status == "" {
			qp.Status = "O"
		}

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

	if qp.Deleted {
		q = q.Where("deleted_at IS NOT NULL")
	} else {
		q = q.Where("deleted_at IS NULL")
	}

	return q
}

func buildFilterExtQuery(q *bun.SelectQuery, filters []string, filterExtOp string, kv []string) *bun.SelectQuery {
	identifier := kv[0] + " = ?"
	var literal any = kv[1]

	if strings.Contains(kv[1], "||") {
		identifier = kv[0] + " IN (?)"
		literal = bun.In(strings.Split(kv[1], "||"))
	}

	if len(filters) > 1 && filterExtOp == "OR" {
		return q.WhereOr(identifier, literal)
	}

	return q.Where(identifier, literal)
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

func setStatus(ctx *gin.Context, tableName, uuid, field string) (id int64, deletedAt time.Time, status, msg string, err error) {
	var temp struct {
		ID        int64     `bun:"id"`
		Status    string    `bun:"status" json:"status"`
		DeletedAt time.Time `bun:"deleted_at" json:"deleted_at"`
	}

	setQuery := "status = CASE WHEN status = 'O' THEN 'V' ELSE 'O' END"
	if field == "deleted_at" {
		setQuery = "deleted_at = CASE WHEN deleted_at IS NULL THEN NOW() ELSE NULL END"
	}

	err = executeTransaction(ctx, func(trx *bun.Tx) error {
		_, err := trx.NewUpdate().
			Table(tableName).
			Where("uuid = ?", uuid).
			Set(setQuery).
			Returning("deleted_at, status, id").
			Exec(ctx, &temp)
		return err
	})

	id = temp.ID

	status = "O"
	msg = "status restored successfully"
	if temp.Status == "V" {
		status = temp.Status
		msg = "status archived successfully"
	}

	if field == "deleted_at" {
		msg = "restored successfully"
		if !temp.DeletedAt.IsZero() {
			deletedAt = temp.DeletedAt
			msg = "deleted successfully"
		}
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

	allowedSortFields["id"] = true
	allowedSortFields["created_at"] = true

	return allowedSortFields[after]
}
