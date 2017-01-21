package usersvc

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/garyburd/redigo/redis"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"golang.org/x/net/context"

	"go/microservice/services/config"
	"go/microservice/services/models"
)

// Service describes a service that adds things together.
type Service interface {
	GetUserinfo(ctx context.Context, id int64) (Userinfo, error)
}

var (
	CacheTime              = 30 * 86400
	CacheUserPrefix        = "app#userinfo#"
	CacheUserFriendsPrefix = "app#userfriends#"
	CacheUserBlocksPrefix  = "app#userblocks#"
	RedisPubSubCmdChannel  = "app#pubsubcmdchannel"
)

type redisCmd struct {
	Node        string
	Cmd         string
	KeysAndArgs []interface{}
}

type basicService struct {
	cfg      *config.Config
	pool     *redis.Pool
	syncPool *redis.Pool
	node     string
}

func NewBasicService(cfg *config.Config, node string) Service {
	pool := &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 300 * time.Second,
		// Other pool configuration not shown in this example.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", cfg.REDIS_HOST_app)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", cfg.REDIS_PWD_app); err != nil {
				c.Close()
				return nil, err
			}
			if _, err := c.Do("SELECT", cfg.REDIS_DB_app); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}

	syncPool := &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 300 * time.Second,
		// Other pool configuration not shown in this example.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", cfg.REDIS_HOST_pubsub)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", cfg.REDIS_PWD_pubsub); err != nil {
				c.Close()
				return nil, err
			}
			if _, err := c.Do("SELECT", cfg.REDIS_DB_pubsub); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}

	bs := basicService{
		cfg:      cfg,
		pool:     pool,
		syncPool: syncPool,
		node:     node,
	}

	bs.syncEvents()
	return bs
}

func (s basicService) publish(cmd string, key string, args ...interface{}) {
	kvs := make([]interface{}, len(args)+1)
	kvs[0] = key

	for _, p := range args {
		kvs = append(kvs, p)
	}

	rc := redisCmd{
		Node:        s.node,
		Cmd:         cmd,
		KeysAndArgs: kvs,
	}

	cmdStr, err := json.Marshal(&rc)
	if err != nil {
		fmt.Printf("pubsub publish json marshal error: %v\n", err)
	}

	conn := s.syncPool.Get()
	defer conn.Close()

	_, err = conn.Do("PUBLISH", RedisPubSubCmdChannel, string(cmdStr))
	if err != nil {
		fmt.Printf("%v, cmd=%s\n", err, string(cmdStr))
	}
}

//同步redis缓存数据
func (s basicService) syncEvents() {
	go func() {
		syncconn := s.syncPool.Get()
		defer syncconn.Close()

		psc := redis.PubSubConn{Conn: syncconn}
		psc.Subscribe(RedisPubSubCmdChannel)

		conn := s.pool.Get()
		defer conn.Close()

		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				if v.Channel == RedisPubSubCmdChannel {
					var rc redisCmd
					err := json.Unmarshal(v.Data, &rc)
					if err == nil && rc.Node != s.node {
						_, err = conn.Do(rc.Cmd, rc.KeysAndArgs...)
						if err != nil {
							fmt.Printf("pubsub do cmd error: %v\n", err)
						}
					}
				}
			case error:
				fmt.Printf("pubsub error: %v\n", v)
			}
		}
	}()
}

func (s basicService) GetUserinfo(ctx context.Context, id int64) (Userinfo, error) {
	userinfo := Userinfo{}
	//判断redis缓存
	conn := s.pool.Get()
	defer conn.Close()

	key := fmt.Sprintf("%s%d", CacheUserPrefix, id)
	data, _ := redis.String(conn.Do("GET", key))
	if data != "" {
		err := json.Unmarshal([]byte(data), &userinfo)
		if err == nil {
			return userinfo, nil
		}
	}

	db, err := gorm.Open("mysql", s.cfg.DB_app)
	if err != nil {
		return userinfo, err
	}
	defer db.Close()

	//获取用户基础信息
	var user models.ModeGame
	err = db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return userinfo, err
	}

	userinfo = Userinfo{
		Id:             int64(user.Id),
		Username:       user.Username,
		Phone:          user.Phone,
		Sex:            int64(user.Sex),
		Avatar:         user.Avatar,
		Gouhao:         int64(user.Gouhao),
		Birthday:       int64(user.Birthday),
		Avatars:        user.Avatars,
		Signature:      user.Signature,
		Appfrom:        user.Appfrom,
		Appver:         user.Appver,
		BackgroudImage: user.BackgroudImage,
		UpdateAppver:   user.UpdateAppver,
		Privacy:        int64(user.Privacy),
		LoadRecTags:    int64(user.LoadRecTags),
		GamePower:      int64(user.GamePower),
		Mark:           int64(user.Mark),
		Level:          int64(user.Level),
		QuestionPhoto:  user.QuestionPhoto,
		Lan:            user.Lan,
		Notify:         int64(user.Notify),
		ImToken:        user.AccessToken,
	}

	bytes, err := json.Marshal(&userinfo)
	if err == nil {
		conn.Do("SETEX", key, CacheTime, string(bytes))
	}
	return userinfo, nil
}

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

// ServiceLoggingMiddleware returns a service middleware that logs the
// parameters and result of each method invocation.
func ServiceLoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return serviceLoggingMiddleware{
			logger: logger,
			next:   next,
		}
	}
}

type serviceLoggingMiddleware struct {
	logger log.Logger
	next   Service
}

func (mw serviceLoggingMiddleware) GetUserinfo(ctx context.Context, id int64) (v Userinfo, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "GetUserinfo",
			"id", id, "result", fmt.Sprintf("%v", v), "error", err,
			"took", time.Since(begin),
		)
	}(time.Now())
	return mw.next.GetUserinfo(ctx, id)
}

// ServiceInstrumentingMiddleware returns a service middleware that instruments
// the number of integers summed and characters concatenated over the lifetime of
// the service.
func ServiceInstrumentingMiddleware(
	getUserinfoRequests metrics.Counter) Middleware {
	return func(next Service) Service {
		return serviceInstrumentingMiddleware{
			getUserinfoRequests: getUserinfoRequests,
			next:                next,
		}
	}
}

type serviceInstrumentingMiddleware struct {
	getUserinfoRequests metrics.Counter
	next                Service
}

func (mw serviceInstrumentingMiddleware) GetUserinfo(ctx context.Context, id int64) (Userinfo, error) {
	v, err := mw.next.GetUserinfo(ctx, id)
	mw.getUserinfoRequests.Add(1)
	return v, err
}
