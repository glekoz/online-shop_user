package cache

import (
	"time"

	"github.com/glekoz/cache"
)

type Cache struct {
	c   *cache.Cache[string, string]
	ttl time.Duration
}

func New(ttl int) (*Cache, error) {
	c, err := cache.New[string, string]()
	if err != nil {
		return nil, err
	}
	return &Cache{
		c:   c,
		ttl: time.Duration(ttl) * time.Second,
	}, nil
}

func (c *Cache) Add(userID string, mailtoken string) error {
	return c.c.Add(userID, mailtoken, c.ttl)
}

func (c *Cache) Get(userID string) (string, bool) {
	return c.c.Get(userID)
}

func (c *Cache) Delete(userID string) {
	c.c.Delete(userID)
}
