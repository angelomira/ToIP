package repo

import (
	"errors"
	"sort"
	"sync"
	"time"

	"example.com/notes-api/internal/core"
)

var ErrNotFound = errors.New("note not found")

type NoteRepoMem struct {
	mu    sync.Mutex
	notes map[int64]*core.Note
	next  int64
}

func NewNoteRepoMem() *NoteRepoMem {
	return &NoteRepoMem{notes: make(map[int64]*core.Note)}
}

func (r *NoteRepoMem) Create(n core.Note) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.next++
	n.ID = r.next
	n.CreatedAt = time.Now()
	r.notes[n.ID] = &n
	return n.ID, nil
}

func (r *NoteRepoMem) GetByID(id int64) (*core.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, ok := r.notes[id]
	if !ok {
		return nil, ErrNotFound
	}
	return n, nil
}

func (r *NoteRepoMem) List() []*core.Note {
	r.mu.Lock()
	defer r.mu.Unlock()

	res := make([]*core.Note, 0, len(r.notes))
	for _, n := range r.notes {
		res = append(res, n)
	}
	sort.Slice(res, func(i, j int) bool { return res[i].ID < res[j].ID })
	return res
}

func (r *NoteRepoMem) Update(id int64, title, content *string) (*core.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, ok := r.notes[id]
	if !ok {
		return nil, ErrNotFound
	}
	if title != nil {
		n.Title = *title
	}
	if content != nil {
		n.Content = *content
	}
	now := time.Now()
	n.UpdatedAt = &now
	return n, nil
}

func (r *NoteRepoMem) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.notes[id]; !ok {
		return ErrNotFound
	}
	delete(r.notes, id)
	return nil
}