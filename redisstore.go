package myredis

import (
	"errors"
	"github.com/FZambia/sentinel"
	rs "github.com/gin-contrib/sessions/redis"
	"github.com/gomodule/redigo/redis"
	"time"
	"os"
)

func newSentinelPool() *redis.Pool {
	redisHost := os.Getenv("REDIS_HOST")
	redisPwd := os.Getenv("REDIS_PWD")
	sntnl := &sentinel.Sentinel{
		Addrs:      []string{redisHost+":26379", redisHost+":26380", redisHost+":26381"},
		MasterName: "mymaster",
		Dial: func(addr string) (redis.Conn, error) {
			timeout := 500 * time.Millisecond
			c, err := redis.Dial("tcp", addr,
				redis.DialPassword(redisPwd),
				redis.DialConnectTimeout(timeout), redis.DialReadTimeout(timeout), redis.DialWriteTimeout(timeout))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
	return &redis.Pool{
		MaxIdle:     3,
		MaxActive:   64,
		Wait:        true,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			masterAddr, err := sntnl.MasterAddr()
			if err != nil {
				return nil, err
			}
			c, err := redis.Dial("tcp", masterAddr)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if !sentinel.TestRole(c, "master") {
				return errors.New("Role check failed")
			} else {
				return nil
			}
		},
	}
}

func InitStore() rs.Store {
	store, err := rs.NewStoreWithPool(newSentinelPool(), []byte(os.Getenv("SESSION_SECRET")))
	if err != nil {
		panic(err)
	}

	return store
}
