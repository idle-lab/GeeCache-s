package geecaches

import (
	"net/http"
	"strings"
)

const defaultBasePath = "/_geecaches/"

type HttpPool struct {
	self string

	opts HttpOptions
}

type HttpOptions struct {
	basePath string
}

func NewHttpPool(ipaddr string) *HttpPool {
	return &HttpPool{
		self: ipaddr,
		opts: HttpOptions{basePath: defaultBasePath},
	}
}

func (h *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse request.
	if !strings.HasPrefix(r.URL.Path, h.opts.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	// /<basepath>/<groupname>/<key> required
	strs := strings.SplitN(r.URL.Path[len(h.opts.basePath):], "/", 2)
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

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}
