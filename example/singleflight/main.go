package main

import (
	"flag"
	"fmt"
	geecaches "geecache-s"
	"geecache-s/cachePolicy"
	"io"
	"log"
	"net/http"
	"time"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *geecaches.Group {
	return geecaches.NewGroup("scores", 2<<10, geecaches.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}), cachePolicy.LruPolicy)
}

func startCacheServer(addr string, addrs []string, gee *geecaches.Group) {
	peers := geecaches.NewHttpPool(addr)
	peers.Set(addrs...)
	gee.RegisterPeers(peers)
	log.Println("geecaches is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gee *geecaches.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func startClient() {
	resp, err := http.Get("http://localhost:9999/api?key=Jack")
	if err != nil {
		log.Fatal(err)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("get data from server: %s", data)
}

func main() {
	var port int
	var api bool
	var tp string
	flag.StringVar(&tp, "type", "serve", "need a run type")
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	if tp == "client" {
		for i := 0; i < 10; i++ {
			go startClient()
		}
		time.Sleep(5 * time.Second)
	} else {
		apiAddr := "http://localhost:9999"
		addrMap := map[int]string{
			8001: "http://localhost:8001",
			8002: "http://localhost:8002",
			8003: "http://localhost:8003",
		}

		var addrs []string
		for _, v := range addrMap {
			addrs = append(addrs, v)
		}

		gee := createGroup()
		if api {
			go startAPIServer(apiAddr, gee)
		}

		startCacheServer(addrMap[port], []string(addrs), gee)
	}
}
