package models

import "time"

// User — автор заметок (1 пользователь → много заметок)
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;not null"`
	Email     string `gorm:"size:200;uniqueIndex;not null"`
	Notes     []Note // 1:N
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Note — заметка (принадлежит пользователю, имеет много тегов)
type Note struct {
	ID        uint   `gorm:"primaryKey"`
	Title     string `gorm:"size:200;not null"`
	Content   string `gorm:"type:text"`
	UserID    uint   `gorm:"not null"`
	User      User
	Tags      []Tag `gorm:"many2many:note_tags;"` // M:N
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Tag — тег (многие ко многим с заметками через note_tags)
type Tag struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:50;uniqueIndex;not null"`
	Notes     []Note `gorm:"many2many:note_tags;"`
	CreatedAt time.Time
	UpdatedAt time.Time
}