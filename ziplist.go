package main

import (
	"fmt"
	"strconv"
  
  "github.com/go-redis/redis"
)

func main()  {

	redisCli := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	redisCli.Del("hello")

	for i := 0; i < 512; i++ {
		redisCli.HSet("hello", strconv.Itoa(i), strconv.Itoa(i))
	}
	encoding, err := redisCli.ObjectEncoding("hello").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("encoding", encoding)

	redisCli.HSet("hello", "512", "512")

	encoding, err = redisCli.ObjectEncoding("hello").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("encoding", encoding)
}


/*
encoding ziplist
encoding hashtable
 */
