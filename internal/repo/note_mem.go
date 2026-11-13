package repo

import (
	"sync"
	"time"

	"example.com/notes-api/internal/core"
)

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
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	r.notes[n.ID] = &n
	return n.ID, nil
}

func (r *NoteRepoMem) List() []core.Note {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]core.Note, 0, len(r.notes))
	for _, n := range r.notes {
		out = append(out, *n)
	}
	return out
}

func (r *NoteRepoMem) Get(id int64) (core.Note, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notes[id]
	if !ok {
		return core.Note{}, false
	}
	return *n, true
}

func (r *NoteRepoMem) Update(id int64, upd func(*core.Note)) (core.Note, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.notes[id]
	if !ok {
		return core.Note{}, false
	}
	upd(n)
	now := time.Now().UTC()
	n.UpdatedAt = &now
	return *n, true
}

func (r *NoteRepoMem) Delete(id int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.notes[id]; ok {
		delete(r.notes, id)
		return true
	}
	return false
}