# S3存储配置指南

本项目已支持S3存储系统。您可以选择使用本地文件系统或S3作为存储后端。

## 功能特性

- ✅ 支持本地文件系统存储
- ✅ 支持AWS S3存储
- ✅ 支持S3兼容存储（如MinIO、阿里云OSS等）
- ✅ 统一的存储接口，无缝切换
- ✅ 保持原有的安全访问控制机制

## 配置说明

### 1. 使用本地文件系统（默认）

在 `config.yaml` 中配置：

```yaml
storage:
  type: "local"
  localRoot: "./storage"
  dataRoot: "./data"
```

### 2. 使用AWS S3

在 `config.yaml` 中配置：

```yaml
storage:
  type: "s3"
  dataRoot: "./data"  # 下载文件的本地保存目录
  s3:
    bucket: "your-bucket-name"
    region: "us-east-1"
    prefix: "zfs/"  # 可选，所有文件的前缀
    # 认证方式1: 直接配置（不推荐在生产环境）
    accessKeyId: "YOUR_ACCESS_KEY_ID"
    secretAccessKey: "YOUR_SECRET_ACCESS_KEY"
    # 认证方式2: 留空使用环境变量或IAM角色（推荐）
    # accessKeyId: ""
    # secretAccessKey: ""
```

### 3. 使用S3兼容存储（如MinIO）

在 `config.yaml` 中配置：

```yaml
storage:
  type: "s3"
  dataRoot: "./data"
  s3:
    bucket: "your-bucket-name"
    region: "us-east-1"
    prefix: "zfs/"
    endpoint: "http://localhost:9000"  # MinIO endpoint
    forcePathStyle: true  # MinIO需要路径风格访问
    accessKeyId: "minioadmin"
    secretAccessKey: "minioadmin"
```

## 认证方式

### 方式1: 使用环境变量（推荐）

设置环境变量：

```bash
export AWS_ACCESS_KEY_ID="your-access-key-id"
export AWS_SECRET_ACCESS_KEY="your-secret-access-key"
export AWS_REGION="us-east-1"
```

然后在配置文件中留空认证信息：

```yaml
storage:
  type: "s3"
  s3:
    bucket: "your-bucket-name"
    region: "us-east-1"
    accessKeyId: ""
    secretAccessKey: ""
```

### 方式2: 使用AWS配置文件

配置 `~/.aws/credentials`:

```ini
[default]
aws_access_key_id = your-access-key-id
aws_secret_access_key = your-secret-access-key
```

配置 `~/.aws/config`:

```ini
[default]
region = us-east-1
```

### 方式3: 使用IAM角色（仅适用于EC2/ECS等）

如果您的应用运行在AWS环境中，可以使用IAM角色自动获取凭证，无需配置任何认证信息。

## S3 Bucket准备

### 创建S3 Bucket

```bash
# 使用AWS CLI创建bucket
aws s3 mb s3://your-bucket-name --region us-east-1
```

### 设置Bucket权限

确保您的IAM用户或角色具有以下权限：

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket",
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject"
      ],
      "Resource": [
        "arn:aws:s3:::your-bucket-name",
        "arn:aws:s3:::your-bucket-name/*"
      ]
    }
  ]
}
```

## 使用MinIO进行本地测试

### 1. 启动MinIO服务器

```bash
# 使用Docker运行MinIO
docker run -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"
```

### 2. 创建Bucket

访问 http://localhost:9001 使用浏览器创建bucket，或使用命令行：

```bash
# 安装mc客户端
wget https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x mc

# 配置mc
./mc alias set myminio http://localhost:9000 minioadmin minioadmin

# 创建bucket
./mc mb myminio/zfs-test
```

### 3. 配置项目使用MinIO

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

## 运行项目

```bash
cd src
go build
./ZFS
```

## 迁移现有数据到S3

如果您已有本地存储的文件，可以使用AWS CLI同步到S3：

```bash
# 同步本地storage目录到S3
aws s3 sync ./storage s3://your-bucket-name/zfs/

# 如果使用MinIO
./mc mirror ./storage myminio/zfs-test/zfs/
```

## 故障排查

### 1. 连接失败

- 检查网络连接和endpoint配置
- 确认AWS凭证配置正确
- 检查防火墙规则

### 2. 权限错误

- 确认IAM用户/角色具有必要的S3权限
- 检查Bucket策略配置

### 3. 找不到Bucket

- 确认Bucket名称正确
- 确认Bucket在指定的region中
- 如果使用MinIO，确认forcePathStyle设置为true

## 性能优化建议

1. **选择合适的Region**: 选择离您最近的AWS区域以减少延迟
2. **使用S3 Transfer Acceleration**: 对于跨地域传输，可以启用S3传输加速
3. **调整缓冲区大小**: 可以在代码中调整文件传输的缓冲区大小
4. **使用VPC Endpoint**: 如果运行在AWS内部，使用VPC Endpoint可以避免公网流量费用

## 安全最佳实践

1. ✅ 使用IAM角色而不是访问密钥
2. ✅ 启用S3 Bucket加密
3. ✅ 启用S3版本控制以防误删
4. ✅ 配置Bucket访问日志
5. ✅ 使用最小权限原则
6. ✅ 定期轮换访问密钥
7. ✅ 永远不要在代码中硬编码凭证

## 常见问题

**Q: 可以同时使用本地存储和S3吗？**

A: 目前每个节点只能配置一种存储类型，但不同节点可以使用不同的存储后端。

**Q: 切换存储类型后需要迁移数据吗？**

A: 是的，需要手动迁移现有数据到新的存储后端。

**Q: S3存储会增加延迟吗？**

A: 会有一定延迟，具体取决于网络条件和S3区域。建议选择就近的区域。

**Q: 支持其他云存储服务吗？**

A: 只要支持S3 API协议的服务都可以使用，如阿里云OSS、腾讯云COS、华为云OBS等（需要配置endpoint和credentials）。
