package ycq

import (
	. "gopkg.in/check.v1"
)

type MemoryReadRepositorySuite struct{}

var _ = Suite(&MemoryReadRepositorySuite{})

func (s *MemoryReadRepositorySuite) TestNewMemoryReadRepository(c *C) {
	repo := NewMemoryReadRepository()
	c.Assert(repo, Not(Equals), nil)
	c.Assert(repo.data, Not(Equals), nil)
	c.Assert(len(repo.data), Equals, 0)
}

func (s *MemoryReadRepositorySuite) TestSave(c *C) {
	// Simple save.
	repo := NewMemoryReadRepository()
	id := yooid()
	repo.Save(id, 42)
	c.Assert(len(repo.data), Equals, 1)
	c.Assert(repo.data[id], Equals, 42)

	// Overwrite with same ID.
	repo = NewMemoryReadRepository()
	id = yooid()
	repo.Save(id, 42)
	repo.Save(id, 43)
	c.Assert(len(repo.data), Equals, 1)
	c.Assert(repo.data[id], Equals, 43)
}

func (s *MemoryReadRepositorySuite) TestFind(c *C) {
	// Simple find.
	repo := NewMemoryReadRepository()
	id := yooid()
	repo.data[id] = 42
	result, err := repo.Find(id)
	c.Assert(err, Equals, nil)
	c.Assert(result, Equals, 42)

	// Empty repo.
	repo = NewMemoryReadRepository()
	result, err = repo.Find(yooid())
	c.Assert(err, ErrorMatches, "could not find model")
	c.Assert(result, Equals, nil)

	// Non existing ID.
	repo = NewMemoryReadRepository()
	repo.data[yooid()] = 42
	result, err = repo.Find(yooid())
	c.Assert(err, ErrorMatches, "could not find model")
	c.Assert(result, Equals, nil)
}

func (s *MemoryReadRepositorySuite) TestFindAll(c *C) {
	// Find one.
	repo := NewMemoryReadRepository()
	repo.data[yooid()] = 42
	result, err := repo.FindAll()
	c.Assert(err, Equals, nil)
	c.Assert(result, DeepEquals, []interface{}{42})

	// Find two.
	repo = NewMemoryReadRepository()
	repo.data[yooid()] = 42
	repo.data[yooid()] = 43
	result, err = repo.FindAll()
	c.Assert(err, Equals, nil)
	sum := 0
	for _, v := range result {
		sum += v.(int)
	}
	c.Assert(sum, Equals, 85)

	// Find none.
	repo = NewMemoryReadRepository()
	result, err = repo.FindAll()
	c.Assert(err, Equals, nil)
	c.Assert(result, DeepEquals, []interface{}{})
}

func (s *MemoryReadRepositorySuite) TestRemove(c *C) {
	// Simple remove.
	repo := NewMemoryReadRepository()
	id := yooid()
	repo.data[id] = 42
	err := repo.Remove(id)
	c.Assert(err, Equals, nil)
	c.Assert(len(repo.data), Equals, 0)

	// Non existing ID.
	repo = NewMemoryReadRepository()
	repo.data[id] = 42
	err = repo.Remove(yooid())
	c.Assert(err, ErrorMatches, "could not find model")
	c.Assert(len(repo.data), Equals, 1)
}
