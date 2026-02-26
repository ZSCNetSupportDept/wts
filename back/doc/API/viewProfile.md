# 查看个人信息 API

- **路径**: `/api/v3/view_profile`
- **方法**: `GET`
- **功能**: 获取用户的个人信息。

## 描述

此接口用于查询用户的详细个人资料。普通用户可以查看自己的信息，而管理员可以查看任何用户的信息。

## 认证

- **需要 JWT**: 是
- **权限要求**:
  - 普通用户 (`user` 或更高权限) 可以查看自己的信息。
  - 管理员 (`admin` 权限) 可以查看任何用户的信息。

## 请求

### 请求头

| Header          | 类型   | 描述                      |
| --------------- | ------ | ------------------------- |
| `Authorization` | string | `Bearer <your_jwt_token>` |

### 查询参数

| 参数 | 类型   | 描述                                                         | 是否必须 |
| ---- | ------ | ------------------------------------------------------------ | -------- |
| `who`  | string | 要查询用户的微信 OpenID。如果留空，则默认为当前登录用户的 OpenID。 | 否       |

**请求示例 (查看自己信息)**:

```
GET /api/v3/view_profile
```

**请求示例 (管理员查看他人信息)**:

```
GET /api/v3/view_profile?who=hajimihajimihajimi
```

## 响应

### 成功响应 (200 OK)

响应体包含一个 `profile` 对象，其中有用户的详细信息。

| 字段      | 类型    | 描述                                       |
| --------- | ------- | ------------------------------------------ |
| `success` | boolean | `true` 表示操作成功                        |
| `msg`     | string  | 成功的提示信息                             |
| `profile` | object  | 用户信息对象                               |
| `profile.sid`     | string  | 学号                                       |
| `profile.name`    | string  | 姓名                                       |
| `profile.block`   | string  | 宿舍楼                                     |
| `profile.access`  | string  | 用户权限等级 (例如: "user", "admin")       |
| `profile.room`    | string  | 房间号                                     |
| `profile.phone`   | string  | 手机号码                                   |
| `profile.isp`     | string  | 宽带运营商                                 |
| `profile.account` | string  | 宽带账号                                   |
| `profile.wx`      | string  | 微信 OpenID                                |

**响应示例**:

```json
{
  "success": true,
  "msg": "user profile",
  "profile": {
    "sid": "20230001001",
    "name": "张三",
    "block": "XH",
    "access": "user",
    "room": "1501",
    "phone": "13800138000",
    "isp": "mobile",
    "account": "12345678901",
    "wx": "hajimihajimihajimi"
  }
}
```

### 失败响应

#### 400 Bad Request (业务逻辑错误)

| `msg` 内容     | `errType` | 描述           |
| -------------- | --------- | -------------- |
| "无法找到该微信账户所请求的用户" | `logic`   | 目标用户不存在 |

#### 403 Forbidden (权限错误)

- 当非管理员用户尝试查看其他用户的信息时返回。
- 当非活跃用户（如 `unregistered`）尝试调用此接口时返回。

```json
{
  "success": false,
  "msg": "only admins can view other users' profiles",
  "errType": "auth"
}
```

```json
{
  "success": false,
  "msg": "only active users can access this API",
  "errType": "auth"
}
```

#### 500 Internal Server Error (服务器内部错误)

当发生未预料到的数据库错误或其他内部错误时返回。

```json
{
  "success": false,
  "msg": "system met a uncaught error,please view logs.",
  "errType": "internal"
}
```

## 注意事项

- 如果不提供 `who` 查询参数，API 将自动查询当前通过 JWT 认证的用户的信息。
- 只有管理员权限的用户才能使用 `who` 参数查询其他用户的信息。