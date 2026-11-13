package core

import "time"

// Note — сущность «заметка».
type Note struct {
	ID        int64      `json:"id" example:"1"`
	Title     string     `json:"title" example:"Первая заметка"`
	Content   string     `json:"content" example:"Текст заметки"`
	CreatedAt time.Time  `json:"createdAt" example:"2025-10-18T12:00:00Z"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

// NoteCreate — DTO для создания заметки.
type NoteCreate struct {
	Title   string `json:"title" example:"Новая заметка"`
	Content string `json:"content" example:"Текст заметки"`
}

// NoteUpdate — DTO для частичного обновления.
type NoteUpdate struct {
	Title   *string `json:"title,omitempty" example:"Обновлено"`
	Content *string `json:"content,omitempty" example:"Новый текст"`
}