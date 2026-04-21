package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

var decrIfPositive = redis.NewScript(`
	local stock = tonumber(redis.call('GET', KEYS[1]))
	if stock <= 0 then
		return -1
	end
	return redis.call('DECR', KEYS[1])
`)

const stockKey = "stock:flash-sale-item-001"

func main() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Invalid REDIS_URL: %v", err)
	}
	rdb = redis.NewClient(opt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer rdb.Close()
	log.Println("Connected to Redis")

	r := gin.Default()
	r.POST("/buy", buyHandler)
	r.GET("/stock", stockHandler)

	fmt.Println("Server starting on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func buyHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	result, err := decrIfPositive.Run(ctx, rdb, []string{stockKey}).Int()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if result == -1 {
		c.JSON(http.StatusConflict, gin.H{"status": "failed", "message": "Product is sold out!"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "Purchase successful!",
		"remaining": result,
	})
}

func stockHandler(c *gin.Context) {
	val, err := rdb.Get(c.Request.Context(), stockKey).Int64()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"available": val,
	})
}
