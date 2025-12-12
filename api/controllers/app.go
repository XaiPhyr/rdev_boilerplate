package controllers

import (
	"api/middleware"
	"api/models"
	"api/utils"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun/driver/pgdriver"
)

type (
	AppController struct {
		mw middleware.Middleware
	}
)

var apiVersion = utils.InitConfig().Server.Endpoint

func (c *AppController) sanitizeCtx(ctx *gin.Context) models.QueryParams {
	limit, _ := strconv.Atoi(ctx.Query("limit"))
	if limit <= 0 {
		limit = 10
	}

	page, _ := strconv.Atoi(ctx.Query("page"))
	if page <= 0 {
		page = 1
	}

	filterExt, _ := url.QueryUnescape(ctx.Query("filterExt"))

	filterExtOp := ctx.Query("filterExtOp")
	if filterExtOp == "" {
		filterExtOp = "AND"
	}

	isDeleted, _ := strconv.ParseBool(ctx.Query("deleted"))

	return models.QueryParams{
		UUID:        ctx.Param("uuid"),
		Sort:        ctx.Query("sort"),
		Status:      ctx.Query("status"),
		Filter:      ctx.Query("filter"),
		FilterExtOp: filterExtOp,
		FilterExt:   filterExt,
		Deleted:     isDeleted,
		Limit:       limit,
		Page:        page,
		Ctx:         ctx,
	}
}

func (c *AppController) cleanErr(err error) string {
	if pgErr, ok := err.(pgdriver.Error); ok {
		return fmt.Sprintf("Postgres Error: %s", pgErr.Field('M'))
	}

	return fmt.Sprintf("Unknown Error: %s", err.Error())
}

func (c *AppController) handleError(ctx *gin.Context, err error, message string) {
	ctx.JSON(http.StatusNotFound, gin.H{"error": message, "details": err.Error()})
}
