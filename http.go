package geecaches

import (
	"bytes"
	"fmt"
	"geecache-s/consistenthash"
	pb "geecache-s/geecachespb"
	"hash/crc32"
	"io"
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

	mu           sync.Mutex // guards peers and httpHandlers
	peers        *consistenthash.Map
	httpHandlers map[string]*httpHandler // keyed by e.g. "http://10.0.0.2:8008"
}

type HttpOptions struct {
	// BasePath specifies the HTTP path that will serve groupcache requests.
	// defaults: "/_geecaches".
	BasePath string

	// Replicas specifies the number of key replicas on the consistent hash.
	// defaults: 50.
	Replicas int

	// HashFn specifies the hash function of the consistent hash.
	// defaults: crc32.ChecksumIEEE.
	HashFn consistenthash.Hsah
}

func NewHttpPoolOptions() *HttpOptions {
	return &HttpOptions{
		BasePath: "/_geecaches",
		Replicas: 50,
		HashFn:   crc32.ChecksumIEEE,
	}
}

// NewHTTPPool initializes an HTTP pool of peers, and registers itself as a PeerPicker.
// For convenience, it also registers itself as an http.Handler with http.DefaultServeMux.
// The self argument should be a valid base URL that points to the current server,
// for example "http://example.net:8000".
func NewHttpPool(self string) *HttpPool {
	p := NewHttpPoolWithOpts(self, nil)
	http.Handle(p.opts.BasePath, p)
	return p
}

// NewHTTPPoolOpts initializes an HTTP pool of peers with the given options.
// Unlike NewHTTPPool, this function does not register the created pool as an HTTP handler.
// The returned *HTTPPool implements http.Handler and must be registered using http.Handle.
func NewHttpPoolWithOpts(self string, opts *HttpOptions) *HttpPool {
	p := &HttpPool{
		self:         self,
		httpHandlers: make(map[string]*httpHandler),
	}
	if opts != nil {
		p.opts = *opts
	}
	if p.opts.BasePath == "" {
		p.opts.BasePath = defaultBasePath
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
	if !strings.HasPrefix(r.URL.Path, p.opts.BasePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	if r.Method == "GET" {
		// /<basepath>/<groupname>/<key> required
		strs := strings.SplitN(r.URL.Path[len(p.opts.BasePath):], "/", 2)
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

		body, err := proto.Marshal(&pb.GetResponse{Value: view.ByteSlice()})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(body)
	} else if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req := pb.AddRequest{}
		if err = proto.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		g := GetGroup(req.Group)

		err = g.Add(req.Key, ByteView{req.Value})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Set updates the pool's list of peers.
// Each peer value should be a valid base URL,
// for example "http://example.net:8000".
func (p *HttpPool) SetPeers(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(p.opts.Replicas, p.opts.HashFn)
	p.peers.Add(peers...)
	p.httpHandlers = make(map[string]*httpHandler, len(peers))
	for _, peer := range peers {
		p.httpHandlers[peer] = &httpHandler{basePath: peer + p.opts.BasePath}
	}
}

func (p *HttpPool) PickPeer(key string) (PeerHandler, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	peer := p.peers.Get(key)

	if peer == "" || p.self == peer {
		return nil, false
	}
	if httpHandler, ok := p.httpHandlers[peer]; ok {
		return httpHandler, true
	} else {
		return nil, false
	}
}

func (p *HttpPool) SelfAddr() string {
	return p.self
}

var _ PeerPicker = (*HttpPool)(nil)

type httpHandler struct {
	basePath string
}

// remote Get
func (g *httpHandler) Get(in *pb.GetRequest, out *pb.GetResponse) error {
	url := fmt.Sprintf(
		"%v/%v/%v",
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

// remote Add
func (g *httpHandler) Add(in *pb.AddRequest, out *pb.Empty) error {
	url := g.basePath

	body, err := proto.Marshal(in)
	if err != nil {
		return err
	}

	br := bytes.NewReader(body)

	resp, err := http.Post(url, "application/octet-stream", br)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("%s", resp.Body)
	}
	return nil
}

var _ PeerHandler = (*httpHandler)(nil)
