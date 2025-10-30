# S3存储支持 - 变更日志

## 概述

本次更新为ZFS分布式文件共享系统添加了完整的S3存储支持，允许项目使用AWS S3或任何S3兼容存储（如MinIO、阿里云OSS等）作为存储后端。

## 新增功能

### 1. 存储抽象层 (`src/storage/`)

创建了统一的存储接口，使得项目可以轻松切换不同的存储后端：

- **`storage.go`**: 定义了Storage接口，包含以下方法：
  - `ListDirectory()`: 列出目录内容
  - `DownloadFile()`: 下载文件
  - `UploadFile()`: 上传文件
  - `DeleteFile()`: 删除文件
  - `IsPathAllowed()`: 路径权限检查
  - `GetRoot()`: 获取存储根路径

- **`local.go`**: 本地文件系统存储实现
  - 完全兼容原有功能
  - 保持原有的安全访问控制

- **`s3.go`**: S3存储实现
  - 支持AWS S3
  - 支持S3兼容存储（MinIO、阿里云OSS等）
  - 支持多种认证方式（访问密钥、环境变量、IAM角色）
  - 支持自定义endpoint
  - 支持路径风格访问（MinIO需要）

- **`factory.go`**: 存储工厂
  - 根据配置自动创建相应的存储实例

### 2. 配置系统增强

#### 更新的文件
- `src/config/config.go`: 添加了StorageConfig和S3Config结构体
- `src/config.yaml`: 添加了storage配置段

#### 新增配置示例
- `src/config.s3.example.yaml`: AWS S3配置示例
- `src/config.minio.example.yaml`: MinIO配置示例

#### 配置项说明

```yaml
storage:
  type: "local|s3"           # 存储类型
  localRoot: "./storage"     # 本地存储根目录
  dataRoot: "./data"         # 下载文件保存目录
  s3:
    bucket: "bucket-name"    # S3存储桶
    region: "us-east-1"      # AWS区域
    prefix: "zfs/"           # 对象key前缀
    accessKeyId: ""          # 访问密钥ID
    secretAccessKey: ""      # 访问密钥
    endpoint: ""             # 自定义endpoint
    forcePathStyle: false    # 路径风格访问
```

### 3. 核心代码重构

#### `src/cmd/rpc.go`
- 移除了硬编码的本地文件系统操作
- FileServer结构体添加storage字段
- StartServer()函数接收storage参数
- ListDirectory()方法使用storage接口
- DownloadFile()方法使用storage接口

#### `src/cmd/op.go`
- Manager结构体添加dataRoot字段
- NewManager()函数接收dataRoot参数
- get()方法使用配置的dataRoot

#### `src/cmd/main.go`
- 添加storage初始化逻辑
- 将storage实例传递给StartServer()
- 从配置读取dataRoot并传递给Manager

### 4. 依赖项更新

在 `go.mod` 中添加了AWS SDK依赖：
- `github.com/aws/aws-sdk-go-v2`
- `github.com/aws/aws-sdk-go-v2/config`
- `github.com/aws/aws-sdk-go-v2/credentials`
- `github.com/aws/aws-sdk-go-v2/service/s3`
- `github.com/aws/aws-sdk-go-v2/feature/s3/manager`

### 5. 文档

新增文档文件：
- `S3_SETUP_GUIDE.md`: S3存储详细配置指南
- `CHANGELOG_S3.md`: 本变更日志
- 更新 `README.md`: 添加S3支持说明

## 技术实现细节

### 存储抽象设计

采用接口驱动设计，定义统一的Storage接口：

```go
type Storage interface {
    ListDirectory(ctx context.Context, path string) ([]FileInfo, error)
    DownloadFile(ctx context.Context, path string) (io.ReadCloser, error)
    UploadFile(ctx context.Context, path string, reader io.Reader) error
    DeleteFile(ctx context.Context, path string) error
    IsPathAllowed(path string) (bool, error)
    GetRoot() string
}
```

### S3实现要点

1. **路径映射**: S3使用对象key而非文件路径，实现了路径到key的转换
2. **目录模拟**: 使用delimiter和prefix模拟目录结构
3. **流式传输**: 保持原有的流式文件传输机制
4. **安全控制**: 通过IsPathAllowed()保持访问控制

### 兼容性

- ✅ 完全向后兼容，默认使用本地存储
- ✅ 无需修改现有代码即可切换存储类型
- ✅ 保持原有的gRPC接口不变
- ✅ 维持原有的安全访问控制机制

## 使用方式

### 本地存储（默认）

```bash
cd src
go build
./ZFS
```

### S3存储

1. 配置 `config.yaml`:
```yaml
storage:
  type: "s3"
  s3:
    bucket: "your-bucket"
    region: "us-east-1"
```

2. 设置AWS凭证（可选）:
```bash
export AWS_ACCESS_KEY_ID="..."
export AWS_SECRET_ACCESS_KEY="..."
```

3. 运行:
```bash
./ZFS
```

### MinIO测试

1. 启动MinIO:
```bash
docker run -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"
```

2. 使用 `config.minio.example.yaml` 配置

3. 运行项目

## 性能考虑

- S3操作相比本地文件系统会有额外的网络延迟
- 建议选择地理位置接近的S3区域
- 可以考虑启用S3传输加速（需额外配置）
- 对于AWS内部部署，建议使用VPC Endpoint

## 安全建议

1. ✅ 优先使用IAM角色而非访问密钥
2. ✅ 启用S3 Bucket加密
3. ✅ 配置合理的Bucket策略
4. ✅ 启用访问日志
5. ✅ 定期审计权限
6. ✅ 不要在代码中硬编码凭证

## 后续计划

- [ ] 添加S3多部分上传支持（大文件优化）
- [ ] 添加文件缓存层
- [ ] 支持对象存储的CDN加速
- [ ] 添加存储使用统计
- [ ] 支持文件版本控制

## 测试

项目已通过以下测试：
- ✅ 编译测试：成功编译为33MB可执行文件
- ✅ 本地存储兼容性：保持原有功能
- ✅ S3配置解析：正确读取S3配置
- ✅ 代码静态分析：无明显错误

建议进行以下测试：
- [ ] MinIO集成测试
- [ ] AWS S3实际环境测试
- [ ] 多节点协同测试
- [ ] 大文件传输测试
- [ ] 并发访问测试

## 贡献者

本次S3存储支持由AI助手完成，遵循项目现有的代码风格和架构设计。

## 相关链接

- [AWS S3 文档](https://docs.aws.amazon.com/s3/)
- [MinIO 文档](https://min.io/docs/)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/)
