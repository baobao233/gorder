package factory

import "sync"

type Supplier func(string) any

type Singleton struct {
	cache    map[string]any
	locker   *sync.Mutex
	supplier Supplier
}

func NewSingleton(supplier Supplier) *Singleton {
	return &Singleton{
		cache:    make(map[string]any), // map中存储的是通过 supplier 根据不同的 conf 实例化出来的不同 redis，由于我们只有一个 redis，因此只有一个 redis
		locker:   &sync.Mutex{},
		supplier: supplier,
	}
}

func (s *Singleton) Get(key string) any {
	if v, hit := s.cache[key]; hit {
		return v
	}
	s.locker.Lock()
	defer s.locker.Unlock()
	if v, hit := s.cache[key]; hit { // double-check
		return v
	}
	s.cache[key] = s.supplier(key)
	return s.cache[key]
}
