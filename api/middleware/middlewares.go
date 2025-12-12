package middleware

import (
	"api/models"
	"api/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"golang.org/x/time/rate"
)

type Middleware struct{}

var limiter = rate.NewLimiter(1, 5)

func SetLoggers(router *gin.Engine) {
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	router.Use(gin.Recovery())
}

func (m Middleware) Authenticate(ctx *gin.Context) {
	m.rateLimiter(ctx)

	ctx.Set("userId", 1)

	// @todo JWT LOGIC
	ctx.Next()
}

func (m Middleware) CheckPermission(module string, perm ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		items := make([]string, len(perm))
		for i, v := range perm {
			items[i] = module + ":" + v
		}

		var userPerm struct {
			models.User
			Permissions []string `bun:"permissions" json:"permissions"`
		}

		utils.GetPermissions(func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Column("u.*").Group("u.id")
		}, ctx.GetHeader("UserUUID"), ctx, &userPerm)

		if userPerm.ID == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Your account has either been deleted or archived; please contact support for more information or assistance."})
			ctx.Abort()
			return
		}

		if !m.checkPerm(items, userPerm.Permissions, userPerm.IsAdmin) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission.", "data": userPerm})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func (m Middleware) checkPerm(items, permissions []string, isAdmin bool) bool {
	if isAdmin {
		return true
	}

	permissionMap := make(map[string]struct{})
	for _, p := range permissions {
		permissionMap[p] = struct{}{}
	}

	for _, item := range items {
		if _, exists := permissionMap[item]; exists {
			return true
		}
	}

	return false
}

func (m Middleware) rateLimiter(ctx *gin.Context) {
	if !limiter.Allow() {
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
		ctx.Abort()
		return
	}
}
