package main

import (
	"fmt"
	"github.com/go-redis/redis"
)

func NewRedisClient(addr string) *redis.Client {

	return redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
}

func main() {
	client1 := NewRedisClient("127.0.0.1:6379")

	pubsub := client1.Subscribe("channel1")
	defer pubsub.Close()

	client2 := NewRedisClient("127.0.0.1:6379")
	err := client2.Publish("channel1", "hello").Err()
	if err != nil {
		panic(err)
	}

	msg, err := pubsub.ReceiveMessage()
	if err != nil {
		panic(err)
	}
	fmt.Println(msg.Channel, msg.Payload)
}
