# ZFS - 分布式文件共享系统

使用etcd作为注册中心，grpc作为通信方式，所有的节点支持全局共享文件。每个节点的下的storage目录用于授权给其他节点访问，代码中有绝对路径判断，杜绝了跨权限访问问题。从其他节点下载的文件将保存在data目录。

## 🎉 新特性：支持S3存储

项目现已支持Amazon S3和S3兼容存储（如MinIO、阿里云OSS等）作为存储后端！

### 主要特性

- ✅ **本地存储**: 使用本地文件系统（默认）
- ✅ **S3存储**: 支持AWS S3作为存储后端
- ✅ **S3兼容存储**: 支持MinIO、阿里云OSS、腾讯云COS等
- ✅ **统一接口**: 无缝切换存储后端，无需修改业务逻辑
- ✅ **安全访问**: 保持原有的安全访问控制机制
- ✅ **灵活认证**: 支持多种AWS认证方式（环境变量、配置文件、IAM角色等）

### 快速配置

#### 使用本地存储（默认）

```yaml
storage:
  type: "local"
  localRoot: "./storage"
  dataRoot: "./data"
```

#### 使用AWS S3

```yaml
storage:
  type: "s3"
  dataRoot: "./data"
  s3:
    bucket: "your-bucket-name"
    region: "us-east-1"
    prefix: "zfs/"
```

#### 使用MinIO（本地测试）

```yaml
storage:
  type: "s3"
  dataRoot: "./data"
  s3:
    bucket: "zfs-test"
    region: "us-east-1"
    endpoint: "http://localhost:9000"
    forcePathStyle: true
    accessKeyId: "minioadmin"
    secretAccessKey: "minioadmin"
```

📖 **详细配置说明请查看**: [S3存储配置指南](./S3_SETUP_GUIDE.md)

## 启动方法

### 方法1: 不使用docker

#### 1. 启动etcd

```shell
etcd --listen-client-urls http://localhost:2379 \
--advertise-client-urls http://localhost:2379 \
--listen-peer-urls http://localhost:2380 \
--initial-advertise-peer-urls http://localhost:2380
```

#### 2. 配置存储

进入src目录，根据需要选择存储类型并配置：

**使用本地存储（默认）**
```bash
# config.yaml 已配置为本地存储，无需修改
```

**使用S3存储**
```bash
# 复制S3配置示例
cp config.s3.example.yaml config.yaml
# 编辑config.yaml，填入你的S3配置
vi config.yaml
```

**使用MinIO测试**
```bash
# 先启动MinIO
docker run -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"

# 复制MinIO配置示例
cp config.minio.example.yaml config.yaml
```

#### 3. 构建并运行

```bash
cd src
go build
./ZFS
```

**使用show命令查看所有节点**

![image-20250321204857878](/example/image-20250321204857878.png)



**使用ls命令，查看节点授权目录下所有文件，d 表示目录类型，- 表示文件类型**

![image-20250321205310924](/example/image-20250321205310924.png)



**使用get 命令下载远程节点文件**

![image-20250321205524356](/example/image-20250321205524356.png)

## 技术架构

- **服务发现**: etcd
- **通信协议**: gRPC
- **存储后端**: 本地文件系统 / Amazon S3 / S3兼容存储
- **日志系统**: zap
- **配置管理**: YAML

## 配置文件说明

主配置文件位于 `src/config.yaml`，包含以下配置项：

- `node`: 节点配置（节点名称等）
- `storage`: 存储配置（类型、路径、S3配置等）
- `etcd`: etcd服务配置
- `log`: 日志配置

示例配置文件：
- `config.yaml`: 默认配置（本地存储）
- `config.s3.example.yaml`: S3存储配置示例
- `config.minio.example.yaml`: MinIO存储配置示例

## 依赖项

- Go 1.23.5+
- etcd 3.5+
- (可选) AWS S3 或 MinIO

## 许可证

本项目采用开源许可证，仓库达到5个星会继续维护和增加新特性。

## 贡献

欢迎提交Issue和Pull Request！