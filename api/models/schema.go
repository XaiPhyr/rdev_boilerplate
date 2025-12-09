package models

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type (
	AppModel struct {
		Active    bool      `bun:"active,default:true" json:"active"`
		Status    string    `bun:"status,default:O" json:"status"`
		Flag      string    `bun:"flag,nullzero,default:null" json:"flag,omitempty"`
		UUID      string    `bun:"uuid,default:gen_random_uuid()" json:"uuid,omitempty"`
		CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at,omitzero"`
		CreatedBy int64     `bun:"created_by,default:0" json:"created_by,omitzero"`
		UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at,omitzero"`
		UpdatedBy int64     `bun:"updated_by,default:0" json:"updated_by,omitzero"`
		DeletedAt time.Time `bun:"deleted_at,nullzero,default:null" json:"deleted_at,omitzero"`
		DeletedBy int64     `bun:"deleted_by,default:0" json:"deleted_by,omitzero"`
	}

	QueryParams struct {
		UUID        string
		Filter      string
		FilterExtOp string
		FilterExt   string
		Sort        string
		Status      string
		Limit       int
		Page        int
		Ctx         *gin.Context
	}

	AuditLog struct {
		bun.BaseModel `bun:"table:audit_logs,alias:au"`

		ID               int64     `bun:"id,pk,autoincrement" json:"id"`
		UserID           int64     `bun:"user_id" json:"user_id"`
		Token            string    `bun:"token,default:null" json:"token,omitempty"`
		Path             string    `bun:"path" json:"path"`
		Action           string    `bun:"action" json:"action"`
		ResponseStatus   int       `bun:"response_status" json:"response_status"`
		ModuleID         int64     `bun:"module_id" json:"module_id"`
		Module           string    `bun:"module" json:"module"`
		BeforeDataChange any       `bun:"before_data_change" json:"before_data_change"`
		AfterDataChange  any       `bun:"after_data_change" json:"after_data_change"`
		Description      string    `bun:"description,default:null" json:"description"`
		IPAddress        string    `bun:"ip_address" json:"ip_address"`
		UserAgent        string    `bun:"user_agent" json:"user_agent"`
		CreatedAt        time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at,omitzero"`
	}

	UserResults struct {
		User  User
		Users []User
		Count int
	}

	User struct {
		bun.BaseModel `bun:"table:users,alias:u"`

		ID         int64           `bun:"id,pk,autoincrement" json:"id"`
		Username   string          `bun:"username" json:"username"`
		Password   string          `bun:"password" json:"-"`
		Email      string          `bun:"email,unique" json:"email"`
		FirstName  *string         `bun:"first_name,nullzero,default:null" json:"first_name"`
		MiddleName *string         `bun:"middle_name,nullzero,default:null" json:"middle_name"`
		LastName   *string         `bun:"last_name,nullzero,default:null" json:"last_name"`
		Mobile     *string         `bun:"mobile,unique,nullzero,default:null" json:"mobile"`
		Address    *map[string]any `bun:"address,type:jsonb" json:"address"`
		Optin      bool            `bun:"optin,default:false" json:"optin"`
		LastLogin  time.Time       `bun:"last_login,nullzero,default:null" json:"last_login,omitzero"`
		Metadata   *map[string]any `bun:"metadata,type:jsonb,default:null" json:"metadata"`
		IsAdmin    bool            `bun:"is_admin" json:"is_admin"`

		AppModel
	}

	UserRoleResults struct {
		UserRole  UserRole
		UserRoles []UserRole
		Count     int
	}

	UserRole struct {
		bun.BaseModel `bun:"table:user_roles,alias:ur"`

		ID     int64 `bun:"id,pk,autoincrement" json:"id"`
		UserID int64 `bun:"user_id" json:"user_id"`
		RoleID int64 `bun:"role_id" json:"role_id"`

		AppModel
	}

	RoleResults struct {
		Role  Role
		Roles []Role
		Count int
	}

	Role struct {
		bun.BaseModel `bun:"table:roles,alias:r"`

		ID          int64  `bun:"id,pk,autoincrement" json:"id"`
		Name        string `bun:"name" json:"name"`
		Description string `bun:"description" json:"description"`

		AppModel
	}

	PermissionResults struct {
		Permission  Permission
		Permissions []Permission
		Count       int
	}

	Permission struct {
		bun.BaseModel `bun:"table:permissions,alias:p"`

		ID          int64  `bun:"id,pk,autoincrement" json:"id"`
		Name        string `bun:"name" json:"name"`
		Description string `bun:"description" json:"description"`

		AppModel
	}

	//@todo user_roles
	//@todo role_permissions
)
