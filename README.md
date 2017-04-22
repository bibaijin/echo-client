# echo-client

一个使用了 [LAIN](https://github.com/laincloud/lain)
[service](https://laincloud.gitbooks.io/white-paper/usermanual/service.html)
的 LAIN 应用。

## 功能

它向 [echo-service](https://github.com/bibaijin/echo-service) 发起 tcp 连接。
每隔 5 秒，它向 `echo-service` 写入一次 "ping\n"；并期待收到 "ping\n"，否则会在日志里报错。
