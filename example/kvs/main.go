package main

import (
	"encoding/json"
	"flag"
	geecaches "geecache-s"
	"geecache-s/lib/cachePolicy"
	"io"
	"log"
	"net/http"
)

var addrMap = map[int]string{
	8001: "http://localhost:8001",
	8002: "http://localhost:8002",
	8003: "http://localhost:8003",
}

var apiAddr = "localhost:9999"

func startAPIServer(g *geecaches.Group) {
	http.Handle("/get", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// eg: http://localhost:9999/get?key=Jack
			key := r.URL.Query().Get("key")
			bytes, err := g.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(bytes.ByteSlice())
		} else {
			http.Error(w, "unable to process the request", http.StatusBadRequest)
		}
	}))

	http.Handle("/add", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// eg: http://localhost:9999/add
			// Use your preferred data format in http body. I use JSON here.
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			req := make(map[string]any)
			err = json.Unmarshal(body, &req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = g.Add(req["key"].(string), geecaches.ByteView{Bytes: []byte(req["value"].(string))})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "unable to process the request", http.StatusBadRequest)
		}
	}))

	log.Printf("fontend server is running at %s", apiAddr)
	log.Fatal(http.ListenAndServe("localhost:9999", nil))
}

func main() {
	var (
		port int
		api  bool
	)
	flag.IntVar(&port, "port", -1, "Geecache-s server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	if port == -1 {
		log.Fatalf("invalid port %d", port)
	}
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	addr := addrMap[port]

	// create kvs group
	gopts := geecaches.NewGroupOptions()
	gopts.CachePolicy = cachePolicy.LfuPolicy
	gopts.MaxBytes = 4 * 1024 * 1024 * 1024 // 4GB
	g := geecaches.NewGroupWithOpts("kvs", gopts)

	popts := geecaches.NewHttpPoolOptions()
	popts.BasePath = "/_kvs"
	p := geecaches.NewHttpPoolWithOpts(addr, popts)
	http.Handle("/_kvs", p)
	p.SetPeers(addrs...)

	g.RegisterPeers(p)

	if api {
		// start api serve
		go startAPIServer(g)
	}

	// start kvs cache serve
	log.Printf("kvs peer is running at %s\n", addr[7:])
	log.Fatal(http.ListenAndServe(addr[7:], p))
}
