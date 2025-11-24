package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var (
	totalRequests uint64
	totalErrors   uint64
	totalLatency  uint64 
)

func main() {
	url := flag.String("url", "http://localhost:8080/kv/", "Target URL")
	clients := flag.Int("clients", 10, "Number of threads")
	duration := flag.Int("duration", 60, "Test duration in seconds")
	workload := flag.String("workload", "get_popular", "Workload: put_all, get_all, get_popular, mixed")
	flag.Parse()

	fmt.Printf("Starting %s Test: %d clients, %ds\n", *workload, *clients, *duration)

	var wg sync.WaitGroup
	stopTime := time.Now().Add(time.Duration(*duration) * time.Second)
	startTest := time.Now()

	for i := 0; i < *clients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			client := &http.Client{
				Transport: &http.Transport{
					MaxIdleConnsPerHost: 100, 
				},
				Timeout: 10 * time.Second,
			}

			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))

			for time.Now().Before(stopTime) {
				req, _ := generateRequest(r, *workload, *url, id)
				
				t0 := time.Now()
				resp, err := client.Do(req)
				lat := time.Since(t0).Microseconds()

				if err != nil {
					atomic.AddUint64(&totalErrors, 1)
				} else {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
					if resp.StatusCode >= 400 && resp.StatusCode != 404 {
						atomic.AddUint64(&totalErrors, 1)
					} else {
						atomic.AddUint64(&totalRequests, 1)
						atomic.AddUint64(&totalLatency, uint64(lat))
					}
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(startTest).Seconds()
	printMetrics(elapsed)
}

func generateRequest(r *rand.Rand, workload, baseURL string, id int) (*http.Request, error) {
	method := "GET"
	targetUrl := baseURL
	var body []byte

	switch workload {
	case "put_all":
		method = "POST"
		key := fmt.Sprintf("w-%d-%d", id, r.Intn(1000000))
		dummyData := make([]byte, 1024)
		for k := range dummyData {
			dummyData[k] = 'x' 
		}
		payload, _ := json.Marshal(map[string]string{
			"key":   key,
			"value": string(dummyData),
		})
		body = payload

	case "get_all":
		method = "GET"
		targetUrl += fmt.Sprintf("miss-%d-%d", id, r.Intn(10000000))

	case "get_popular":
		method = "GET"
		targetUrl += fmt.Sprintf("popular-%d", r.Intn(100))

	case "mixed":
		if r.Float32() < 0.5 {
			method = "GET"
			targetUrl += fmt.Sprintf("mixed-%d-%d", id, r.Intn(10000))
		} else {
			method = "POST"
			dummyData := make([]byte, 100) 
			for k := range dummyData {
				dummyData[k] = 'y'
			}

			payload, _ := json.Marshal(map[string]string{
				"key":   fmt.Sprintf("mixed-%d-%d", id, r.Intn(10000)),
				"value": string(dummyData),
			})
			body = payload
		}
	}

	req, err := http.NewRequest(method, targetUrl, bytes.NewBuffer(body))
	return req, err
}

func printMetrics(elapsed float64) {
	reqs := atomic.LoadUint64(&totalRequests)
	lat := atomic.LoadUint64(&totalLatency)
	
	throughput := float64(reqs) / elapsed
	avgLat := float64(lat) / float64(reqs) / 1000.0
	fmt.Printf("throughput,%.2f\n", throughput)
	fmt.Printf("latency,%.2f\n", avgLat)
}
