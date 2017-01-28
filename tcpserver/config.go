package tcpserver

type CometConfig struct {
	TcpHost  string
	NsqdHost string

	RedisHost string
	RedisPwd  string
	RedisDb   int
}

type DispatchConfig struct {
	WorkerId int64
	NsqdHost string

	RedisHost string
	RedisPwd  string
	RedisDb   int
}

type StoreConfig struct {
	NsqdHost string

	RedisHost string
	RedisPwd  string
	RedisDb   int

	DbHost    string
	DbUser    string
	DbPwd     string
	DbName    string
	DbCharset string
}

type PushConfig struct {
	NsqdHost string

	RedisHost string
	RedisPwd  string
	RedisDb   int
}

func NewCometConfig() *CometConfig {
	return &CometConfig{
		TcpHost:   ":12000",
		NsqdHost:  ":4150",
		RedisHost: "127.0.0.1:6379",
		RedisPwd:  "123456",
		RedisDb:   1,
	}
}

func NewDispatchConfig() *DispatchConfig {
	return &DispatchConfig{
		WorkerId:  1,
		NsqdHost:  ":4150",
		RedisHost: "127.0.0.1:6379",
		RedisPwd:  "123456",
		RedisDb:   1,
	}
}

func NewStoreConfig() *StoreConfig {
	return &StoreConfig{
		NsqdHost:  ":4150",
		RedisHost: "127.0.0.1:6379",
		RedisPwd:  "123456",
		RedisDb:   1,
		DbHost:    "127.0.0.1:3306",
		DbUser:    "root",
		DbPwd:     "1160616612",
		DbName:    "im",
		DbCharset: "utf8mb4",
	}
}

func NewPushConfig() *PushConfig {
	return &PushConfig{
		NsqdHost:  ":4150",
		RedisHost: "127.0.0.1:6379",
		RedisPwd:  "123456",
		RedisDb:   1,
	}
}
