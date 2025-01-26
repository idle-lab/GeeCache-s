package geecaches

import (
	"fmt"
	"geecache-s/consistenthash"
	pb "geecache-s/geecachespb"
	"hash/crc32"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
)

const (
	defaultBasePath = "/_geecaches/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HttpPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self string

	opts HttpOptions

	mu          sync.Mutex // guards peers and httpGetters
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter // keyed by e.g. "http://10.0.0.2:8008"
}

type HttpOptions struct {
	// BasePath specifies the HTTP path that will serve groupcache requests.
	// If blank, it defaults to "/_geecaches/".
	basePath string

	// Replicas specifies the number of key replicas on the consistent hash.
	// If blank, it defaults to 50.
	Replicas int

	// HashFn specifies the hash function of the consistent hash.
	// If blank, it defaults to crc32.ChecksumIEEE.
	HashFn consistenthash.Hsah
}

// NewHTTPPool initializes an HTTP pool of peers, and registers itself as a PeerPicker.
// For convenience, it also registers itself as an http.Handler with http.DefaultServeMux.
// The self argument should be a valid base URL that points to the current server,
// for example "http://example.net:8000".
func NewHttpPool(self string) *HttpPool {
	p := NewHttpPoolOpts(self, nil)
	http.Handle(p.opts.basePath, p)
	return p
}

// NewHTTPPoolOpts initializes an HTTP pool of peers with the given options.
// Unlike NewHTTPPool, this function does not register the created pool as an HTTP handler.
// The returned *HTTPPool implements http.Handler and must be registered using http.Handle.
func NewHttpPoolOpts(self string, opts *HttpOptions) *HttpPool {
	p := &HttpPool{
		self:        self,
		httpGetters: make(map[string]*httpGetter),
	}
	if opts != nil {
		p.opts = *opts
	}
	if p.opts.basePath == "" {
		p.opts.basePath = defaultBasePath
	}
	if p.opts.HashFn == nil {
		p.opts.HashFn = crc32.ChecksumIEEE
	}
	if p.opts.Replicas == 0 {
		p.opts.Replicas = 50
	}

	return p
}

func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse request.
	if !strings.HasPrefix(r.URL.Path, p.opts.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	// /<basepath>/<groupname>/<key> required
	strs := strings.SplitN(r.URL.Path[len(p.opts.basePath):], "/", 2)
	if len(strs) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	g := GetGroup(strs[0])
	if g == nil {
		http.Error(w, "no such group: "+strs[0], http.StatusNotFound)
	}

	view, err := g.Get(strs[1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// Set updates the pool's list of peers.
// Each peer value should be a valid base URL,
// for example "http://example.net:8000".
func (p *HttpPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(p.opts.Replicas, p.opts.HashFn)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{basePath: peer + p.opts.basePath}
	}
}

func (p *HttpPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	peer := p.peers.Get(key)

	if peer == "" || p.self == peer {
		return nil, false
	}
	if peerGetter, ok := p.httpGetters[peer]; ok {
		log.Printf("Pick peer %s", peer)
		return peerGetter, true
	} else {
		return nil, false
	}
}

func (p *HttpPool) SelfAddr() string {
	return p.self
}

var _ PeerPicker = (*HttpPool)(nil)

type httpGetter struct {
	basePath string
}

func (g *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	url := fmt.Sprintf(
		"%v%v/%v",
		g.basePath,
		in.GetGroup(),
		in.GetKey(),
	)

	resp, err := http.Get(url)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("peer[%s] return %d", g.basePath, resp.StatusCode)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err := proto.Unmarshal(bytes, out); err != nil {
		return nil
	}

	if err != nil {
		return err
	}
	return nil
}

var _ PeerGetter = (*httpGetter)(nil)
