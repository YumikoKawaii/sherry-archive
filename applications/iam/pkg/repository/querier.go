package repository

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sherry.archive.com/applications/iam/pkg/constants"
)

type Querier interface {
	GetUsers(ctx context.Context, filter *GetUsersFilter) ([]User, error)
	UpsertUser(ctx context.Context, user *User) error
	InitialUser(ctx context.Context, user *User) error
	GetResources(ctx context.Context) ([]Resource, error)
	UpsertResource(ctx context.Context, resource *Resource) error
	GetRoles(ctx context.Context) ([]Role, error)
	UpsertRole(ctx context.Context, role *Role) error
	UpsertUserRole(ctx context.Context, relation *UserRole) error
	GetPolicies(ctx context.Context) ([]Policy, error)
	UpsertPolicy(ctx context.Context, policy *Policy) error
}

type querierImpl struct {
	db *gorm.DB
}

func NewQuerier(db *gorm.DB) Querier {
	return &querierImpl{db: db}
}

func (q *querierImpl) GetUsers(ctx context.Context, filter *GetUsersFilter) ([]User, error) {
	accounts := make([]User, 0)
	query := q.db.Model(&User{})
	if filter != nil {
		if len(filter.Ids) != 0 {
			query = query.Where("id in (?)", filter.Ids)
		}

		if filter.Status != nil {
			query = query.Where("status = ?", filter.Status)
		}

		if filter.Email != nil {
			query = query.Where("email = ?", filter.Email)
		}

		if filter.HashedPassword != nil {
			query = query.Where("hashed_password = ?", filter.HashedPassword)
		}
	}
	return accounts, query.WithContext(ctx).Find(&accounts).Error
}

func (q *querierImpl) UpsertUser(ctx context.Context, user *User) error {
	return q.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(user).Error
}

func (q *querierImpl) InitialUser(ctx context.Context, user *User) error {
	return q.db.Transaction(func(tx *gorm.DB) error {
		// create user
		// create default role assign to user
		if err := tx.Model(&User{}).WithContext(ctx).Create(user).Error; err != nil {
			return err
		}

		return tx.Model(&UserRole{}).WithContext(ctx).Create(&UserRole{
			UserId: user.Id,
			Role:   constants.ReaderRole,
		}).Error
	})
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
