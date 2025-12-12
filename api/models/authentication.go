package models

import (
	"github.com/gin-gonic/gin"
)

type Authentication struct{}

func (m Authentication) Login(ctx *gin.Context, username string) (user User, err error) {
	q := db.NewSelect().Model(&user).Where("username = ?", username).WhereOr("email = ?", username)
	err = q.Scan(ctx)

	return user, err
}
