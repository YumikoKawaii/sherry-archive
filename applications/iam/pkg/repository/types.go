package repository

import "time"

type User struct {
	Id             int64     `gorm:"id"`
	Email          string    `gorm:"email"`
	HashedPassword string    `gorm:"hashed_password"`
	Username       string    `gorm:"username"`
	Department     string    `gorm:"department"`
	Status         string    `gorm:"status"`
	CreatedAt      time.Time `gorm:"created_at"`
	UpdatedAt      time.Time `gorm:"updated_at"`

	// Relationships
	Roles []Role `gorm:"many2many:user_roles;"`
}

type Resource struct {
	Id          int64     `gorm:"id"`
	Name        string    `gorm:"name"`
	Path        string    `gorm:"path"`
	Description string    `gorm:"description"`
	CreatedAt   time.Time `gorm:"created_at"`
	UpdatedAt   time.Time `gorm:"updated_at"`
}

type Role struct {
	Id          int64     `gorm:"id"`
	Name        string    `gorm:"name"`
	Description string    `gorm:"description"`
	CreatedAt   time.Time `gorm:"created_at"`
	UpdatedAt   time.Time `gorm:"updated_at"`
}

type Policy struct {
	Id            int64     `gorm:"id"`
	Name          string    `gorm:"name"`
	Description   string    `gorm:"description"`
	Effect        string    `gorm:"effect"`
	PrincipalType string    `gorm:"principal_type"`
	PrincipalId   int64     `gorm:"principal_id"`
	ResourceId    int64     `gorm:"resource_id"`
	CreatedAt     time.Time `gorm:"created_at"`
	UpdatedAt     time.Time `gorm:"updated_at"`

	// Foreign key for Permission
	Resource Resource `gorm:"foreignKey:resource_id"`
}

type UserRole struct {
	Id        int64     `gorm:"id"`
	UserId    int64     `gorm:"user_id"`
	Role      string    `gorm:"role"`
	CreatedAt time.Time `gorm:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at"`
}

type GetUsersFilter struct {
	Ids            []int64
	Status         *string
	Email          *string
	HashedPassword *string
}
