package repository

import "time"

type Book struct {
	ID              uint32     `gorm:"id"`
	Title           string     `gorm:"title"`
	Description     string     `gorm:"description"`
	ImageUrl        string     `gorm:"image_url"`
	AuthorId        uint32     `gorm:"author_id"`
	PublisherId     uint32     `gorm:"publisher_id"`
	CategoryId      uint32     `gorm:"category_id"`
	PublicationDate time.Time  `gorm:"publication_date"`
	CreatedAt       time.Time  `gorm:"created_at"`
	UpdatedAt       time.Time  `gorm:"updated_at"`
	DeletedAt       *time.Time `gorm:"deleted_at"`
}

type GetBooksFilter struct {
	IDs         []uint32
	AuthorId    uint32
	PublisherId uint32
	CategoryId  uint32
	Pagination  *Pagination
}

type Page struct {
	ID        uint32     `gorm:"id"`
	BookID    uint32     `gorm:"book_id"`
	ImageUrl  string     `gorm:"image_url"`
	CreatedAt time.Time  `gorm:"created_at"`
	UpdatedAt time.Time  `gorm:"updated_at"`
	DeletedAt *time.Time `gorm:"deleted_at"`
}

type GetPagesFilter struct {
	IDs        []uint32
	BookId     uint32
	Pagination *Pagination
}

type Author struct {
	ID          uint32     `gorm:"id"`
	Name        string     `gorm:"name"`
	ImageUrl    string     `gorm:"image_url"`
	Description string     `gorm:"description"`
	CreatedAt   time.Time  `gorm:"created_at"`
	UpdatedAt   time.Time  `gorm:"updated_at"`
	DeletedAt   *time.Time `gorm:"deleted_at"`
}

type GetAuthorsFilter struct {
	IDs        []uint32
	Pagination *Pagination
}

type Publisher struct {
	ID          uint32     `gorm:"id"`
	Name        string     `gorm:"name"`
	ImageUrl    string     `gorm:"image_url"`
	Description string     `gorm:"description"`
	CreatedAt   time.Time  `gorm:"created_at"`
	UpdatedAt   time.Time  `gorm:"updated_at"`
	DeletedAt   *time.Time `gorm:"deleted_at"`
}

type GetPublishersFilter struct {
	IDs        []uint32
	Pagination *Pagination
}

type Pagination struct {
	Page     uint32
	PageSize uint32
}
