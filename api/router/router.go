package router

import (
	"api/controllers"

	"github.com/gin-gonic/gin"
)

func InitRouters(router *gin.Engine) {
	var authentication = controllers.AuthenticationController{}
	authentication.InitUserController(router)

	var user = controllers.UserController{}
	user.InitUserController(router)

	var user_role = controllers.UserRoleController{}
	user_role.InitUserRoleController(router)

	var role = controllers.RoleController{}
	role.InitRoleController(router)

	var role_permission = controllers.RolePermissionController{}
	role_permission.InitRolePermissionController(router)

	var permission = controllers.PermissionController{}
	permission.InitPermissionController(router)

	var group = controllers.GroupController{}
	group.InitGroupController(router)

	var user_group = controllers.UserGroupController{}
	user_group.InitUserGroupController(router)

	var group_permission = controllers.GroupPermissionController{}
	group_permission.InitGroupPermissionController(router)
}
