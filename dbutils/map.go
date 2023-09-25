package dbutils

type CollectionMap[K comparable, T any] struct {
	m    map[K]T
	keys []K
}

func NewCollectionMap[K comparable, T any]() *CollectionMap[K, T] {
	return &CollectionMap[K, T]{
		m:    make(map[K]T, 0),
		keys: make([]K, 0),
	}
}

func (c *CollectionMap[K, T]) Set(k K, v T) {
	_, ok := c.m[k]

	c.m[k] = v

	if !ok {
		c.keys = append(c.keys, k)
	}
}

func (c *CollectionMap[K, T]) Has(k K) bool {
	_, ok := c.m[k]

	return ok
}

func (c *CollectionMap[K, T]) Get(k K) T {
	return c.m[k]
}

// Returns underline slice in order.
func (c *CollectionMap[K, T]) All() []T {
	var results []T

	for _, k := range c.keys {
		results = append(results, c.m[k])
	}

	return results
}
