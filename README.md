# go-web-framework

基于 Go + Echo v5 的 Web 服务快速开发模板，使用 ent ORM 管理数据模型，内置完整可用的用户认证体系（JWT Access Token + Refresh Token 轮换），开箱即用，适合快速搭建后端 API 服务。

## 特性

- 基于 Echo v5 的 HTTP 服务，支持优雅停机
- 数据模型由 ent（entgo.io/ent）定义与生成，类型安全、查询简洁
- 高性能 JSON 序列化（bytedance/sonic）
- 完整的用户认证方案：
  - JWT Access Token（`Authorization: Bearer <token>`）
  - Refresh Token 轮换机制：存入数据库（HMAC-SHA256 加盐哈希）、通过 HttpOnly Cookie 承载
  - 密码使用 bcrypt 加盐哈希存储
  - 修改密码后自动撤销该用户全部会话
- 统一 JSON 响应格式与业务错误码
- 结构化 JSON 日志（`log/slog`），`ENV=development` 输出 Debug 级别
- 内置中间件：CORS、请求日志、Request ID、JWT 鉴权、登录限流、Recover
- 健康检查接口（存活 / 就绪）
- 参数校验（go-playground/validator）、分页解析等开箱即用工具

## 技术栈

| 分类     | 技术                                        |
| -------- | ------------------------------------------- |
| 语言     | Go 1.25+                                    |
| Web 框架 | [Echo v5](https://github.com/labstack/echo) |
| ORM      | entgo.io/ent v0.14 + go-sql-driver/mysql    |
| 认证     | golang-jwt/jwt/v5 + golang.org/x/crypto/bcrypt |
| JSON     | github.com/bytedance/sonic                  |
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
│   ├── server/                # Echo 服务封装（函数式选项）、sonic JSON、校验器
│   ├── domain/                # 业务领域层
│   │   ├── health/            # 健康检查
│   │   └── user/              # 用户域：入口组装、仓库、服务、处理器、错误定义
│   ├── middleware/            # 自定义中间件（JWT 鉴权）
│   ├── reqctx/                # 请求上下文（当前登录用户注入/读取）
│   ├── request/               # 请求参数解析（分页）
│   └── response/              # 统一响应封装
├── ent/                       # ent ORM：schema 定义与生成的代码（随仓库提交）
│   └── schema/                # user.go / token.go 数据模型定义
├── pkg/
│   └── utils/                 # 哈希、随机字符串等工具
├── config.yaml                # 配置文件示例
├── Makefile                   # 多平台交叉编译
└── go.mod                     # module: myframework
```

## 快速开始

### 1. 准备数据库

创建数据库（ent 建表要求连接串中带 `charset=utf8mb4&parseTime=True`）：

```sql
CREATE DATABASE blog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. 修改配置

编辑 `config.yaml`，至少需要修改：

- `mysql_dsn`：数据库连接串（ent 支持 `mysql://` 或 `user:pass@tcp(host:port)/db` 格式）
- `jwt_secret`：JWT 签名密钥，**必须替换为随机值**
- `token_salt`：Refresh Token 哈希盐，**必须替换为随机值**
- `cookie_secure`：本地 HTTP 调试时设为 `false`，生产 HTTPS 环境保持 `true`

### 3. 同步数据表结构

数据表结构由 `ent/schema/` 中的定义决定，修改 schema 后需重新生成 ORM 代码：

```bash
go generate ./ent
```

表结构同步可使用 ent 的 Atlas 迁移（`atlas migrate diff` / `atlas migrate apply`），或在代码中调用 `client.Schema.Create(ctx)`。

### 4. 启动服务

```bash
# 直接运行（默认读取 ./config.yaml）
go run ./cmd/api

# 指定配置文件
go run ./cmd/api -config ./config.yaml

# 以 Debug 级别输出日志
ENV=development go run ./cmd/api
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
| `allow_origins` | `[]`    | CORS 允许的跨域来源                        |
| `mysql_dsn`   | 必填      | MySQL 连接串，未配置将启动失败             |
| `jwt_secret`  | 必填      | JWT 签名密钥，未配置将启动失败             |
| `token_salt`  | 无        | Refresh Token 哈希盐                       |
| `access_ttl`  | `900`     | Access Token 有效期（秒）                  |
| `refresh_ttl` | `604800`  | Refresh Token 有效期（秒，默认 7 天）      |
| `cookie_secure` | `true`  | refresh_token Cookie 是否仅限 HTTPS 传输   |

> 日志级别不再由配置文件控制，设置环境变量 `ENV=development` 时输出 Debug 日志，否则为 Info。

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

\* `/api/login` 内置内存限流（按 IP），防止暴力破解。

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
make api      # 构建 linux-amd64、darwin-amd64、darwin-arm64

make api-linux-amd64     # 单独构建某一平台
make api-darwin-arm64
```

产物命名：`build/myframework-api-<os>-<arch>`（静态编译，`CGO_ENABLED=0`）。

## 新增业务模块参考

按 `internal/domain/user` 的既有模式扩展：

1. `ent/schema/`：新增实体定义（字段、边、表名注解），执行 `go generate ./ent` 重新生成代码
2. `internal/domain/<module>/`：定义数据仓库（`repository.go`，基于 ent 的 `*ent.Client`）、领域服务（`service.go`）、HTTP 处理器（`handler.go`）、错误定义
3. `cmd/api/main.go`：组装依赖并注册路由（通过 `server.New(...)` 的函数式选项配置服务，`s.Router()` 获取路由组）

常用工具速查：

- 响应：`response.JsonData(c, data)` / `response.JsonError(c, code, msg)` / `response.JsonList(c, list, page, size, total)`
- 分页：`request.ParsePagination(c)`（page 默认 1，page_size 默认 10、上限 100）
- 当前用户：`reqctx.GetUser(c)`（由鉴权中间件注入）
- 校验：定义 struct 时加 `validate:"required"` 等 tag，处理器中调用 `c.Validate(&req)`
- 哈希/随机数：`utils.HMACSHA256Hex(key, msg)` / `utils.SHA256Hex` / `utils.RandomString(n)`
