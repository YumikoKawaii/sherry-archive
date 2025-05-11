package repository

import "time"

type User struct {
	Id         int64     `gorm:"id"`
	Uuid       string    `gorm:"uuid"`
	Username   string    `gorm:"username"`
	Email      string    `gorm:"email"`
	Department string    `gorm:"department"`
	Status     string    `gorm:"status"`
	CreatedAt  time.Time `gorm:"created_at"`
	UpdatedAt  time.Time `gorm:"updated_at"`

	// Relationships
	Groups []Group `gorm:"many2many:user_groups;"`
	Roles  []Role  `gorm:"many2many:user_roles;"`
}

type Group struct {
	Id          int64     `gorm:"id"`
	Name        string    `gorm:"name"`
	Description string    `gorm:"description"`
	CreatedAt   time.Time `gorm:"created_at"`
	UpdatedAt   time.Time `gorm:"updated_at"`

	// Relationships
	Users []User `gorm:"many2many:user_groups;"`
	Roles []Role `gorm:"many2many:group_roles;"`
}

type Resource struct {
	Id          int64     `gorm:"id"`
	Name        string    `gorm:"name"`
	Path        string    `gorm:"path"`
	Description string    `gorm:"description"`
	CreatedAt   time.Time `gorm:"created_at"`
	UpdatedAt   time.Time `gorm:"updated_at"`
}

type Permission struct {
	Id          int64     `gorm:"id"`
	Name        string    `gorm:"name"`
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

	// Relationships
	Users  []User  `gorm:"many2many:user_roles;"`
	Groups []Group `gorm:"many2many:group_roles;"`
}

type Policy struct {
	Id            int64     `gorm:"id"`
	Name          string    `gorm:"name"`
	Description   string    `gorm:"description"`
	Effect        string    `gorm:"effect"`
	PrincipalType string    `gorm:"principal_type"`
	PrincipalId   int64     `gorm:"principal_id"`
	PermissionId  int64     `gorm:"permission_id"`
	CreatedAt     time.Time `gorm:"created_at"`
	UpdatedAt     time.Time `gorm:"updated_at"`

	// Foreign key for Permission
	Permission Permission `gorm:"foreignKey:PermissionID"`
}

type UserGroup struct {
	ID        int64     `gorm:"id"`
	UserID    int64     `gorm:"user_id"`
	GroupID   int64     `gorm:"group_id"`
	CreatedAt time.Time `gorm:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at"`
}

type PermissionResource struct {
	ID           int64     `gorm:"id"`
	PermissionID int64     `gorm:"permission_id"`
	ResourceID   int64     `gorm:"resource_id"`
	Action       string    `gorm:"action"`
	CreatedAt    time.Time `gorm:"created_at"`
	UpdatedAt    time.Time `gorm:"updated_at"`

	// Foreign keys
	Permission Permission `gorm:"foreignKey:PermissionID"`
	Resource   Resource   `gorm:"foreignKey:ResourceID"`
}

type UserRole struct {
	ID        int64     `gorm:"id"`
	UserID    int64     `gorm:"user_id"`
	RoleID    int64     `gorm:"role_id"`
	CreatedAt time.Time `gorm:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at"`
}

type GroupRole struct {
	ID        int64     `gorm:"id"`
	GroupID   int64     `gorm:"group_id"`
	RoleID    int64     `gorm:"role_id"`
	CreatedAt time.Time `gorm:"created_at"`
	UpdatedAt time.Time `gorm:"updated_at"`
}
