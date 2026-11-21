package model

import "time"

type AdminUser struct {
	ID           string        `gorm:"column:id;primaryKey" json:"id"`
	Username     string        `gorm:"column:username;uniqueIndex:idx_user_username" json:"username"`
	PasswordHash string        `gorm:"column:passwordHash" json:"-"`
	Salt         string        `gorm:"column:salt" json:"-"`
	Status       AccountStatus `gorm:"column:status" json:"status"`
	CreatedAt    time.Time     `gorm:"column:createdAt" json:"createdAt"`
	LastLoginAt  *time.Time    `gorm:"column:lastLoginAt" json:"lastLoginAt"`
	FailCount    int           `gorm:"column:failCount" json:"failCount"`
	LockedUntil  *time.Time    `gorm:"column:lockedUntil" json:"lockedUntil"`
	Roles        []Role        `gorm:"many2many:user_roles;joinForeignKey:userId;joinReferences:roleId" json:"roles"`
}

func (AdminUser) TableName() string { return "admin_users" }

type Role struct {
	ID          string   `gorm:"column:id;primaryKey" json:"id"`
	Name        string   `gorm:"column:name;uniqueIndex:idx_role_name" json:"name"`
	Permissions []string `gorm:"column:permissions;type:text;serializer:json" json:"permissions"`
	Builtin     bool     `gorm:"column:builtin" json:"builtin"`
	Description string   `gorm:"column:description;type:text" json:"description"`
}

func (Role) TableName() string { return "roles" }

type UserRole struct {
	UserID string `gorm:"column:userId;primaryKey" json:"userId"`
	RoleID string `gorm:"column:roleId;primaryKey" json:"roleId"`
}

func (UserRole) TableName() string { return "user_roles" }
