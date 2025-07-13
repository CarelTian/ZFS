
使用etcd作为注册中心，grpc作为通信方式，所有的节点支持全局共享文件。每个节点的下的storage目录用于授权给其他节点访问，代码中有绝对路径判断，杜绝了跨权限访问问题。从其他节点下载的文件将保存在data目录。目前本项目只实现了基础功能，安全性访问，最佳实践暂未考虑，仓库达到5个星会继续维护和增加新特性.

### 启动方法(不使用docker)

进入src目录，配置etcd的Endpoints等参数。

启动etcd

```shell
etcd --listen-client-urls http://localhost:2379 \
--advertise-client-urls http://localhost:2379 \
--listen-peer-urls http://localhost:2380 \
--initial-advertise-peer-urls http://localhost:2380
```

go build

执行  ZFS

**使用show命令查看所有节点**

![image-20250321204857878](/example/image-20250321204857878.png)



**使用ls命令，查看节点授权目录下所有文件，d 表示目录类型，- 表示文件类型**

![image-20250321205310924](/example/image-20250321205310924.png)



**使用get 命令下载远程节点文件**

![image-20250321205524356](/example/image-20250321205524356.png)




