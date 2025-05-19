package repository

import "time"

type Document struct {
	Id              uint32     `gorm:"id"`
	Title           string     `gorm:"title"`
	Description     *string    `gorm:"description"`
	ImageUrl        *string    `gorm:"image_url"`
	AuthorId        *uint32    `gorm:"author_id"`
	PublisherId     *uint32    `gorm:"publisher_id"`
	CategoryId      *uint32    `gorm:"category_id"`
	PublicationDate *time.Time `gorm:"publication_date"`
	CreatedAt       time.Time  `gorm:"created_at"`
	UpdatedAt       time.Time  `gorm:"updated_at"`
	DeletedAt       *time.Time `gorm:"deleted_at"`
}

type GetDocumentsFilter struct {
	Ids         []uint32
	AuthorId    *uint32
	PublisherId *uint32
	CategoryId  *uint32
	Pagination  *Pagination
}

type Chapter struct {
	Id         uint32     `gorm:"id"`
	Title      string     `gorm:"title"`
	Index      uint32     `gorm:"index"`
	DocumentId uint32     `gorm:"document_id"`
	CreatedAt  time.Time  `gorm:"created_at"`
	UpdatedAt  time.Time  `gorm:"updated_at"`
	DeletedAt  *time.Time `gorm:"deleted_at"`
}

type GetChaptersFilter struct {
	Ids        []uint32
	DocumentId *uint32
	Pagination *Pagination
}

type Page struct {
	Id        uint32     `gorm:"id"`
	ChapterId uint32     `gorm:"chapter_id"`
	ImageUrl  string     `gorm:"image_url"`
	Index     uint32     `gorm:"index"`
	CreatedAt time.Time  `gorm:"created_at"`
	UpdatedAt time.Time  `gorm:"updated_at"`
	DeletedAt *time.Time `gorm:"deleted_at"`
}

type GetPagesFilter struct {
	Ids        []uint32
	ChapterId  *uint32
	Pagination *Pagination
}

type Author struct {
	Id          uint32     `gorm:"id"`
	Name        string     `gorm:"name"`
	ImageUrl    string     `gorm:"image_url"`
	Description string     `gorm:"description"`
	CreatedAt   time.Time  `gorm:"created_at"`
	UpdatedAt   time.Time  `gorm:"updated_at"`
	DeletedAt   *time.Time `gorm:"deleted_at"`
}

type GetAuthorsFilter struct {
	Ids        []uint32
	Pagination *Pagination
}

type Publisher struct {
	Id          uint32     `gorm:"id"`
	Name        string     `gorm:"name"`
	ImageUrl    string     `gorm:"image_url"`
	Description string     `gorm:"description"`
	CreatedAt   time.Time  `gorm:"created_at"`
	UpdatedAt   time.Time  `gorm:"updated_at"`
	DeletedAt   *time.Time `gorm:"deleted_at"`
}

type GetPublishersFilter struct {
	Ids        []uint32
	Pagination *Pagination
}

type Pagination struct {
	Page     uint32
	PageSize uint32
}
