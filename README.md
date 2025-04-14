# IP Location Service

一个基于 Go 语言的 IP 地理位置查询服务，支持 GeoIP 和 ip2region 两种数据库。

## 功能特点

- 支持 GeoIP 和 ip2region 两种数据库
- 提供 RESTful API 接口
- 支持查询指定 IP 或自动获取客户端 IP
- 返回详细的地理位置信息，包括国家、省份、城市等
- 支持 CORS 跨域请求
- 提供健康检查接口

## 安装

### 前提条件

- Go 1.21 或更高版本
- Make 工具

### 下载数据库

项目支持两种 IP 地理位置数据库：

1. GeoIP 数据库（需要 MaxMind 许可证密钥）
2. ip2region 数据库（开源免费）

#### 下载 GeoIP 数据库

```bash
# 使用默认许可证密钥
make download-geoip

# 或指定自己的许可证密钥
make download-geoip LICENSE_KEY=your_key
```

> **注意**：使用 GeoIP 数据库需要 MaxMind 许可证密钥。您可以在以下地址申请免费的许可证密钥：
> - https://www.maxmind.com/en/geolite2/signup
> 
> 申请步骤：
> 1. 访问上述链接并注册账号
> 2. 登录后进入 "Services" -> "My License Key"
> 3. 点击 "Generate new license key" 生成新的许可证密钥
> 4. 复制生成的密钥，用于下载数据库

#### 下载 ip2region 数据库

```bash
make download-ip2region
```

**注意**：由于网络原因，自动下载 ip2region 数据库可能会失败。如果遇到此问题，您可以手动从以下地址下载 ip2region.xdb 文件，并将其放置在 `data/ip2region/` 目录下：

- https://github.com/lionsoul2014/ip2region/releases/download/v2.7.0/ip2region.xdb

### 构建项目

```bash
# 构建当前平台版本
make build

# 构建 Linux 平台版本
make build-linux
```

### 运行打包后的文件

构建完成后，可执行文件将位于 `bin` 目录下。运行方式如下：

```bash
# 使用默认配置运行（GeoIP 数据库，端口 8080）
./bin/ip-location-service

# 指定端口运行
./bin/ip-location-service -port 8081

# 使用 ip2region 数据库运行
./bin/ip-location-service -db ip2region

# 查看帮助信息
./bin/ip-location-service -help
```

可用的命令行参数：
- `-port`: 指定服务运行的端口号（默认：8080）
- `-db`: 指定使用的数据库类型（可选值：geoip 或 ip2region，默认：geoip）
- `-help`: 显示帮助信息

## 运行

```bash
# 使用默认配置运行（GeoIP 数据库，端口 8080）
make run

# 指定端口运行
make run PORT=8081

# 使用 ip2region 数据库运行
make run DB=ip2region
```

## API 接口

### 查询 IP 地理位置

```
GET /api/ip-location?ip=1.2.3.4
```

参数：
- `ip`：可选，要查询的 IP 地址。如果不提供，将使用请求者的 IP 地址。

响应示例：

```json
{
  "ip": "1.2.3.4",
  "country": "中国",
  "country_code": "CN",
  "province": "北京",
  "city": "北京",
  "district": "海淀区",
  "isp": "联通",
  "latitude": 39.9042,
  "longitude": 116.4074,
  "timezone": "Asia/Shanghai",
  "continent_en": "Asia",
  "continent_zh": "亚洲",
  "country_en": "China",
  "country_zh": "中国",
  "region_en": "Beijing",
  "region_zh": "北京",
  "city_en": "Beijing",
  "city_zh": "北京",
  "accuracy_km": 5,
  "source": "GeoIP"
}
```

### 健康检查

```
GET /health
```

响应示例：

```json
{
  "status": "ok",
  "db_type": "geoip"
}
```

## 数据库更新

### 更新 GeoIP 数据库

```bash
# 使用默认许可证密钥
make update-geoip

# 或指定自己的许可证密钥
make update-geoip LICENSE_KEY=your_key
```

### 更新 ip2region 数据库

```bash
make update-ip2region
```

## 许可证

本项目使用 MIT 许可证。请注意，GeoIP 数据库的使用需要遵守 MaxMind 的许可条款。
