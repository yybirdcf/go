package tcpserver

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Group interface {
	GetMembers(id int64) ([]int64, error)
}

var (
	KEY_PREFIX_GROUP_MEMBERS = "group#members#"
)

type RedisGroup struct {
	pool *redis.Pool
}

func NewRedisGroup(host string, pwd string, db int) *RedisGroup {
	return &RedisGroup{
		pool: &redis.Pool{
			MaxIdle:     5,
			IdleTimeout: 300 * time.Second,
			// Other pool configuration not shown in this example.
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", host)
				if err != nil {
					return nil, err
				}
				if _, err := c.Do("AUTH", pwd); err != nil {
					c.Close()
					return nil, err
				}
				if _, err := c.Do("SELECT", db); err != nil {
					c.Close()
					return nil, err
				}
				return c, nil
			},
		},
	}
}

func (g *RedisGroup) GetMembers(id int64) ([]int64, error) {
	conn := g.pool.Get()
	defer conn.Close()

	key := fmt.Sprintf("%s%d", KEY_PREFIX_GROUP_MEMBERS, id)
	return Int64s(conn.Do("SMEMBERS", key))
}

func Int64s(reply interface{}, err error) ([]int64, error) {
	var int64s []int64
	values, err := redis.Values(reply, err)
	if err != nil {
		return int64s, err
	}
	if err := redis.ScanSlice(values, &int64s); err != nil {
		return int64s, err
	}
	return int64s, nil
}
