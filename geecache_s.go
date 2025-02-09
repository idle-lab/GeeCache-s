package geecaches

import (
	"geecache-s/cachePolicy"
	pb "geecache-s/geecachespb"
	"geecache-s/singleflight"
	"log"
	"sync"
)

//	Main process of GeeCache-s:
// 											Yes
// Receive key --> Check if cached ----------------------> Return cached value ⑴
//                    | No                                  | Yes
//                    |-------> Should fetch from remote node? --------> Interact with remote node --> Return cached value ⑵
//                                | No
//                                |-------> Call `callback function`, retrieve value, and add to cache --> Return cached value ⑶

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	groupsMut sync.RWMutex
	groups    = make(map[string]*Group)
)

func GetGroup(name string) *Group {
	groupsMut.RLock()
	g := groups[name]
	groupsMut.RUnlock()
	return g
}

// A Group is a cache namespace and associated data loaded spread over
// a group of 1 or more machines.
type Group struct {
	name string

	getter    Getter
	mainCache cache

	peersPicker PeerPicker

	loader *singleflight.Group
}

func NewGroup(name string, maxBytes int64, getter Getter, policy cachePolicy.CachePolicy) *Group {
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			maxBytes: maxBytes,
			policy:   policy,
		},
		loader: &singleflight.Group{},
	}

	groupsMut.Lock()
	groups[name] = g
	groupsMut.Unlock()

	return g
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) RegisterPeers(peersPicker PeerPicker) {
	if g.peersPicker != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peersPicker = peersPicker
}

func (g *Group) Get(key string) (ByteView, error) {
	if value, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache-s hit]")
		return value, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.
	bytes, err := g.loader.Do(key, func() (any, error) {
		if g.peersPicker != nil {
			peerGetter, ok := g.peersPicker.PickPeer(key)
			if ok {
				return g.loadRemotely(key, peerGetter)
			}
		}

		return g.loadLocally(key)
	})
	return bytes.(ByteView), err
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

func (g *Group) loadRemotely(key string, peerGetter PeerGetter) (ByteView, error) {
	in := &pb.Request{
		Group: g.Name(),
		Key:   key,
	}
	out := &pb.Response{}
	err := peerGetter.Get(in, out)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{out.Value}, nil
}
