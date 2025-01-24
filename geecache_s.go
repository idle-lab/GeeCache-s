package geecaches

import (
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

var (
	groupsMut sync.RWMutex
	groups    = make(map[string]*Group)
)

func NewGroup(name string, maxBytes int64, getter Getter) *Group {
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			maxBytes: maxBytes,
		},
	}

	groupsMut.Lock()
	groups[name] = g
	groupsMut.Unlock()

	return g
}

func GetGroup(name string) *Group {
	groupsMut.RLock()
	g := groups[name]
	groupsMut.RUnlock()
	return g
}

// A Group is a cache namespace and associated data loaded spread over
// a group of 1 or more machines.
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Get(key string) (ByteView, error) {
	if value, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache-s hit]")
		return value, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	return g.loadLocally(key)
}

func (g *Group) loadLocally(key string) (ByteView, error) {
	value, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	v := ByteView{value}
	g.mainCache.add(key, v)
	return v, nil
}
