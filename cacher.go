package cacher

import lru "github.com/hashicorp/golang-lru"

type Cache struct {
	c *lru.Cache
	l *lru.Cache
}
type LongProcess func() (interface{}, error)

func New(size int) (*Cache, error) {
	c, e := lru.New(size)
	if e != nil {
		return nil, e
	}
	l, e := lru.New(size)
	if e != nil {
		return nil, e
	}
	return &Cache{c: c, l: l}, nil
}

func MustNew(size int) *Cache {
	c, e := New(size)
	if e != nil {
		panic(e)
	}
	return c
}

func (c *Cache) GetOrProcess(key interface{}, process LongProcess) (interface{}, error) {
	r, ok := c.c.Get(key)
	if ok {
		return r, nil
	} else {
		for {
			contained, _ := c.l.ContainsOrAdd(key, 1)
			if !contained {
				defer c.l.Remove(key)
				r, ok := c.c.Get(key)
				if ok {
					return r, nil
				}
				return c.process(process, key)
			} else {
				r, ok := c.c.Get(key)
				if ok {
					return r, nil
				}
			}
		}
	}
}

func (c *Cache) process(process LongProcess, key interface{}) (interface{}, error) {
	value, e := process()
	if e != nil {
		return nil, e
	}
	c.c.Add(key, value)
	return value, nil
}

func (c *Cache) Purge() {
	c.l.Purge()
	c.c.Purge()
}
