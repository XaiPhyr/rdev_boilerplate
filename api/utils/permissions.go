package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func GetPermissions(fn func(*bun.SelectQuery) *bun.SelectQuery, uuid string, ctx *gin.Context, dest ...any) error {
	q := db.NewSelect().
		With("UserPermissions", userPermissionsUnion()).
		TableExpr("users AS u").
		ColumnExpr("JSON_ARRAYAGG(p.name) AS permissions").
		Join(`LEFT JOIN "UserPermissions" up ON up.user_id = u.id`).
		Join("LEFT JOIN permissions p ON p.id = up.permission_id").
		WhereGroup("AND", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Where("u.uuid = ?", uuid).
				Where("u.status = 'O'").
				Where("u.deleted_at IS NULL")
		})

	if fn != nil {
		q = fn(q)
	}

	return q.Scan(ctx, &dest)
}

func userPermissionsUnion() *bun.SelectQuery {
	rolePermsQuery := db.NewSelect().TableExpr("users AS u").
		ColumnExpr("u.id AS user_id, rp.permission_id").
		Join("LEFT JOIN user_roles ur ON ur.user_id = u.id AND ur.deleted_at IS NULL AND ur.status = 'O'").
		Join("LEFT JOIN role_permissions rp ON rp.role_id = ur.role_id AND rp.deleted_at IS NULL AND rp.status = 'O'")

	groupPermsQuery := db.NewSelect().TableExpr("users AS U").
		ColumnExpr("u.id AS user_id, gp.permission_id").
		Join("LEFT JOIN user_groups ug ON ug.user_id = u.id AND ug.deleted_at IS NULL AND ug.status = 'O'").
		Join("LEFT JOIN group_permissions gp ON gp.group_id = ug.group_id AND gp.deleted_at IS NULL AND gp.status = 'O'")

	return rolePermsQuery.Union(groupPermsQuery)
}
