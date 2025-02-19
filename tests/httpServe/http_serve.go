package main

import (
	"fmt"
	geecaches "geecache-s"
	"geecache-s/lib/cachePolicy"
	"log"
	"net/http"
)

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	geecaches.NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}), cachePolicy.LruPolicy)

	addr := "localhost:8080"
	peers := geecaches.NewHttpPool(addr)
	log.Println("geecaches is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
