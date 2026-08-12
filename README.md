# go-web-framework

[![Go Version](https://img.shields.io/badge/Go-1.25.4-blue.svg)](https://go.dev/)

基于 Go + Echo v5 的 Web 服务快速开发模板，内置完整可用的用户认证体系（JWT Access Token + Refresh Token 轮换），开箱即用，适合快速搭建后端 API 服务。

## 特性

- 基于 Echo v5 的 HTTP 服务，支持优雅停机
- 完整的用户认证方案：
  - JWT Access Token（`Authorization: Bearer <token>`）
  - Refresh Token 轮换机制：存入 MySQL（HMAC-SHA256 加盐哈希）、通过 HttpOnly Cookie 承载
  - 密码使用 bcrypt 加盐哈希存储
  - 修改密码后自动撤销该用户全部会话
- 统一 JSON 响应格式与业务错误码
- 结构化 JSON 日志（`log/slog`），慢请求自动告警
- 内置中间件：CORS、请求日志、JWT 鉴权、登录限流、Recover
- 健康检查接口（存活 / 就绪）
- 参数校验（go-playground/validator）、分页解析等开箱即用工具

## 技术栈

| 分类     | 技术                                        |
| -------- | ------------------------------------------- |
| 语言     | Go 1.25+                                    |
| Web 框架 | [Echo v5](https://github.com/labstack/echo) |
| 数据库   | MySQL（go-sql-driver/mysql + jmoiron/sqlx） |
| 认证     | golang-jwt/jwt/v5 + golang.org/x/crypto/bcrypt |
| 校验     | go-playground/validator/v10                 |
| 配置     | YAML（go.yaml.in/yaml/v4）                  |
| 日志     | 标准库 `log/slog`（JSON 格式）              |

## 目录结构

```
.
├── cmd/
│   └── api/
│       └── main.go            # 程序入口：加载配置、组装依赖、注册路由
├── internal/
│   ├── config/                # 配置加载与默认值填充
│   ├── server/                # Echo 实例、全局中间件、参数校验器
│   ├── domain/                # 业务领域层
│   │   ├── health/            # 健康检查
│   │   └── user/              # 用户域：实体、服务、处理器、错误定义
│   ├── infra/
│   │   └── repo/              # 数据访问层（MySQL）
│   ├── middleware/            # 自定义中间件（JWT 鉴权、请求日志）
│   ├── reqctx/                # 请求上下文（当前登录用户注入/读取）
│   ├── request/               # 请求参数解析（分页）
│   └── response/              # 统一响应封装
├── pkg/
│   ├── mysql/                 # MySQL 连接与 sqlx 封装
│   └── utils/                 # 哈希、随机字符串等工具
├── migration/                 # 数据库初始化 SQL
├── config.yaml                # 配置文件示例
├── Makefile                   # 多平台交叉编译
└── go.mod                     # module: myframework
```

## 快速开始

### 1. 初始化数据库

```sql
CREATE DATABASE web_framework CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

```bash
mysql -uroot -p web_framework < migration/001_user.sql
mysql -uroot -p web_framework < migration/002_refresh_token.sql
```

### 2. 修改配置

编辑 `config.yaml`，至少需要修改：

- `mysql_dsn`：数据库连接串（本地开发如 `root:123456@tcp(127.0.0.1:3306)/web_framework`）
- `jwt_secret`：JWT 签名密钥，**必须替换为随机值**
- `token_salt`：Refresh Token 哈希盐，**必须替换为随机值**
- `cookie_secure`：本地 HTTP 调试时设为 `false`，生产 HTTPS 环境保持 `true`

### 3. 启动服务

```bash
# 直接运行
go run ./cmd/api

# 指定配置文件
go run ./cmd/api -config ./config.yaml
```

服务默认监听 `:8080`，启动后可验证：

```bash
curl http://127.0.0.1:8080/health   # OK
curl http://127.0.0.1:8080/ready    # OK（会探测数据库连通性）
```

## 配置说明

| 配置项        | 默认值    | 说明                                       |
| ------------- | --------- | ------------------------------------------ |
| `server_addr` | `:8080`   | 服务监听地址                               |
| `log_level`   | `info`    | 日志级别，设为 `debug` 输出调试日志        |
| `allow_origins` | `[]`    | CORS 允许的跨域来源                        |
| `mysql_dsn`   | 必填      | MySQL 连接串，未配置将启动失败             |
| `jwt_secret`  | 必填      | JWT 签名密钥，未配置将启动失败             |
| `token_salt`  | 无        | Refresh Token 哈希盐                       |
| `access_ttl`  | `900`     | Access Token 有效期（秒）                  |
| `refresh_ttl` | `604800`  | Refresh Token 有效期（秒，默认 7 天）      |
| `cookie_secure` | `true`  | refresh_token Cookie 是否仅限 HTTPS 传输   |

## 接口列表

| 方法 | 路径               | 鉴权 | 说明                                       |
| ---- | ------------------ | ---- | ------------------------------------------ |
| GET  | `/health`          | 无   | 存活探针                                   |
| GET  | `/ready`           | 无   | 就绪探针（探测数据库连通性）               |
| POST | `/api/login`       | 无*  | 登录，返回 access_token，写入 refresh_token Cookie |
| POST | `/api/logout`      | 无   | 登出，清除 refresh_token                   |
| POST | `/api/refresh`     | 无   | 刷新令牌（轮换 refresh_token）             |
| GET  | `/api/user`        | Bearer | 获取当前登录用户信息                     |
| PUT  | `/api/user/password` | Bearer | 修改密码，成功后撤销所有会话             |

\* `/api/login` 内置内存限流（按 IP，每 10 个请求/滑动窗口），防止暴力破解。

### 认证流程

1. `POST /api/login` 携带 `{ "username": "...", "password": "..." }`
2. 校验通过后返回 Access Token：

```json
{
  "code": 0,
  "data": {
    "access_token": "<jwt>",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

同时通过 `Set-Cookie` 写入 `refresh_token`（HttpOnly + SameSite=Strict，按 `cookie_secure` 决定是否 Secure）。

3. 请求受保护接口时携带 `Authorization: Bearer <access_token>`
4. Access Token 过期后，调用 `POST /api/refresh`（Cookie 自动携带）轮换令牌，旧 refresh token 立即作废。

## 响应格式

统一响应包裹：

```json
// 成功
{ "code": 0, "data": { ... } }

// 业务错误
{ "code": 401, "message": "请先登录" }

// 列表
{ "code": 0, "data": { "page": 1, "page_size": 10, "total": 100, "list": [ ... ] } }
```

> 注意：业务错误码位于响应体 `code` 字段（`0` 表示成功），HTTP 状态码始终为 `200`；健康检查等非 JSON 接口除外。

## 构建

`Makefile` 支持多平台交叉编译，产物输出到 `build/` 目录：

```bash
make          # 等价于 make all：clean + 构建全部平台
make server   # 构建 linux-amd64、darwin-amd64、darwin-arm64

make server-linux-amd64     # 单独构建某一平台
make server-darwin-arm64
```

产物命名：`build/myframework-server-<os>-<arch>`（静态编译，`CGO_ENABLED=0`）。

## 新增业务模块参考

按 `internal/domain/user` 的既有模式扩展：

1. `internal/domain/<module>/`：定义实体（`entity.go`）、领域服务（`service.go`）、HTTP 处理器（`handler.go`）、错误定义
2. `internal/infra/repo/`：新增数据访问实现，实现 `domain` 中定义的 repo 接口
3. `cmd/api/main.go`：组装依赖并在 `register` 中注册路由

常用工具速查：

- 响应：`response.JsonData(c, data)` / `response.JsonError(c, code, msg)` / `response.JsonList(c, list, page, size, total)`
- 分页：`request.ParsePagination(c)`（page 默认 1，page_size 默认 10、上限 100）
- 当前用户：`reqctx.GetUser(c)`（由鉴权中间件注入）
- 校验：定义 struct 时加 `validate:"required"` 等 tag，处理器中调用 `c.Validate(&req)`
- 哈希/随机数：`utils.HMACSHA256Hex(key, msg)` / `utils.SHA256Hex` / `utils.RandomString(n)`
