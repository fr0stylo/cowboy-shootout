package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type cowboy struct {
	Name   string `json:"name"`
	Health int    `json:"health"`
	Damage int    `json:"damage"`
}

var mutex sync.Mutex

func main() {
	shooterName := flag.String("name", "", "Shooter name")
	redisUrl := flag.String("redis", func() string {
		if url := os.Getenv("REDIS_CONNECTION_STRING"); url != "" {
			return url
		}
		return "localhost:6379"
	}(), "Redis URL")

	flag.Parse()

	if *shooterName == "" {
		log.Fatal("Shooter name is required")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     *redisUrl,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer client.Close()

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatal(err)
	}

	client.Publish(context.Background(), "log", "Shooter "+*shooterName+" is ready to fight")
	pubsub := client.Subscribe(context.Background(), *shooterName)
	go func() {
		for hit := range pubsub.Channel() {
			log.Print(hit.Payload)
			self := getCowboy(client, *shooterName)
			damage, _ := strconv.Atoi(hit.Payload)
			self.Health -= damage
			if self.Health <= 0 {
				mutex.Lock()
				client.Del(context.Background(), *shooterName).Result()
				client.Publish(context.Background(), "log", fmt.Sprintf("[%s] dead", *shooterName))

				os.Exit(0)
			} else {
				buf, _ := json.Marshal(self)
				client.Set(context.Background(), *shooterName, string(buf), 0).Result()
			}
		}
	}()

	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		self := getCowboy(client, *shooterName)

		target := getAttackTarget(client, *shooterName)
		if target == nil {
			client.Publish(context.Background(), "log", fmt.Sprintf("[%s] Wow, no one to attack, I won", *shooterName))
			os.Exit(0)
		}
		client.Publish(context.Background(), *target, strconv.Itoa(self.Damage))
		client.Publish(context.Background(), "log", fmt.Sprintf("[%s] attacking %s", *shooterName, *target))
	}
}

func getAttackTarget(client *redis.Client, name string) *string {
	val, err := client.Keys(context.Background(), "*").Result()
	if err == redis.Nil {
		log.Fatal("Shooter keys not found")
	} else if err != nil {
		panic(err)
	}

	keys := []string{}

	for _, targetName := range val {
		if name != targetName {
			keys = append(keys, targetName)
		}
	}

	if len(keys) == 0 {
		return nil
	}
	if len(keys) == 1 {
		return &keys[0]
	}

	return &keys[rand.Intn(len(keys)-1)]
}

func getCowboy(client *redis.Client, name string) cowboy {
	val, err := client.Get(context.Background(), name).Result()
	if err == redis.Nil {
		log.Fatal("Shooter not found")
	} else if err != nil {
		log.Fatal(err)
	}

	var cowboy cowboy
	if err := json.Unmarshal([]byte(val), &cowboy); err != nil {
		log.Fatal(err)
	}

	return cowboy
}
