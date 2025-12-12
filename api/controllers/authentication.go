package controllers

import (
	"api/models"
	"api/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthenticationController struct {
	AppController
	u models.Authentication
}

func (c AuthenticationController) InitUserController(router *gin.Engine) {
	r := router.Group(fmt.Sprintf("/%s/auth", apiVersion))

	r.POST("/login", c.Login)
}

func (c AuthenticationController) Login(ctx *gin.Context) {
	var form struct {
		Username string `json:"username"`
	}

	if err := ctx.ShouldBindJSON(&form); err != nil {
		c.handleError(ctx, err, c.cleanErr(err))
		return
	}

	if u, err := c.u.Login(ctx, form.Username); err == nil {
		if jwt, err := utils.GenerateJWT(u.UUID, u.Username); err == nil {
			ctx.JSON(http.StatusOK, gin.H{"success": "Login successfully.", "data": jwt})
			return
		}
	}

	ctx.JSON(http.StatusNotFound, gin.H{"error": "Username does not exist."})
}
