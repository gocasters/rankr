package inmemoryrepository

import "github.com/patrickmn/go-cache"

type InMemoryBuffered struct {
	c *cache.Cache
}

func New() InMemoryBuffered {
	return InMemoryBuffered{}
}
