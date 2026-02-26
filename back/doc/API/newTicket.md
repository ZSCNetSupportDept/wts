# 创建新工单 API

- **路径**: `/api/v3/new_ticket`
- **方法**: `POST`
- **功能**: 用户提交新的报修工单。

## 描述

此接口允许已注册用户提交新的网络报修工单。用户需要描述问题、问题发生时间、选择问题分类等。管理员可以为其他用户创建工单，并指定工单的初始状态和优先级。

## 认证

- **需要 JWT**: 是
- **权限要求**:
  - 任何已激活的用户 (`user` 或更高权限) 都可以为自己创建工单。
  - 只有管理员 (`admin`) 可以为其他用户创建工单，或在创建时指定 `status` 和 `priority`。

## 请求

### 请求头

| Header          | 类型   | 描述                      |
| --------------- | ------ | ------------------------- |
| `Authorization` | string | `Bearer <your_jwt_token>` |
| `Content-Type`  | string | 必须是 `application/json` |

### 请求体 (JSON)

| 字段          | 类型     | 描述                                                                 | 校验规则     |
| ------------- | -------- | -------------------------------------------------------------------- | ------------ |
| `issuer_sid`   | string   | 报修人的学号                                                 | `required`   |
| `description` | string   | 问题描述                                                             | `required`   |
| `category`    | string   | 问题分类  | `required`   |
| `occur_at`    | string   | 问题发生时间 (RFC3339 格式, 例如: "2025-12-06T10:00:00Z")             | `required`   |
| `appointed_at`| string   | 预约上门维修日期 (RFC3339 日期格式, 例如: "2025-12-07T00:00:00Z")     | (可选)       |
| `notes`       | string   | 备注信息                                                             | (可选)       |
| `status`      | string   | **(仅管理员)** 工单状态) | (可选)       |
| `priority`    | string   | **(仅管理员)** 工单优先级           | (可选)       |

**请求示例 (普通用户)**:

```json
{
  "issuer_sid": "2025020202022",
  "description": "宿舍 WIFI 突然无法连接，路由器灯正常。",
  "category": "ip-or-device",
  "occur_at": "2025-12-06T09:30:00Z",
  "appointed_at": "2025-12-07T00:00:00Z",
  "notes": "明天下午都有空。"
}
```

**请求示例 (管理员)**:

```json
{
  "issuer_sid": "2025020202022",
  "description": "用户反映整个楼层网络波动。",
  "category": "网络掉线",
  "occur_at": "2025-12-06T08:00:00Z",
  "status": "fresh",
  "priority": "assigned"
}
```

## 响应

### 成功响应 (201 Created)

| 字段    | 类型    | 描述                 |
| ------- | ------- | -------------------- |
| `success` | boolean | `true` 表示操作成功  |
| `msg`     | string  | 成功的提示信息     |
| `tid`     | string  | 编号              |

**响应示例**:

```json
{
  "success": true,
  "msg": "new ticket created",
  "tid": 12345
}
```

### 失败响应

#### 400 Bad Request (请求错误)

请求体绑定失败、格式错误或未通过验证。

```json
{
  "success": false,
  "msg": "invalid request body: ...",
  "errType": "request"
}
```

#### 400 Bad Request (业务逻辑错误)

| `msg` 内容                               | `errType` | 描述                                       |
| ---------------------------------------- | --------- | ------------------------------------------ |
| `no such user`                           | `logic`   | `issuer_sid` 对应的用户不存在             |
| `appointment time is invalid`            | `logic`   | 预约时间 `appointed_at` 在当前时间之前     |
| `occur time is invalid`                  | `logic`   | 问题发生时间 `occur_at` 在当前时间之后     |
| `you have too many active tickets`       | `logic`   | 用户有过多（例如超过3个）未完成的工单      |

#### 403 Forbidden (权限错误)

- 当非管理员用户尝试为他人创建工单时。
- 当非管理员用户在创建工单时尝试设置 `status` 或 `priority` 字段时。
- 当非活跃用户（如 `unregistered`）尝试调用此接口时。

```json
{
  "success": false,
  "msg": "only admins can create tickets for other users",
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

- 普通用户调用时，`issuer_sid` 必须是自己的学号；
- `occur_at` 时间不能晚于当前提交时间。
- `appointed_at` 如果提供，必须是未来的日期。如果提供了 `appointed_at`，工单状态会自动设置为 `scheduled`（除非管理员手动指定了其他状态）。
- 一个用户不能有太多（当前限制为3个）处于活动状态（非 `canceled` 或 `completed`）的工单。
- `status` 和 `priority` 字段仅供管理员在创建时设置，普通用户提交的工单将使用系统默认值。