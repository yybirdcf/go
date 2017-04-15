package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/valyala/fasthttp"
)

var host string
var client *fasthttp.HostClient

type Ret struct {
	Errcode int         `json:"errcode"`
	Errmsg  string      `json:"errmsg"`
	Data    interface{} `json:"data"`
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	host = *flag.String("host", "127.0.0.1:18082", "admin host:port")
	flag.Parse()

	client = &fasthttp.HostClient{
		Addr: host,
	}
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	t := &Template{
		templates: template.Must(template.ParseGlob("./theme/*.html")),
	}
	e.Renderer = t

	e.Static("assets", "./theme/assets")
	// Route => handler
	e.GET("/", Index)
	e.GET("/index.html", Index)
	e.GET("/service/status", ServiceStatus)
	e.GET("/route.html", Route)

	// Start server
	e.Logger.Fatal(e.Start(":8181"))
}

func ServiceStatus(c echo.Context) error {
	statusCode, body, err := client.Get(nil, fmt.Sprintf("http://\\%s/api/services", host))
	if err != nil {
		log.Printf("request /api/services error: %s\n", err)
	}

	ret := Ret{}
	if statusCode == fasthttp.StatusOK {
		err := json.Unmarshal(body, &ret)
		if err != nil {
			log.Printf("request /api/services json.Unmarshal error: %s", err)
		}
	}

	data := map[string]interface{}{
		"data": ret.Data,
	}

	return c.JSON(http.StatusOK, data)
}

func Index(c echo.Context) error {
	s := c.QueryParam("down_service")
	h := c.QueryParam("down_host")

	if s != "" && h != "" {
		args := &fasthttp.Args{}
		args.Add("service", s)
		args.Add("host", h)
		client.Post(nil, fmt.Sprintf("http://\\%s/api/deregister", host), args)
	}

	us := c.FormValue("up_service")
	uh := c.FormValue("up_host")
	upu := c.FormValue("up_ping_uri")
	uph := c.FormValue("up_ping_host")
	ut := c.FormValue("up_title")

	if us != "" && uh != "" {
		args := &fasthttp.Args{}
		args.Add("service", us)
		args.Add("host", uh)
		args.Add("ping_uri", upu)
		args.Add("ping_host", uph)
		args.Add("title", ut)
		client.Post(nil, fmt.Sprintf("http://\\%s/api/register", host), args)
	}

	statusCode, body, err := client.Get(nil, fmt.Sprintf("http://\\%s/api/services", host))
	if err != nil {
		log.Printf("request /api/services error: %s", err)
	}

	ret := Ret{}
	if statusCode == fasthttp.StatusOK {
		err := json.Unmarshal(body, &ret)
		if err != nil {
			log.Printf("request /api/services json.Unmarshal error: %s", err)
		}
	}

	data := map[string]interface{}{
		"data": ret.Data,
	}

	return c.Render(http.StatusOK, "index", data)
}

func Route(c echo.Context) error {
	s := c.QueryParam("down_service")
	r := c.QueryParam("down_rule")

	if s != "" && r != "" {
		args := &fasthttp.Args{}
		args.Add("service", s)
		args.Add("rule", r)
		client.Post(nil, fmt.Sprintf("http://\\%s/api/del", host), args)
	}

	us := c.FormValue("up_service")
	ur := c.FormValue("up_rule")
	ut := c.FormValue("up_title")

	if us != "" && ur != "" {
		args := &fasthttp.Args{}
		args.Add("service", us)
		args.Add("rule", ur)
		args.Add("title", ut)
		client.Post(nil, fmt.Sprintf("http://\\%s/api/put", host), args)
	}

	action := c.QueryParam("action")
	if action == "refresh" {
		client.Get(nil, fmt.Sprintf("http://\\%s/api/reload", host))
	}

	statusCode, body, err := client.Get(nil, fmt.Sprintf("http://\\%s/api/list", host))
	if err != nil {
		log.Printf("request /api/list error: %s", err)
	}

	ret := Ret{}
	if statusCode == fasthttp.StatusOK {
		err := json.Unmarshal(body, &ret)
		if err != nil {
			log.Printf("request /api/list json.Unmarshal error: %s", err)
		}
	}

	data := map[string]interface{}{
		"data": ret.Data,
	}

	return c.Render(http.StatusOK, "route", data)
}
