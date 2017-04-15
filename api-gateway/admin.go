package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/kit/log"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/fasthttp"
)

type Ret struct {
	Errcode int         `json:"errcode"`
	Errmsg  string      `json:"errmsg"`
	Data    interface{} `json:"data"`
}

//服务实例
type Service struct {
	Service  string   `json:"service"`   //服务名字
	Iport    string   `json:"iport"`     //ip:port信息
	Title    string   `json:"title"`     //备注
	PingUri  string   `json:"ping_uri"`  //后端服务检测uri
	PingHost string   `json:"ping_host"` //后端服务检测host
	Status   int      `json:"status"`    //后端服务返回状态码
	State    string   `json:"state"`     //后端服务状态
	Quit     chan int `json:"-"`         //退出定时器
}

//Admin route管理工具，维护apigateway rule->service 映射表
type Admin struct {
	proxy    *Proxy
	client   *Subclient
	cfg      *Config
	db       *sql.DB
	logger   log.Logger
	services map[string]*Service
}

func NewAdmin(logger log.Logger, cfg *Config, proxy *Proxy, client *Subclient) (*Admin, error) {
	log.With(logger, "module", "admin")

	db, err := sql.Open("mysql", cfg.DB.Host)
	if err != nil {
		logger.Log("failed open mysql", cfg.DB.Host)
	}

	admin := &Admin{
		logger:   logger,
		cfg:      cfg,
		proxy:    proxy,
		client:   client,
		db:       db,
		services: make(map[string]*Service),
	}

	err = admin.loadServices()
	err = admin.loadTable()

	return admin, err
}

func (self *Admin) ret(ctx *fasthttp.RequestCtx, errcode int, errmsg string, data interface{}) {
	ctx.SetContentType("application/json; charset=utf8")

	ret := Ret{
		Errcode: errcode,
		Errmsg:  errmsg,
		Data:    data,
	}

	res, err := json.Marshal(ret)
	if err != nil {
		self.logger.Log("admin.ret", err.Error())
	}
	fmt.Fprintf(ctx, "%s", string(res))
}

//handleApiRegister 注册一个服务实例
func (self *Admin) handleApiRegister(ctx *fasthttp.RequestCtx) {
	args := ctx.PostArgs()
	service := string(args.Peek("service"))
	host := string(args.Peek("host"))
	title := string(args.Peek("title"))
	ping_uri := string(args.Peek("ping_uri"))
	ping_host := string(args.Peek("ping_host"))

	if service == "" || host == "" {
		self.ret(ctx, -1, "参数错误", nil)
		return
	}

	//表rule唯一
	_, err := self.db.Exec("INSERT INTO service (service, iport, title, ping_uri, ping_host) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE title=?, ping_uri=?, ping_host=?",
		service, host, title, ping_uri, ping_host, title, ping_uri, ping_host,
	)
	if err != nil {
		self.ret(ctx, -1, "写入数据库失败", nil)
		return
	}

	key := fmt.Sprintf("%s-%s", service, host)
	key = strings.Replace(key, ".", "-", -1)
	key = strings.Replace(key, ":", "-", -1)
	if s, ok := self.services[key]; ok {
		s.Quit <- 1
	}

	self.services[key] = &Service{
		Service:  service,
		Iport:    host,
		Title:    title,
		PingUri:  ping_uri,
		PingHost: ping_host,
		Quit:     make(chan int),
	}
	go self.ping(key)

	key = fmt.Sprintf("%s/%s", service, host)
	self.client.Register(key, host)

	self.ret(ctx, 0, "", nil)
}

//handleApiDeRegister 删除一个服务实例
func (self *Admin) handleApiDeRegister(ctx *fasthttp.RequestCtx) {
	args := ctx.PostArgs()
	service := string(args.Peek("service"))
	host := string(args.Peek("host"))

	if service == "" || host == "" {
		self.ret(ctx, -1, "参数错误", nil)
		return
	}

	_, err := self.db.Exec("DELETE FROM service WHERE service=? AND iport=?", service, host)
	if err != nil {
		self.ret(ctx, -1, "操作数据库失败", nil)
		return
	}

	key := fmt.Sprintf("%s-%s", service, host)
	key = strings.Replace(key, ".", "-", -1)
	key = strings.Replace(key, ":", "-", -1)

	self.services[key].Quit <- 1
	delete(self.services, key)

	key = fmt.Sprintf("%s/%s", service, host)
	self.client.Deregister(key)

	self.ret(ctx, 0, "", nil)
}

