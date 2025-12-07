package controllers

import (
	"api/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	AppController
	u models.User
}

func (c UserController) InitUserController(router *gin.Engine) {
	r := router.Group(fmt.Sprintf("/%s/user", apiVersion))

	r.POST("", c.mw.Authenticate, c.Upsert)
	r.GET("/:uuid", c.mw.Authenticate, c.Read)
	r.DELETE("/:uuid", c.mw.Authenticate, c.Delete)
}

func (c UserController) Upsert(ctx *gin.Context) {
	var form *models.User
	if err := ctx.ShouldBindJSON(&form); err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	httpStatus, res, err := c.u.Upsert(ctx, *form)

	if err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	ctx.JSON(httpStatus, gin.H{"data": res})
}

func (c UserController) Read(ctx *gin.Context) {
	uuid, res, err := c.u.Read(c.sanitizeCtx(ctx))

	if err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	if uuid != "all" {
		ctx.JSON(http.StatusOK, gin.H{"data": res.User})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"total": res.Count, "data": res.Users})
	}
}

func (c UserController) Delete(ctx *gin.Context) {
	deletedAt, msg, err := c.u.Delete(ctx, ctx.Param("uuid"))

	if err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"deleted_at": deletedAt.String(), "message": msg})
}
