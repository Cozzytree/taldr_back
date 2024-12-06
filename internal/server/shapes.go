package server

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Cozzytree/taldrBack/internal/database"
	"github.com/Cozzytree/taldrBack/internal/models"
)

type SandDoc struct {
	Utype       string         `json:"u_type"`
	WorkspaceId string         `json:"workspaceId"`
	UserId      string         `json:"userId"`
	Document    string         `json:"document"`
	Shapes      []models.Shape `json:"shapes"`
}

type ShapeStore struct {
	mu       sync.Mutex
	store    map[string]SandDoc
	docStore map[string]SandDoc
}

func NewStore() *ShapeStore {
	return &ShapeStore{
		store:    make(map[string]SandDoc),
		docStore: make(map[string]SandDoc),
	}
}

func (s *ShapeStore) storeShapes(id string, shapes []models.Shape) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[id] = SandDoc{
		Shapes: shapes,
	}

}

func (s *ShapeStore) saveDocument(id string, document string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.docStore[id] = SandDoc{
		Document: document,
	}
}

func (s *ShapeStore) storeinDb(db database.Service) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	if len(s.store) > 0 {
		for key, v := range s.store {
			id := strings.Split(key, "+")
			if err := db.SaveShapetoDb(ctx, id[0], id[1], v.Shapes); err != nil {
				s.clear()
				return err
			}
		}
		s.clear()
	}

	if len(s.docStore) > 0 {
		for key, v := range s.docStore {
			id := strings.Split(key, "+")
			if err := db.SaveDocsToDb(ctx, id[0], id[1], v.Document); err != nil {
				s.clearDocS()
				return err
			}
		}
		s.clearDocS()
	}

	return nil
}

func (s *ShapeStore) clear() {
	for k := range s.store {
		delete(s.store, k)
	}
}
func (s *ShapeStore) clearDocS() {
	for k := range s.docStore {
		delete(s.docStore, k)
	}
}