//供给zabbix用
func (self *Admin) handleApiServicesStatus(ctx *fasthttp.RequestCtx) {
	rs := []Service{}
	for _, v := range self.services {
		v.State = "success"
		if v.Status < 200 || v.Status >= 400 {
			v.State = "fail"
		}
		rs = append(rs, *v)
	}
	self.ret(ctx, 0, "", rs)
}

//handleApiServices 获取当前正在运行的服务和对应的后端实例列表
func (self *Admin) handleApiServices(ctx *fasthttp.RequestCtx) {
	self.ret(ctx, 0, "", self.services)
}

//handleApiList 列出所有当前生效的api配置信息
func (self *Admin) handleApiList(ctx *fasthttp.RequestCtx) {
	routes := self.proxy.GetRoutes() //map[string]*Route
	rs := []Route{}
	for _, route := range routes {
		rs = append(rs, *route)
	}

	self.ret(ctx, 0, "", rs)
}

//handleApiTable 列出所有的api配置信息
func (self *Admin) handleApiTable(ctx *fasthttp.RequestCtx) {
	rs := []Route{}

	rows, err := self.db.Query("SELECT rule, service, cache, sign, auth, title, create_time, update_time FROM route")
	if err != nil {
		self.logger.Log("admin.handleApiTable", err.Error())
		self.ret(ctx, -1, err.Error(), nil)
		return
	}

	for rows.Next() {
		var rule string
		var service string
		var cache int
		var sign int
		var auth int
		var title string
		var ct int
		var ut int
		if err := rows.Scan(&rule, &service, &cache, &sign, &auth, &title, &ct, &ut); err != nil {
			self.logger.Log("admin.handleApiTable", err.Error())
			continue
		}

		rs = append(rs, Route{
			Rule:       rule,
			Service:    service,
			Cache:      cache,
			Sign:       sign,
			Auth:       auth,
			Title:      title,
			CreateTime: ct,
			UpdateTime: ut,
		})
	}
	rows.Close()

	self.ret(ctx, 0, "", rs)
}

//handleApiPut 新增一条配置信息，如果之前有则覆盖
func (self *Admin) handleApiPut(ctx *fasthttp.RequestCtx) {
	args := ctx.PostArgs()
	name := string(args.Peek("name"))
	rule := string(args.Peek("rule"))
	service := string(args.Peek("service"))
	cache := args.GetUintOrZero("cache")
	auth := args.GetUintOrZero("auth")
	sign := args.GetUintOrZero("sign")
	title := string(args.Peek("title"))

	if rule == "" {
		self.ret(ctx, -1, "不合法的参数：规则空", nil)
		return
	}

	now := int(time.Now().Unix())
	//表rule唯一
	_, err := self.db.Exec("INSERT INTO route (name, rule, service, cache, sign, auth, title, create_time, update_time) VALUES (?,?,?,?,?,?,?,?, ?) ON DUPLICATE KEY UPDATE service=?, cache=?, sign=?, auth=?, title=?, update_time=?",
		name, rule, service, cache, sign, auth, title, now, now, service, cache, sign, auth, title, now,
	)

	if err != nil {
		self.logger.Log("admin.handleApiPut", err.Error())
		self.ret(ctx, -1, err.Error(), nil)
	} else {
		//写入proxy
		self.proxy.PutRoutes(&Route{
			Name:       name,
			Rule:       rule,
			Service:    service,
			Cache:      cache,
			Sign:       sign,
			Auth:       auth,
			Title:      title,
			CreateTime: now,
			UpdateTime: now,
		})
		self.ret(ctx, 0, "", nil)
	}
}

