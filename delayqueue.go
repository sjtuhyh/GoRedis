package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func main()  {
	redisCli := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	queue := RedisDelayingQueue{
		redisCli: redisCli,
		queueKey: "queue-demo",
	}

	go func() {
		for i := 0; i < 100; i ++ {
			queue.delay("hyh" + strconv.Itoa(i))
		}
	}()

	go func() {
		queue.loop()
	}()
}

type RedisDelayingQueue struct {
	redisCli *redis.Client
	queueKey string
}

type Task struct {
	Id string `json:"id"`
	Msg string `json:"msg"`
}

func (r *RedisDelayingQueue) delay(msg string) error {

	id := strconv.FormatInt(time.Now().UnixNano(), 10)
	task := Task{Id: id, Msg: msg}
	t, err := json.Marshal(task)
	if err != nil {
		return err
	}
	_, err =r.redisCli.ZAdd(r.queueKey, &redis.Z{Score: float64(time.Now().UnixNano()/1000000 + 5000), Member: t}).Result()
	return err
}

func (r *RedisDelayingQueue) loop() {
	for {
		values, err := r.redisCli.ZRangeByScore(r.queueKey, &redis.ZRangeBy{Min: "-inf", Max: "+inf"}).Result()
		if err != nil {
			time.Sleep(time.Millisecond* 500)
			continue
		}
		_, err = r.redisCli.ZRem(r.queueKey, values[0]).Result()
		if err != nil {
			var task Task
			json.Unmarshal([]byte(values[0]), &task)
			r.handleMsg(task.Msg, task.Id)
		}
	}
}

func (r *RedisDelayingQueue) handleMsg(msg string, id string) {
	fmt.Println(msg, id)
}