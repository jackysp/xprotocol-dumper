1. 在tcpdump目录下，`sudo ./dumper lo 33060`。
  命令的第二个参数是网卡名，第三个参数是x protocol端口。
2. 启动mysqlsh客户端，进行一些操作后，可以停止`dumper`程序。
3. 第二步可以得到两个`*.tcpdump`文件，分别是客户端到服务端和服务端到客户端的，可以通过IP和端口来区分。
4. 之后可以在protocol目录下，执行`./protocol -client x.tcpdump -server y.tcpdump`。
  命令的第三、第五个参数分别是两个tcpdump文件名。
5. 之后，便在tcpdump目录下得到两个txt文件，便是协议的内容了。

ps: mysql server 启动时要增加skip-ssl这个选项，否则会解析错误。具体的，
1. 如果是源码编译的mysqld，直接启动的时候增加选项`--skip-ssl`或者增加到指定的
defaults-file里;
2. 如果是直接安装的二进制包，如deb包，可以在`/etc/mysql/mysql.conf.d/mysqld.cnf`里加入`skip-ssl`后,
重启mysql server `service mysql restart`.
