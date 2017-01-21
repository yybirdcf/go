用户基础信息服务usersvc说明文档
==============

1.服务说明
---------------------

####1.1.网关服务####
a.主要负责负载均衡</br>
b.etcd服务实现后端service自动发现</br>
c.多个后端service统一对外服务，屏蔽service实际业务逻辑</br>
d.多协议支持，提供RPC协议和HTTP接口协议</br>
e.发生错误，自动重试，重试超时</br>
f.自动限流，单个服务qps限制在10000</br>
g.断路器设置，支持全开，半开，关闭模式，保护后端服务</br>
h.protobuf协议支持，扩展多语言

目录为 <b>services/modules/usersvc/gateway/main.go</b>

安装：在该目录下执行

```
go install .
```
运行：go ROOT bin目录下执行

前台运行：

```
 ./gateway --grpc.addr=:9080  --http.addr=:9081  --etcd.addr=http://localhost:2379  --retry.max=3  --retry.timeout=500ms
```
后台运行：

```
  ./gateway --grpc.addr=:9080  --http.addr=:9081  --etcd.addr=http://localhost:2379  --retry.max=3  --retry.timeout=500ms > /alidata/logs/gateway.log 2>&1 &
```

参数说明：

  --grpc.addr: 对外服务的rpc地址</br>
  --http.addr: 对外服务的http地址，访问后端的实际是grpc方式</br>
  --etcd.addr: etcd服务host列表，多个用英文逗号分隔</br>
  --retry.max: 错误重试次数</br>
  --retry.timeout: 错误重试超时</br>

####1.2.service服务####

a.实际的用户基础信息服务，提供用户信息增改查</br>
b.etcd自动注册服务，运行以后可以被gateway自动发现并访问</br>
c.提供RPC和HTTP协议访问数据</br>
d.提供debug，metrics接口访问相关信息</br>
e.支持多实例运行</br>

目录为 <b>services/modules/usersvc/cmd/svc/main.go</b>

安装：在该目录下执行

```
go install .
```

运行：go ROOT bin目录下执行

运行第一个实例

```
 ./svc --debug.addr=:8080  --etcd.addr=http://localhost:2379  --grpc.addr=:8081 --http.addr=:8082  --env=pro --etcd.name=10.25.52.170:8081  > /alidata/logs/svc1.log 2>&1 &
```

运行第二个实例

```
 ./svc --debug.addr=:8090  --etcd.addr=http://localhost:2379  --grpc.addr=:8091 --http.addr=:8092  --env=pro --etcd.name=10.25.52.170:8091  > /alidata/logs/svc2.log 2>&1 &
```

参数说明：

  --debug.addr: 调试接口host</br>
  --etcd.addr: etcd服务host列表，多个用英文逗号分隔</br>
  --grpc.addr: 对外服务的rpc地址, 会被etcd添加到节点，负载均衡调度</br>
  --http.addr: 对外服务的http地址, 会被etcd添加到节点，负载均衡调度</br>
  --env: 运行的环境，读取不同的配置文件，如：dev, test, pro</br>
  --etcd.name: gateway读取的node节点，目前是当前机器ip rpc端口，统一rpc访问后端


2.测试服务(基于开发环境)
---------------------

####2.1.php客户端通过gateway访问服务####

services/modules/usersvc/cmd/cli里面：

```
  php main.php 127.0.0.1:9080  //php main.php [gateway rpc 地址]
```

####2.2.go客户端通过gateway访问服务####

services/modules/usersvc/cmd/cli里面：

```
	go install .
```

运行：go ROOT bin目录下执行

```
  ./cli --grpc.addr=:9080 --method=getUserinfo 97
```

####2.3.http方式:####

```
curl -d '{"id":97}' http://127.0.0.1:9081/usersvc/GetUserinfo  //访问gateway
```

```
curl -d '{"id":97,"attrs":"{\"username\":\"97\"}"}' http://127.0.0.1:9081/usersvc/UpdateUserinfo //访问gateway
```

3.API说明(测试数据基于开发环境)
---------------------

```
./compile.sh  //services/modules/usersvc/pb下执行生成go和php版本客户端文件
```

导入包或者引入文件参考services/modules/usersvc/cmd/cli下面测试例子

####3.1.获取单个用户基础信息####

grpc版本:

```
	client.GetUserinfo(context.Background(), &pb.GetUserinfoRequest{Id: 97})
```

参数结构

```
message GetUserinfoRequest {
  int64 id = 1;//用户id
}
```

http版本:

```
post方式
curl -d '{"id":97}' http://127.0.0.1:9081/usersvc/GetUserinfo
```

参数结构

```
	请求body为字典json串
	{
		"id": 用户id
	}
```
