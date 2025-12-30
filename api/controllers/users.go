package controllers

import (
	"api/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	AppController
	m models.User
}

func (c UserController) InitUserController(router *gin.Engine) {
	r := router.Group(fmt.Sprintf("/%s/user", apiVersion))

	r.POST("", c.mw.Authenticate, c.mw.CheckPermission("user", "manage", "edit"), c.Upsert)
	r.GET("/:uuid", c.mw.Authenticate, c.mw.CheckPermission("user", "manage", "read"), c.Read)
	r.DELETE("/:uuid", c.mw.Authenticate, c.mw.CheckPermission("user", "manage", "delete"), c.Delete)
	r.PATCH("/:uuid", c.mw.Authenticate, c.mw.CheckPermission("user", "manage", "update_status"), c.UpdateStatus)
}

func (c UserController) Upsert(ctx *gin.Context) {
	var form *models.User
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

func (c UserController) Read(ctx *gin.Context) {
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

func (c UserController) Delete(ctx *gin.Context) {
	deletedAt, msg, err := c.m.Delete(ctx, ctx.Param("uuid"))

	if err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted_at": deletedAt.String(), "message": msg})
}

func (c UserController) UpdateStatus(ctx *gin.Context) {
	status, msg, err := c.m.UpdateStatus(ctx, ctx.Param("uuid"))

	if err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": status, "message": msg})
}
