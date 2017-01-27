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
	return &CometConfig{}
}

func NewDispatchConfig() *DispatchConfig {
	return &DispatchConfig{}
}

func NewStoreConfig() *StoreConfig {
	return &StoreConfig{}
}

func NewPushConfig() *PushConfig {
	return &PushConfig{}
}
