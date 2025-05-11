package repository

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Querier interface {
	GetUsers(context.Context) ([]User, error)
	UpsertUser(ctx context.Context, user *User) error
	GetGroups(ctx context.Context) ([]Group, error)
	UpsertGroup(ctx context.Context, group *Group) error
	UpsertUserGroup(ctx context.Context, relation *UserGroup) error
	GetResources(ctx context.Context) ([]Resource, error)
	UpsertResource(ctx context.Context, resource *Resource) error
	GetPermissions(ctx context.Context) ([]Permission, error)
	UpsertPermission(ctx context.Context, permission *Permission) error
	UpsertPermissionResource(ctx context.Context, relation *PermissionResource) error
	GetRoles(ctx context.Context) ([]Role, error)
	UpsertRole(ctx context.Context, role *Role) error
	UpsertUserRole(ctx context.Context, relation *UserRole) error
	UpsertGroupRole(ctx context.Context, relation *GroupRole) error
	GetPolicies(ctx context.Context) ([]Policy, error)
	UpsertPolicy(ctx context.Context, policy *Policy) error
}

type querierImpl struct {
	db *gorm.DB
}

func NewQuerier(db *gorm.DB) Querier {
	return &querierImpl{db: db}
}

func (q *querierImpl) GetUsers(ctx context.Context) ([]User, error) {
	var users []User
	err := q.db.WithContext(ctx).Find(&users).Error
	return users, err
}

func (q *querierImpl) UpsertUser(ctx context.Context, user *User) error {
	return q.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(user).Error
}

func (q *querierImpl) GetGroups(ctx context.Context) ([]Group, error) {
	var groups []Group
	err := q.db.WithContext(ctx).Find(&groups).Error
	return groups, err
}

func (q *querierImpl) UpsertGroup(ctx context.Context, group *Group) error {
	return q.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(group).Error
}

func (q *querierImpl) UpsertUserGroup(ctx context.Context, relation *UserGroup) error {
	return q.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "group_id"}},
		UpdateAll: true,
	}).Create(relation).Error
}

func (q *querierImpl) GetResources(ctx context.Context) ([]Resource, error) {
	var resources []Resource
	err := q.db.WithContext(ctx).Find(&resources).Error
	return resources, err
}

func (q *querierImpl) UpsertResource(ctx context.Context, resource *Resource) error {
	return q.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(resource).Error
}

func (q *querierImpl) GetPermissions(ctx context.Context) ([]Permission, error) {
	var permissions []Permission
	err := q.db.WithContext(ctx).Find(&permissions).Error
	return permissions, err
}

func (q *querierImpl) UpsertPermission(ctx context.Context, permission *Permission) error {
	return q.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(permission).Error
}

func (q *querierImpl) UpsertPermissionResource(ctx context.Context, relation *PermissionResource) error {
	return q.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "permission_id"}, {Name: "resource_id"}, {Name: "action"}},
		UpdateAll: true,
	}).Create(relation).Error
}

func (q *querierImpl) GetRoles(ctx context.Context) ([]Role, error) {
	var roles []Role
	err := q.db.WithContext(ctx).Find(&roles).Error
	return roles, err
}

func (q *querierImpl) UpsertRole(ctx context.Context, role *Role) error {
	return q.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(role).Error
}

func (q *querierImpl) UpsertUserRole(ctx context.Context, relation *UserRole) error {
	return q.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "role_id"}},
		UpdateAll: true,
	}).Create(relation).Error
}

func (q *querierImpl) UpsertGroupRole(ctx context.Context, relation *GroupRole) error {
	return q.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "role_id"}},
		UpdateAll: true,
	}).Create(relation).Error
}

func (q *querierImpl) GetPolicies(ctx context.Context) ([]Policy, error) {
	var policies []Policy
	err := q.db.WithContext(ctx).Find(&policies).Error
	return policies, err
}

func (q *querierImpl) UpsertPolicy(ctx context.Context, policy *Policy) error {
	return q.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(policy).Error
}
