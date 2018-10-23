package db

import (
	"errors"
	"github.com/garyburd/redigo/redis"
)

type RedisCli struct {
	pool *redis.Pool
}

func NewRedisClient() *RedisCli {
	pool := &redis.Pool{
		MaxIdle:   3,
		MaxActive: 3,    //some operation system does not allow too many socket files
		Wait:      true, //if false, redis operation will fail,have to wait
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				panic(err)
				return nil, err
			}
			return c, err
		},
	}
	return &RedisCli{pool: pool}
}

func (rp *RedisCli) GetMap(key string) (map[string]string, error) {
	m, err := redis.StringMap(rp.Do("HGETALL", key))
	if err != nil || len(m) == 0 {
		return nil, errors.New("HGETALL return nothing")
	}
	return m, nil
}

func (rp *RedisCli) StoreMap(key string, m map[string]string) error {
	_, err := rp.Do("HMSET", redis.Args{}.Add(key).AddFlat(m)...)
	return err
}

func (rp *RedisCli) Do(command string, args ...interface{}) (ret interface{}, err error) {
	con := rp.pool.Get()
	err = con.Err()
	if err != nil {
		return
	}
	defer con.Close()
	return con.Do(command, args...)
}
