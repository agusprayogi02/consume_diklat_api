package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/panjf2000/ants/v2"
)

func fetchURL(baseURL string, queryParams map[string]string, token string) string {
	fmt.Println("Fetching ke-", queryParams["batch"])
	// Build the URL with query parameters
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Sprintf("Error parsing URL %s: %v", baseURL, err)
	}

	// Add query parameters to the URL
	q := u.Query()
	for key, value := range queryParams {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	// Create a new request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return fmt.Sprintf("Error creating request for %s: %v", u.String(), err)
	}

	// Add Bearer token to the request header
	req.Header.Set("Authorization", "Bearer "+token)

	// Perform the request
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Sprintf("Error fetching %s: %v", u.String(), err)
	}
	defer response.Body.Close()

	// Read and process the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response body from %s: %v", u.String(), err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Sprintf("Error parsing JSON from %s: %v", u.String(), err)
	}

	return fmt.Sprintf("Response from %s: %+v", u.String(), result)
}

func main() {
	// Define base URL and query parameters
	baseURL := "http://10.1.1.122:3000/api/hms-ptk-index/import"
	token := "eyJhbGciOiJQUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MywiZW1haWwiOiJib2VAZ21haWwuY29tIiwicm9sZSI6ImFkbWluIiwidXB0X2lkIjoyLCJ1cHRfbmFtZSI6IkJCUFBNUFYgQk9FIE1hbGFuZyIsImlzcyI6IlN0YXJ0ZXItR29maWJlciIsImV4cCI6MTcyNjA3MzUxM30.UjqIrUwb0eV5jCU1a46kcWNtr3BzD1elWeSAe79F1t7kJz7kslYmh5EkAib350PYFenHsNAI3P3_3OYU_5xGRq_n81SjvmhDFgtq_CAplTugBmQlUIM7HkqF6fIolfS9R4MlaWBGSJ6DX0Qf5igihNWQZLDNgO0aDYceReJ-oquQpCCISuJfUJjM8jx4TTR9SxPygcAhUwtAffpQ9RjSEG_YBeHFscqnPKfMNFcAf2E2SWqSdHfrnt8r1HuzFj4es13KV2DqyGr08dr9d0a4dF9cE5dJr--lznLklaL92tLSslCJobrVXQqitOAOmEL-GQMpa3FCAQKObcxGGzVzFZ1n0f8CTHbb9e8pTgOAUCoEDJD9TyVkRYx-JGwlswyaj7mWIaVMspNqS5lPLThUp-hqYB6yB3tqkLxHKqgYl_3Y-1KQLKHcY7QQhBsb7HRKbKJSb9b3mZ2D5sfW-VAXnuQ-vlOQrIx0G13QCB8XjUpP0aTNPlXYQnIa3Mm4oiX5FTTZayh9j9oTDlyWm3j-BKmgpomJaSUFP1x7JCqFwQCoPKH2cav2eoG3qjP83SCs8enkyQK_fpG1M1sSloauqqW2enhg7lnzVS0ck4BmHwLXQXw_vn2cIv1S4tjaC14tJGctA9yOo39fK50RWWI34RB9B9Cd6anwO-4miM9Qreg" // Replace with your actual Bearer token

	// Create a slice of 100 URLs
	start := 128
	batches := make([]int, 30)
	for i := 0; i < 30; i++ {
		batches[i] = start + i
	}

	// Create a goroutine pool with a maximum of 6 concurrent workers
	pool, err := ants.NewPool(6)
	if err != nil {
		log.Fatalf("Error creating pool: %v", err)
	}
	defer pool.Release()

	var wg sync.WaitGroup
	results := make(chan string, len(batches))

	for _, b := range batches {
		wg.Add(1)
		batch := b // capture range variable
		_ = pool.Submit(func() {
			defer wg.Done()
			result := fetchURL(baseURL, map[string]string{
				"per_batch": "4000",
				"batch":     fmt.Sprintf("%d", batch),
			}, token)
			results <- result
		})
	}

	// Close the results channel once all tasks are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process the results
	for result := range results {
		fmt.Println(result)
	}
}
