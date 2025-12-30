package controllers

import (
	"api/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserRoleController struct {
	AppController
	m models.UserRole
}

func (c UserRoleController) InitUserRoleController(router *gin.Engine) {
	r := router.Group(fmt.Sprintf("/%s/user_role", apiVersion))

	r.POST("", c.mw.Authenticate, c.Upsert)
	r.GET("/:uuid", c.mw.Authenticate, c.Read)
	r.DELETE("/:uuid", c.mw.Authenticate, c.Delete)
	r.PATCH("/:uuid", c.mw.Authenticate, c.UpdateStatus)
}

func (c UserRoleController) Upsert(ctx *gin.Context) {
	var form *models.UserRole
	if err := ctx.ShouldBindJSON(&form); err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	httpStatus, res, err := c.m.Upsert(ctx, *form)

	if err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	ctx.JSON(httpStatus, gin.H{"data": res})
}

func (c UserRoleController) Read(ctx *gin.Context) {
	res, err := c.m.Read(c.sanitizeCtx(ctx))

	if err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	data := gin.H{"total": res.Count, "data": res.Items}

	if ctx.Param("uuid") != "all" {
		data = gin.H{"data": res.Item}
	}

	ctx.JSON(http.StatusOK, data)
}

func (c UserRoleController) Delete(ctx *gin.Context) {
	deletedAt, msg, err := c.m.Delete(ctx, ctx.Param("uuid"))

	if err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted_at": deletedAt.String(), "message": msg})
}

func (c UserRoleController) UpdateStatus(ctx *gin.Context) {
	status, msg, err := c.m.UpdateStatus(ctx, ctx.Param("uuid"))

	if err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": status, "message": msg})
}
