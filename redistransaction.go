package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
)

var incr func(string) error

func main()  {
	redisCli := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	userId := "hyh"
	key := "account:"+ userId
	_, err := redisCli.SetNX(key, 5, 0).Result()
	if err != nil {
		log.Print("redis setnx err", err)
		return
	}
	doubleAccount(redisCli, key)

	fmt.Println(redisCli.Get(key).Result())

}

func doubleAccount(redisCli *redis.Client, key string) error {

	err := redisCli.Watch(func(tx *redis.Tx) error {
		n, err := tx.Get(key).Int()
		if err != nil {
			return err
		}
		_, err = tx.Pipelined(func(pipeliner redis.Pipeliner) error {
			pipeliner.Set(key, n * 2, 0)
			return nil
		})
		if err == redis.TxFailedErr {
			return incr(key)
		}
		return err
	}, key)

	return err
}
