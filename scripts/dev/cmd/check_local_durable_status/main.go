package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	baseURL := os.Getenv("OPS_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8088"
	}

	printJSON("Durable summary", fetchJSON(baseURL+"/v1/ops/durable/summary"))
}

func fetchJSON(url string) any {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("GET %s returned status %d", url, resp.StatusCode)
	}

	var payload any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		log.Fatalf("decode %s failed: %v", url, err)
	}
	return payload
}

func printJSON(label string, payload any) {
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Fatalf("marshal %s failed: %v", label, err)
	}
	fmt.Printf("%s\n%s\n", label, string(raw))
}
