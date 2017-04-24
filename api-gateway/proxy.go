package main

import (
	"fmt"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/valyala/fasthttp"
)

//单条规则
type Route struct {
	Name       string         `json:"name"`        //标签名字，唯一标示规则
	Rule       string         `json:"rule"`        //规则, cn.memebox.com/good/(.*?)
	Service    string         `json:"service"`     //服务名, goods
	Cache      int            `json:"cache"`       //0表示不缓存，缓存unix时间戳，单位s
	Sign       int            `json:"sign"`        //0表示不判断签名，否则判断签名
	Auth       int            `json:"auth"`        //0表示不判断登录态，否则判断登录态
	Title      string         `json:"title"`       //描述
	CreateTime int            `json:"create_time"` //创建时间
	UpdateTime int            `json:"update_time"` //更新时间
	Compile    *regexp.Regexp `json:"-"`           //正则模版
}

//Proxy 代理内核
type Proxy struct {
	routes          map[string]*Route
	hostClients     map[string]*fasthttp.HostClient
	mutex           sync.Mutex
	sub             *Subscribe
	servicesCounter metrics.Counter
	c               uint64
}

func NewProxy(sub *Subscribe, servicesCounter metrics.Counter) *Proxy {
	p := &Proxy{
		routes:          make(map[string]*Route),
		hostClients:     make(map[string]*fasthttp.HostClient),
		sub:             sub,
		servicesCounter: servicesCounter,
		c:               0,
	}

	go p.hcClean()
	return p
}

//HandleProxy 代理主方法
func (self *Proxy) ServeHTTP(ctx *fasthttp.RequestCtx) {
	//获取请求url
	url := append(ctx.Host(), ctx.RequestURI()...)

	//遍历rule配置
	var destRoute *Route
	for _, route := range self.routes {
		if len(route.Compile.FindIndex(url)) > 0 {
			//找到一个匹配
			destRoute = route
			break
		}
	}

	if destRoute == nil {
		//返回404
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		fmt.Fprintf(ctx, "not found route path")
		return
	}

	//查找对应的后端服务实例
	entries := self.sub.GetEntries(destRoute.Service)
	length := len(entries)
	if len(entries) == 0 {
		//返回403
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		fmt.Fprintf(ctx, "not found service route")
		return
	}

	//暂时随机选一个实例
	// i := rand.Intn(length)
	//轮询
	i := atomic.AddUint64(&self.c, 1) - 1
	proxyClient := self.getHostClient(entries[i%uint64(length)])

	req := &ctx.Request
	resp := &ctx.Response
	prepareRequest(ctx, req)
	if err := proxyClient.Do(req, resp); err != nil {
		ctx.Logger().Printf("error when proxying the request: %s", err)
	}
	postProcessResponse(ctx, resp)

	//计数
	self.servicesCounter.With("name", destRoute.Name, "service", destRoute.Service, "code", fmt.Sprint(resp.StatusCode())).Add(1)
}

func (self *Proxy) getHostClient(host string) *fasthttp.HostClient {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	hc := self.hostClients[host]
	if hc == nil {
		hc = &fasthttp.HostClient{
			Addr:     host,
			MaxConns: 1024,
		}
	}

	self.hostClients[host] = hc
	return hc
}

func (self *Proxy) hcClean() {
	for {
		t := time.Now()
		self.mutex.Lock()
		for k, v := range self.hostClients {
			if t.Sub(v.LastUseTime()) > time.Minute {
				delete(self.hostClients, k)
			}
		}
		self.mutex.Unlock()
		time.Sleep(10 * time.Second)
	}
}

func prepareRequest(ctx *fasthttp.RequestCtx, req *fasthttp.Request) {
	req.Header.Add("Client-Real-IP", ctx.RemoteIP().String())
	// do not proxy "Connection" header.
	req.Header.Del("Connection")
	// req.Header.Set("Connection", "Keep-Alive")
	// strip other unneeded headers.

	// alter other request params before sending them to upstream host
}

func postProcessResponse(ctx *fasthttp.RequestCtx, resp *fasthttp.Response) {
	resp.Header.Del("Client-Real-IP")
	// do not proxy "Connection" header
	resp.Header.Del("Connection")

	// strip other unneeded headers

	// alter other response data if needed
}

//ReloadRoutes 加载路由规则信息
func (self *Proxy) ReloadRoutes(routes []*Route) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	//清空规则
	for rule, _ := range self.routes {
		delete(self.routes, rule)
	}

	//生成新的规则
	for _, route := range routes {
		route.Compile, _ = regexp.Compile(route.Rule)
		self.routes[route.Rule] = route
	}
}

//GetRoutes 获取路由列表
func (self *Proxy) GetRoutes() map[string]*Route {
	return self.routes
}

//DelRoutes 删除某个路由
func (self *Proxy) DelRoutes(rule string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	delete(self.routes, rule)
}

//PutRoutes 配置某个路由
func (self *Proxy) PutRoutes(route *Route) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	route.Compile, _ = regexp.Compile(route.Rule)
	self.routes[route.Rule] = route
}
