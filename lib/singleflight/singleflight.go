package singleflight

import "sync"

// Package singleflight provides a duplicate function call suppression
// mechanism.
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex // protects mp
	mp map[string]*call
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()
	if g.mp == nil {
		g.mp = make(map[string]*call)
	}

	if c, ok := g.mp[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	nc := new(call)
	nc.wg.Add(1)
	g.mp[key] = nc
	g.mu.Unlock()

	nc.val, nc.err = fn()
	nc.wg.Done()

	g.mu.Lock()
	delete(g.mp, key)
	g.mu.Unlock()

	return nc.val, nc.err
}
