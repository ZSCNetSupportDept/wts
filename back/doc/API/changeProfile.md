# 修改个人信息 API

- **路径**: `/api/v3/change_profile`
- **方法**: `POST`
- **功能**: 更新用户的个人信息。

## 描述

此接口允许用户修改他们的个人资料，如宿舍信息、联系电话、宽带运营商和账号。管理员有权修改任何用户的信息，而普通用户只能修改自己的信息。

## 认证

- **需要 JWT**: 是
- **权限要求**:
  - 普通用户 (`user` 或更高权限) 只能修改自己的信息。
  - 管理员 (`admin` 权限) 可以修改任何用户的信息。

## 请求

### 请求头

| Header          | 类型   | 描述                      |
| --------------- | ------ | ------------------------- |
| `Authorization` | string | `Bearer <your_jwt_token>` |
| `Content-Type`  | string | 必须是 `application/json` |

### 请求体 (JSON)

| 字段      | 类型     | 描述                                       | 校验规则     |
| --------- | -------- | ------------------------------------------ | ------------ |
| `who`     | string   | 要修改信息用户的微信 OpenID                | `required`   |
| `block`   | string   | 新的宿舍区         | `required`   |
| `room`    | string   | 新的房间号                                 | `required`   |
| `phone`   | string   | 新的手机号码                               | `required`   |
| `isp`     | string   | 新的宽带运营商| `required`   |
| `account` | string   | 新的宽带账号                               | `required`   |

**请求示例 (普通用户修改自己信息)**:

```json
{
  "who": "user_openid_123",
  "block": "15",
  "room": "202",
  "phone": "13900139000",
  "isp": "telecom",
  "account": "987654321"
}
```

**请求示例 (管理员修改他人信息)**:

```json
{
  "who": "another_user_openid_456",
  "block": "9",
  "room": "303",
  "phone": "13700137000",
  "isp": "mobile",
  "account": "555555555"
}
```

## 响应

### 成功响应 (200 OK)

| 字段    | 类型    | 描述                 |
| ------- | ------- | -------------------- |
| `success` | boolean | `true` 表示操作成功  |
| `msg`     | string  | 成功的提示信息     |

**响应示例**:

```json
{
  "success": true,
  "msg": "profile change success~"
}
```

### 失败响应

#### 400 Bad Request (请求错误)

请求体绑定失败、格式错误或未通过验证。

```json
{
  "success": false,
  "msg": "invalid request body: Key: 'ChangeUserProfileRequest.Phone' Error:Field validation for 'Phone' failed on the 'required' tag",
  "errType": 2
}
```

#### 400 Bad Request (业务逻辑错误)

| `msg` 内容                   | `errType` | 描述                     |
| ---------------------------- | --------- | ------------------------ |
| `no such user`               | `logic`   | 目标用户不存在         |
| `phone number has been used` | `logic`   | 该手机号已被其他用户注册 |

#### 403 Forbidden (权限错误)

- 当非管理员用户尝试修改其他用户的信息时返回。
- 当非活跃用户（如 `unregistered`）尝试调用此接口时返回。

```json
{
  "success": false,
  "msg": "only admins can change other users' profiles",
  "errType": 3
}
```

```json
{
  "success": false,
  "msg": "only active users can access this API",
  "errType": 3
}
```

#### 500 Internal Server Error (服务器内部错误)

当发生未预料到的数据库错误或其他内部错误时返回。

```json
{
  "success": false,
  "msg": "system met a uncaught error,please view logs.",
  "errType": 1
}
```

## 注意事项

- 普通用户调用时，请求体中的 `who` 字段必须是自己的微信 OpenID，该 OpenID 从 JWT 中解析得出。
- 管理员可以指定任意用户的 OpenID 到 `who` 字段来修改其信息。
- 所有字段都是必需的，即使只修改其中一项，也需要提供所有字段的当前值或新值。
- 手机号码在系统中必须是唯一的。