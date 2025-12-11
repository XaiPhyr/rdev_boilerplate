package controllers

import (
	"api/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GroupController struct {
	AppController
	m models.Group
}

func (c GroupController) InitGroupController(router *gin.Engine) {
	r := router.Group(fmt.Sprintf("/%s/group", apiVersion))

	r.POST("", c.mw.Authenticate, c.Upsert)
	r.GET("/:uuid", c.mw.Authenticate, c.Read)
	r.DELETE("/:uuid", c.mw.Authenticate, c.Delete)
	r.PATCH("/:uuid", c.mw.Authenticate, c.UpdateStatus)
}

func (c GroupController) Upsert(ctx *gin.Context) {
	var form *models.Group
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

func (c GroupController) Read(ctx *gin.Context) {
	res, err := c.m.Read(c.sanitizeCtx(ctx))

	if err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	if ctx.Param("uuid") != "all" {
		ctx.JSON(http.StatusOK, gin.H{"data": res.Group})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"total": res.Count, "data": res.Groups})
	}
}

func (c GroupController) Delete(ctx *gin.Context) {
	deletedAt, msg, err := c.m.Delete(ctx, ctx.Param("uuid"))

	if err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted_at": deletedAt.String(), "message": msg})
}

func (c GroupController) UpdateStatus(ctx *gin.Context) {
	status, msg, err := c.m.UpdateStatus(ctx, ctx.Param("uuid"))

	if err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": status, "message": msg})
}
