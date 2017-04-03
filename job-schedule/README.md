## job-schedule说明文档

### 1.简单介绍

实现秒级运行任务管理，通过配置参数允许同时运行多个脚本进程，更好的弥补crontab缺乏的秒级控制能力，适合需要实施运行的脚本任务

只需要更新配置文件就可以更新任务，不需要重启，动态维护


### 2.配置文件介绍

times: 脚本运行尝试最大次数，避免脚本错误，无限起进程运行脚本

jobs: 任务配置列表

单个任务配置参数:

description: 任务描述，备注，本身对于脚本执行没有影响

cmd: 脚本执行的命令环境，如/usr/local/bin/php, /bin/bash等

params: 脚本文件参数，一般只需要填写脚本文件，自动后台运行

max: 该任务最大同时运行多少个

----
如下：

```
---
###job-schedule config file
#try max times
times: 50
#jobs
jobs:
- description: "test job"
  cmd: "/usr/local/bin/php"
  params: "/Users/administrator/Documents/workspace/meme-product/application/cronds/Example.php"
  max: 1
- description: "test job 2"
  cmd: "/usr/local/bin/php"
  params: "/Users/administrator/Documents/workspace/meme-product/application/cronds/ExampleRabbitReceiver.php"
  max: 1
- description: "test job 3"
  cmd: "/usr/local/bin/php"
  params: "/Users/administrator/Documents/workspace/meme-product/application/cronds/ExampleRabbitSender.php"
  max: 1
```
