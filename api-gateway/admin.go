package main

import (
	"encoding/json"
	"fmt"
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

//Admin route管理工具，维护apigateway rule->service 映射表
type Admin struct {
	proxy  *Proxy
	client *Subclient
	cfg    *Config
	db     *sql.DB
	logger log.Logger
}

func NewAdmin(logger log.Logger, cfg *Config, proxy *Proxy, client *Subclient) (*Admin, error) {
	log.With(logger, "module", "admin")

	db, err := sql.Open("mysql", cfg.DB.Host)
	if err != nil {
		logger.Log("failed open mysql", cfg.DB.Host)
	}

	admin := &Admin{
		logger: logger,
		cfg:    cfg,
		proxy:  proxy,
		client: client,
		db:     db,
	}

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

	if service == "" || host == "" {
		self.ret(ctx, -1, "参数错误", nil)
		return
	}

	key := fmt.Sprintf("%s/%s", service, host)
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

	key := fmt.Sprintf("%s/%s", service, host)
	self.client.Deregister(key)

	self.ret(ctx, 0, "", nil)
}

//handleApiServices 获取当前正在运行的服务和对应的后端实例列表
func (self *Admin) handleApiServices(ctx *fasthttp.RequestCtx) {
	data, err := self.client.GetEntries()
	if err != nil {
		self.ret(ctx, -1, err.Error(), nil)
		return
	}

	self.ret(ctx, 0, "", data)
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

//HandleFastHttp 监听服务
func (self *Admin) HandleFastHttp(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/api/register":
		self.handleApiRegister(ctx)
	case "/api/deregister":
		self.handleApiDeRegister(ctx)
	case "/api/services":
		self.handleApiServices(ctx)
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
