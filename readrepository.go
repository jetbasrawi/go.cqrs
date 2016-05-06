package ycq

import (
	"errors"
	"github.com/jetbasrawi/yoono-uuid"
)

// Error returned when a model could not be found.
var ErrCouldNotSaveModel = errors.New("could not save model")

// Error returned when a model could not be found.
var ErrModelNotFound = errors.New("could not find model")

// ReadRepository is a storage for read models.
type ReadRepository interface {
	// Save saves a read model with id to the repository.
	Save(uuid.UUID, interface{}) error

	// Find returns one read model with using an id.
	Find(uuid.UUID) (interface{}, error)

	// FindAll returns all read models in the repository.
	FindAll() ([]interface{}, error)

	// Remove removes a read model with id from the repository.
	Remove(uuid.UUID) error
}

// MemoryReadRepository implements an in memory repository of read models.
type MemoryReadRepository struct {
	data map[uuid.UUID]interface{}
}

// NewMemoryReadRepository creates a new MemoryReadRepository.
func NewMemoryReadRepository() *MemoryReadRepository {
	r := &MemoryReadRepository{
		data: make(map[uuid.UUID]interface{}),
	}
	return r
}

// Save saves a read model with id to the repository.
func (r *MemoryReadRepository) Save(id uuid.UUID, model interface{}) error {
	r.data[id] = model
	return nil
}

// Find returns one read model with using an id. Returns
// ErrModelNotFound if no model could be found.
func (r *MemoryReadRepository) Find(id uuid.UUID) (interface{}, error) {
	if model, ok := r.data[id]; ok {
		return model, nil
	}

	return nil, ErrModelNotFound
}

// FindAll returns all read models in the repository.
func (r *MemoryReadRepository) FindAll() ([]interface{}, error) {
	models := []interface{}{}
	for _, model := range r.data {
		models = append(models, model)
	}
	return models, nil
}

// Remove removes a read model with id from the repository. Returns
// ErrModelNotFound if no model could be found.
func (r *MemoryReadRepository) Remove(id uuid.UUID) error {
	if _, ok := r.data[id]; ok {
		delete(r.data, id)
		return nil
	}

	return ErrModelNotFound
}
