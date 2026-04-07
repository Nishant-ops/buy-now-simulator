package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Inventory struct {
	mu         sync.Mutex
	TotalStock int64
	SoldStock  int64
}

var inventory = Inventory{
	TotalStock: 100, // example stock for the flash sale
	SoldStock:  0,
}

func main() {
	// Standard library router
	http.HandleFunc("/buy", buyHandler)
	http.HandleFunc("/stock", stockHandler)

	fmt.Println("Server starting on :8080...")
	fmt.Println("Total Flash Sale Stock:", inventory.TotalStock)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// buyHandler simulates the Flipkart "Buy Now" button logic
func buyHandler(w http.ResponseWriter, r *http.Request) {
	// Typically this would be a POST request as it's an action that changes state
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed, use POST", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Lock the inventory to prevent race conditions during concurrent flash sale requests
	inventory.mu.Lock()
	defer inventory.mu.Unlock()

	// 1. Check if we still have stock
	if inventory.SoldStock >= inventory.TotalStock {
		// Sold out!
		w.WriteHeader(http.StatusConflict) // 409 Conflict
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "failed",
			"message": "Product is sold out!",
		})
		return
	}

	// 2. Process purchase
	inventory.SoldStock++
	
	// Simulate some work like DB updates, user validation, queueing the order
	// time.Sleep(10 * time.Millisecond)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"message":    "Purchase successful!",
		"order_id":   fmt.Sprintf("ORDID%04d", inventory.SoldStock),
	})
}

// stockHandler allows us to check the remaining stock easily
func stockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	inventory.mu.Lock()
	defer inventory.mu.Unlock()

	available := inventory.TotalStock - inventory.SoldStock
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_stock": inventory.TotalStock,
		"sold_stock":  inventory.SoldStock,
		"available":   available,
	})
}
