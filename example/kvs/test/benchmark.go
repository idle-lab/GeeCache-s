package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	serverURL   = "http://localhost:9999"
	numRequests = 10000 // Total requests
	concurrency = 50    // Concurrent workload
)

type kvPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func sendPostRequest(key, value string) error {
	data, _ := json.Marshal(kvPair{Key: key, Value: value})
	resp, err := http.Post(serverURL+"/add", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	return nil
}

func sendGetRequest(key string) error {
	resp, err := http.Get(fmt.Sprintf("%s/get?key=%s", serverURL, key))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	return nil
}

func benchmarkRequests(testFunc func(int), name string) {
	start := time.Now()
	testFunc(numRequests)

	duration := time.Since(start)
	fmt.Printf("%s took: %v, QPS: %.2f\n", name, duration, float64(numRequests)/duration.Seconds())
}

func concurrentRequests(num int, requestFunc func(int)) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	for i := 0; i < num; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			requestFunc(i)
			<-sem
		}(i)
	}
	wg.Wait()
}

func main() {
	// Write Test
	fmt.Println("Starting Write Test...")
	benchmarkRequests(func(num int) {
		concurrentRequests(num, func(i int) {
			sendPostRequest(fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i))
		})
	}, "Write Test")

	// Read Test
	fmt.Println("Starting Read Test...")
	benchmarkRequests(func(num int) {
		concurrentRequests(num, func(i int) {
			_ = sendGetRequest(fmt.Sprintf("key-%d", i))
		})
	}, "Read Test")
}
