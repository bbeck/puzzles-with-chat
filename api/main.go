package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"

	"github.com/bbeck/twitch-plays-crosswords/api/crossword"
)

func main() {
	pool := NewRedisPool()
	defer func() { _ = pool.Close() }()

	router := gin.Default()

	// Register handlers for our paths.
	api := router.Group("/api")
	crossword.RegisterRoutes(api, pool)

	// Start the server.
	err := router.Run(":5000")
	if err != nil {
		log.Fatalf("error from main: %+v", err)
	}
}

func NewRedisPool() *redis.Pool {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = ":6379"
	}

	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 300 * time.Second,

		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", host)
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
