## api-gateway说明文档

### 1.简单介绍

随着业务发展，原来的单体式的应用架构显得越来越臃肿，越来越难以维护，业务互相依赖，极大的降低了业务运行效率和维护成本；服务化就是将原本复杂的单体应用架构按照业务模块拆分成不同的子服务，每个服务独立对外处理请求。但是每个服务都要独立维护一整套监控，授权，缓存，服务发现，负载均衡等等逻辑，每个服务对外都需要暴露自己的端口，增加了安全风险等；

api-gateway主要用于提供统一的网关服务，将不同业务请求分发到后端对应的服务实例，将响应信息返回给客户端，统一入口做统一的基础逻辑，隐藏内部服务，提高了安全性，目前支持主要功能：

```
 - 1.监控，统计请求相关数据，服务负载情况等等<br />
 - 2.日志，记录请求信息，成功错误情况，打印错误堆栈信息<br />
 - 3.授权，校验用户请求权限，ip黑名单等策略防止对后端服务ddos<br />
 - 4.缓存，对数据进行缓存，具体看策略<br />
 - 5.服务发现，针对后端服务实例<br />
 - 6.负载均衡，根据注册的后端服务实例进行不同策略负载均衡<br />
 - 7.请求转发，基础功能，将客户端请求转发到特定服务实例<br />
 - 8.管理控制台，请求对应服务映射，动态维护etcd注册信息，手动操作实例挂载<br />
 - 9.限流<br />
 - 10.失败重试<br />
 - 11.请求合并<br />
```

### 2.架构说明

架构图

<img src="https://git.cn.memebox.com/global/api-gateway/raw/c19f62da2662c5ad38b3c256ce794ac6bdf4dcf1/files/jiagou1.png" width="600" height="300"/>

流程图

<img src="https://git.cn.memebox.com/global/api-gateway/raw/master/files/jiagou2.png" width="600" height="300"/>

性能对比

ab压测下来，简单的echo 字符串出来，每秒处理请求能力是nginx作为网关时候的2-3倍

### 3.启动服务

启动api gateway:

```
nohup ./api-gateway --config=/usr/local/go/src/api-gateway/config.test.yml >access.log 2>&1 &
```

启动一个测试后端goods服务:

```
nohup ./sa --addr=127.0.0.1:19080 --service=goods --etcd.addr=http://127.0.0.1:2379 >19080.log 2>&1 &
```

启动一个测试后端orders服务:

```
nohup ./sa --addr=127.0.0.1:19081 --service=orders --etcd.addr=http://127.0.0.1:2379 >19081.log 2>&1 &
```

添加服务goods请求映射表

```
curl -d "name=goods&rule=127.0.0.1%3A18080%2Fgoods&service=goods&title=goods+service" "http://127.0.0.1:18082/api/put"
```

添加服务orders请求映射表

```
curl -d "name=orders&rule=127.0.0.1%3A18080%2Forders&service=orders&title=orders+service" "http://127.0.0.1:18082/api/put"
```

测试api gateway请求返回

```
curl "http://127.0.0.1:18080/goods/test"
```


### 4.管理接口

管理服务端口查看配置文件admin端口配置

- 添加一个实例到某个服务，立即生效

  Uri: /api/register<br />
  Method: post<br />
  Params: <br />

  - service(服务名字，需要同路由映射配置的service一样)
  - host(实例ip:port)

  比如：添加实例127.0.0.1:19081到服务orders

```
curl -d "service=orders&host=127.0.0.1:19081" "http://127.0.0.1:18082/api/register"
```

- 删除某个服务的一个实例，立即生效

  Uri: /api/deregister<br />
  Method: post<br />
  Params:<br />

  - service(服务名字，需要同路由映射配置的service一样)
  - host(实例ip:port)

  比如：删除服务orders的实例127.0.0.1:19081

```
curl -d "service=orders&host=127.0.0.1:19081" "http://127.0.0.1:18082/api/deregister"
```

- 获取当前正在运行的服务实例列表

  Uri: /api/services<br />
  Method: get<br />
  Params:<br />

  比如：获取获取所有服务实例列表

```
curl "http://127.0.0.1:18082/api/services"
```

- 列出所有当前生效的路由映射配置信息

  Uri: /api/list<br />
  Method: get<br />
  Params:<br />

  比如：列出所有当前生效的所有路由映射配置信息

```
curl "http://127.0.0.1:18082/api/list"
```

- 列出所有当前配置的路由映射配置信息表

  Uri: /api/table<br />
  Method: get<br />
  Params:<br />

  比如：列出所有当前配置的路由映射配置信息表

```
curl "http://127.0.0.1:18082/api/table"
```

- 重新加载当前配置的路由映射配置信息表到内存

  Uri: /api/reload<br />
  Method: get<br />
  Params:<br />

  比如：重新加载当前配置的路由映射配置信息表到内存

```
curl "http://127.0.0.1:18082/api/reload"
```

- 根据rule更新某条路由配置规则，之前rule存在则更新，否则新增，立即生效

  Uri: /api/put<br />
  Method: post<br />
  Params:<br />

  - name(规则描述名，暂无用)
  - rule(规则，不带http或者https前缀的url配置，支持正则，比如"127.0.0.1:8080/orders/(.\*?)"，请求"127.0.0.1:8080/orders/detail?id=1"将会命中)
  - service(服务名，表示这条规则会分发到哪个服务)
  - cache(缓存配置，秒数，0表示不缓存)
  - auth(是否验证登录态，0表示不验证登录态)
  - sign(是否需要验证签名，0表示不验签)
  - title(路由规则配置描述，备注信息)

  比如：更新配置rule [127.0.0.1:18080/orders/(.*?)]配置

```
curl -d "name=orders&rule=127.0.0.1%3A18080%2Forders%2F%28.%2A%3F%29&service=orders&title=orders+service+rule+config&cache=0&auth=0&sign=0" "http://127.0.0.1:18082/api/put"
```

- 根据rule删除某条路由配置规则，立即生效

  Uri: /api/del<br />
  Method: post<br />
  Params:<br />

  - rule(规则)

  比如：删除配置rule [127.0.0.1:18080/orders/(.*?)]配置

```
curl -d "rule=127.0.0.1%3A18080%2Forders%2F%28.%2A%3F%29" "http://127.0.0.1:18082/api/del"
```


### 5.简单的控制台

<img src="https://git.cn.memebox.com/global/api-gateway/raw/master/files/console.png" width="800" height="400"/>