//handleApiDel 删除一条配置信息
func (self *Admin) handleApiDel(ctx *fasthttp.RequestCtx) {
	args := ctx.PostArgs()
	rule := string(args.Peek("rule"))

	if rule == "" {
		self.ret(ctx, -1, "不合法的参数：规则空", nil)
		return
	}

	_, err := self.db.Exec("DELETE FROM route WHERE rule=?", rule)
	if err != nil {
		self.logger.Log("admin.handleApiDel", err.Error())
		self.ret(ctx, -1, err.Error(), nil)
	} else {
		//写入proxy
		self.proxy.DelRoutes(rule)
		self.ret(ctx, 0, "", nil)
	}
}

//handleApiReload 全量加载配置信息
func (self *Admin) handleApiReload(ctx *fasthttp.RequestCtx) {
	err := self.loadTable()
	if err != nil {
		self.ret(ctx, -1, err.Error(), nil)
	} else {
		self.ret(ctx, 0, "", nil)
	}
}

func (self *Admin) loadTable() error {
	rs := []*Route{}

	rows, err := self.db.Query("SELECT name, rule, service, cache, sign, auth, title, create_time, update_time FROM route")
	if err != nil {
		self.logger.Log("admin.handleApiReload", err.Error())
		return err
	}

	for rows.Next() {
		var name string
		var rule string
		var service string
		var cache int
		var sign int
		var auth int
		var title string
		var ct int
		var ut int
		if err := rows.Scan(&name, &rule, &service, &cache, &sign, &auth, &title, &ct, &ut); err != nil {
			self.logger.Log("admin.handleApiReload", err.Error())
			continue
		}

		rs = append(rs, &Route{
			Name:       name,
			Rule:       rule,
			Service:    service,
			Cache:      cache,
			Sign:       sign,
			Auth:       auth,
			Title:      title,
			CreateTime: ct,
			UpdateTime: ut,
		})
	}
	rows.Close()

	self.proxy.ReloadRoutes(rs)

	return nil
}

func (self *Admin) loadServices() error {
	rows, err := self.db.Query("SELECT service, iport, title, ping_uri, ping_host FROM service")
	if err != nil {
		self.logger.Log("admin.loadServices", err.Error())
		return err
	}

	for rows.Next() {
		var service string
		var iport string
		var title string
		var ping_uri string
		var ping_host string
		if err := rows.Scan(&service, &iport, &title, &ping_uri, &ping_host); err != nil {
			self.logger.Log("admin.loadServices", err.Error())
			continue
		}

		key := fmt.Sprintf("%s-%s", service, iport)
		key = strings.Replace(key, ".", "-", -1)
		key = strings.Replace(key, ":", "-", -1)

		self.services[key] = &Service{
			Service:  service,
			Iport:    iport,
			Title:    title,
			PingUri:  ping_uri,
			PingHost: ping_host,
			Quit:     make(chan int),
		}

		go self.ping(key)
	}
	rows.Close()

	return nil
}

//检测后端服务状态
func (self *Admin) ping(key string) {
	s := self.services[key]
	hc := &fasthttp.HostClient{}
	hc.Addr = s.Iport
	var url string
	if s.PingHost == "" {
		url = fmt.Sprintf("http://%s%s", s.Iport, s.PingUri)
	} else {
		url = fmt.Sprintf("http://%s%s", s.PingHost, s.PingUri)
	}

	t := time.NewTimer(5 * time.Second)
	for {
		select {
		case <-s.Quit:
			return
		case <-t.C:
			code, _, _ := hc.Get(nil, url)
			s.Status = code
			t.Reset(5 * time.Second)
		}
	}
}

//HandleFastHttp 监听服务
func (self *Admin) HandleFastHttp(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/api/register":
		self.handleApiRegister(ctx)
	case "/api/deregister":
		self.handleApiDeRegister(ctx)
	case "/api/services":
		self.handleApiServices(ctx)
	case "/api/services/status":
		self.handleApiServicesStatus(ctx)
	case "/api/list":
		self.handleApiList(ctx)
	case "/api/table":
		self.handleApiTable(ctx)
	case "/api/put":
		self.handleApiPut(ctx)
	case "/api/del":
		self.handleApiDel(ctx)
	case "/api/reload":
		self.handleApiReload(ctx)
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

func (self *Admin) Close() {
	if self.db != nil {
		self.db.Close()
	}
}
