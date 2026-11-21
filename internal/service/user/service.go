package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"pclsguserver/internal/audit"
	"pclsguserver/internal/middleware"
	"pclsguserver/internal/model"
	"pclsguserver/internal/repo"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrProtectedUser     = errors.New("protected user")
	ErrRoleNotFound      = errors.New("role not found")
	ErrBuiltinRole       = errors.New("builtin role cannot be deleted")
)

type Service struct {
	userRepo  *repo.AdminUserRepo
	roleRepo  *repo.RoleRepo
	audit     *audit.Service
	permCache *middleware.PermissionCache
	db        *gorm.DB
}

func NewService(ur *repo.AdminUserRepo, rr *repo.RoleRepo, a *audit.Service, pc *middleware.PermissionCache, db *gorm.DB) *Service {
	return &Service{userRepo: ur, roleRepo: rr, audit: a, permCache: pc, db: db}
}

func (s *Service) CreateUser(username, password string, roleIDs []string, operator, requestID, sourceIP string) (*model.AdminUser, error) {
	if _, err := s.userRepo.GetByUsername(username); err == nil {
		return nil, ErrUserAlreadyExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.AdminUser{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: string(hash),
		Status:       model.AccountStatusActive,
		CreatedAt:    time.Now(),
	}
	tx := s.db.Begin()
	if err := s.userRepo.Create(tx, u); err != nil {
		tx.Rollback()
		return nil, err
	}
	for _, rid := range roleIDs {
		if err := tx.Create(&model.UserRole{UserID: u.ID, RoleID: rid}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator: operator, Action: model.AuditActionUserRoleChanged, TargetType: model.AuditTargetUser,
		TargetID: u.ID, Result: model.AuditResultSuccess, Detail: "create user " + username, SourceIP: sourceIP, RequestID: requestID,
	}); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) ListUsers(page, size int) ([]model.AdminUser, int64, error) {
	return s.userRepo.List(page, size)
}

func (s *Service) GetUser(id string) (*model.AdminUser, error) {
	return s.userRepo.GetByID(id)
}

func (s *Service) Disable(id, operator, requestID, sourceIP string) error {
	u, err := s.userRepo.GetByID(id)
	if err != nil {
		return ErrUserNotFound
	}
	if u.ID == "user-super-admin" {
		return ErrProtectedUser
	}
	u.Status = model.AccountStatusDisabled
	tx := s.db.Begin()
	if err := s.userRepo.Update(tx, u); err != nil {
		tx.Rollback()
		return err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator: operator, Action: model.AuditActionUserRoleChanged, TargetType: model.AuditTargetUser,
		TargetID: id, Result: model.AuditResultSuccess, Detail: "disable user", SourceIP: sourceIP, RequestID: requestID,
	}); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	s.permCache.Invalidate(id)
	return nil
}

func (s *Service) Enable(id, operator, requestID, sourceIP string) error {
	u, err := s.userRepo.GetByID(id)
	if err != nil {
		return ErrUserNotFound
	}
	u.Status = model.AccountStatusActive
	u.FailCount = 0
	u.LockedUntil = nil
	tx := s.db.Begin()
	if err := s.userRepo.Update(tx, u); err != nil {
		tx.Rollback()
		return err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator: operator, Action: model.AuditActionUserRoleChanged, TargetType: model.AuditTargetUser,
		TargetID: id, Result: model.AuditResultSuccess, Detail: "enable user", SourceIP: sourceIP, RequestID: requestID,
	}); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	s.permCache.Invalidate(id)
	return nil
}

func (s *Service) Delete(id string, confirm bool, operator, requestID, sourceIP string) error {
	if !confirm {
		return errors.New("confirmation required")
	}
	u, err := s.userRepo.GetByID(id)
	if err != nil {
		return ErrUserNotFound
	}
	if u.ID == "user-super-admin" {
		return ErrProtectedUser
	}
	count, err := s.userRepo.CountSuperAdmins()
	if err != nil {
		return err
	}
	isSuper := false
	for _, r := range u.Roles {
		if r.ID == "role-super-admin" {
			isSuper = true
			break
		}
	}
	if isSuper && count <= 1 {
		return ErrProtectedUser
	}
	tx := s.db.Begin()
	if err := s.userRepo.Delete(tx, id); err != nil {
		tx.Rollback()
		return err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator: operator, Action: model.AuditActionDelete, TargetType: model.AuditTargetUser,
		TargetID: id, Result: model.AuditResultSuccess, Detail: "delete user " + u.Username, SourceIP: sourceIP, RequestID: requestID,
	}); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	s.permCache.Invalidate(id)
	return nil
}

func (s *Service) AssignRoles(id string, roleIDs []string, operator, requestID, sourceIP string) error {
	tx := s.db.Begin()
	if err := s.userRepo.AssignRoles(tx, id, roleIDs); err != nil {
		tx.Rollback()
		return err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator: operator, Action: model.AuditActionUserRoleChanged, TargetType: model.AuditTargetUser,
		TargetID: id, Result: model.AuditResultSuccess, Detail: "assign roles", SourceIP: sourceIP, RequestID: requestID,
	}); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	s.permCache.Invalidate(id)
	return nil
}

func (s *Service) ListRoles() ([]model.Role, error) {
	return s.roleRepo.List()
}

func (s *Service) CreateRole(name string, permissions []string, description, operator, requestID, sourceIP string) (*model.Role, error) {
	r := &model.Role{
		ID:          uuid.NewString(),
		Name:        name,
		Permissions: permissions,
		Builtin:     false,
		Description: description,
	}
	tx := s.db.Begin()
	if err := s.roleRepo.Create(tx, r); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator: operator, Action: model.AuditActionUserRoleChanged, TargetType: model.AuditTargetRole,
		TargetID: r.ID, Result: model.AuditResultSuccess, Detail: "create role " + name, SourceIP: sourceIP, RequestID: requestID,
	}); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	s.permCache.InvalidateAll()
	return r, nil
}

func (s *Service) UpdateRole(id string, permissions []string, operator, requestID, sourceIP string) error {
	r, err := s.roleRepo.GetByID(id)
	if err != nil {
		return ErrRoleNotFound
	}
	r.Permissions = permissions
	tx := s.db.Begin()
	if err := s.roleRepo.Update(tx, r); err != nil {
		tx.Rollback()
		return err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator: operator, Action: model.AuditActionUserRoleChanged, TargetType: model.AuditTargetRole,
		TargetID: id, Result: model.AuditResultSuccess, Detail: "update role", SourceIP: sourceIP, RequestID: requestID,
	}); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	s.permCache.InvalidateAll()
	return nil
}

func (s *Service) DeleteRole(id, operator, requestID, sourceIP string) error {
	r, err := s.roleRepo.GetByID(id)
	if err != nil {
		return ErrRoleNotFound
	}
	if r.Builtin {
		return ErrBuiltinRole
	}
	tx := s.db.Begin()
	if err := s.roleRepo.Delete(tx, id); err != nil {
		tx.Rollback()
		return err
	}
	if err := s.audit.Record(tx, audit.Record{
		Operator: operator, Action: model.AuditActionDelete, TargetType: model.AuditTargetRole,
		TargetID: id, Result: model.AuditResultSuccess, Detail: "delete role " + r.Name, SourceIP: sourceIP, RequestID: requestID,
	}); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	s.permCache.InvalidateAll()
	return nil
}
