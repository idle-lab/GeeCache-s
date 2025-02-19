package geecaches

import (
	"fmt"
	"geecache-s/lib/cachePolicy"
	pb "geecache-s/lib/geecachespb"
	"geecache-s/lib/singleflight"
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

type GroupOptions struct {
	// The caching policy of this group
	// Default: LRU
	CachePolicy cachePolicy.CachePolicy

	// If the cache does not contain the specified key and that key is managed by the current node,
	// the Getter function is automatically invoked to fetch its corresponding value.
	// If the user does not specify this function, the Get() function will return an error when the cache misses
	// Default: nil
	Getter Getter

	// The maximum number of bytes that a cache can occupy in memory.
	// When the value is 0, there is no limit on memory usage.
	// Default: 0
	MaxBytes int64
}

func NewGroupOptions() *GroupOptions {
	return &GroupOptions{
		CachePolicy: cachePolicy.LruPolicy,
	}
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
	if g, ok := groups[name]; ok {
		return g
	}
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

func NewGroupWithOpts(name string, opts *GroupOptions) *Group {
	if g, ok := groups[name]; ok {
		return g
	}
	g := &Group{
		name:   name,
		getter: opts.Getter,
		mainCache: cache{
			maxBytes: opts.MaxBytes,
			policy:   opts.CachePolicy,
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
		panic("registerPeerPicker called more than once")
	}
	g.peersPicker = peersPicker
}

func (g *Group) Get(key string) (ByteView, error) {
	if value, ok := g.mainCache.get(key); ok {
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
	if g.getter == nil {
		return ByteView{}, fmt.Errorf("no getter specified, unable to retrieve data")
	}
	value, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	v := ByteView{value}
	g.mainCache.add(key, v)
	return v, nil
}

func (g *Group) loadRemotely(key string, peer PeerHandler) (ByteView, error) {
	in := &pb.GetRequest{
		Group: g.Name(),
		Key:   key,
	}
	out := &pb.GetResponse{}
	err := peer.Get(in, out)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{out.Value}, nil
}

func (g *Group) Add(key string, value ByteView) error {
	if peer, ok := g.peersPicker.PickPeer(key); ok {
		in := &pb.AddRequest{
			Group: g.name,
			Key:   key,
			Value: value.Bytes,
		}
		empty := &pb.Empty{}
		// add remotely
		if err := peer.Add(in, empty); err != nil {
			return err
		}
	}

	// add locally
	return g.mainCache.add(key, value)
}
