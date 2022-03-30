package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type cowboy struct {
	Name   string `json:"name"`
	Health int8   `json:"health"`
	Damage int8   `json:"damage"`
}

var wg sync.WaitGroup

var (
	executablePath = flag.String("executable", os.Getenv("EXECUTABLE_PATH"), "Path to executable")
	shootersPath   = flag.String("shooters", "../shooters.json", "Path to executable")
	redisUrl       = flag.String("redis", func() string {
		if url := os.Getenv("REDIS_CONNECTION_STRING"); url != "" {
			return url
		}
		return "localhost:6379"
	}(), "Redis URL")
)

func main() {
	flag.Parse()
	client := redis.NewClient(&redis.Options{
		Addr:     *redisUrl,
		Password: "",
		DB:       0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatal(err)
	}

	if err := setShootersToRedis(client); err != nil {
		log.Fatal(err)
	}

	keys, err := client.Keys(context.Background(), "*").Result()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		logs := client.Subscribe(context.Background(), "log")
		for logEntry := range logs.Channel() {
			log.Printf("%s", logEntry.Payload)
		}
	}()

	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			cmd := exec.Command(*executablePath, "--name", k)
			cmd.Env = append(cmd.Env, "REDIS_CONNECTION_STRING="+*redisUrl)

			cmd.Run()

			wg.Done()
		}(key)
	}

	wg.Wait()
}

func setShootersToRedis(client *redis.Client) error {
	var shooters []cowboy
	fd, err := os.Open(*shootersPath)
	if err != nil {
		log.Fatal("Failed to open file: ", err)
	}
	defer fd.Close()

	if err := json.NewDecoder(fd).Decode(&shooters); err != nil {
		return errors.Wrap(err, "Failed to decode file: ")
	}

	for _, shooter := range shooters {
		var bytes []byte

		if bytes, err = json.Marshal(shooter); err != nil {
			return errors.Wrap(err, "Failed to marshal: ")
		}
		client.Set(context.Background(), shooter.Name, string(bytes), 0)
	}

	return nil
}
