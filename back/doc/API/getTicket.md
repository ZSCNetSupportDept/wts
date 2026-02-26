# 获取用户工单列表 API

- **路径**: `/api/v3/get_ticket`
- **方法**: `GET`
- **功能**: 获取指定用户提交的所有工单列表。

## 描述

此接口用于查询某个用户创建的所有报修工单。普通用户只能查询自己的工单，而管理员可以查询任何用户的工单。

## 认证

- **需要 JWT**: 是
- **权限要求**:
  - 普通用户 (`user` 或更高权限) 可以查看自己的工单。
  - 管理员 (`admin` 权限) 可以查看任何用户的工单。

## 请求

### 请求头

| Header          | 类型   | 描述                      |
| --------------- | ------ | ------------------------- |
| `Authorization` | string | `Bearer <your_jwt_token>` |

### 查询参数

| 参数 | 类型   | 描述                                                         | 是否必须 |
| ---- | ------ | ------------------------------------------------------------ | -------- |
| `who`  | string | 要查询工单的用户的微信 OpenID。如果留空，则默认为当前登录用户的 OpenID。 | 否       |

**请求示例 (查看自己工单)**:

```
GET /api/v3/get_ticket
```

**请求示例 (管理员查看他人所有工单)**:

```
GET /api/v3/get_ticket?who=another_user_openid_456
```

## 响应

### 成功响应 (200 OK)

响应体包含一个 `tickets` 数组，其中每个元素都是一个工单的详细信息。

| 字段        | 类型    | 描述                                                         |
| ----------- | ------- | ------------------------------------------------------------ |
| `success`   | boolean | `true` 表示操作成功                                          |
| `msg`       | string  | 成功的提示信息                                               |
| `tickets`   | array   | 工单对象数组                                                 |
| `ticket.tid`          | string  | 工单 ID                                                      |
| `ticket.submitted_at` | string  | 提交时间                                                     |
| `ticket.occur_at`     | string  | 问题发生时间                                                 |
| `ticket.description`  | string  | 问题描述                                                     |
| `ticket.appointed_at` | string  | 预约上门时间                                                 |
| `ticket.notes`        | string  | 备注                                                         |
| `ticket.priority`     | string  | 优先级                                                       |
| `ticket.category`     | string  | 问题分类                                                     |
| `ticket.status`       | string  | 当前状态                                                     |
| `ticket.last_updated_at`| string  | 最后更新时间                                                 |
| `ticket.issuer`       | object  | 报修人信息对象 (结构同 `view_profile` API 的 `profile` 对象) |

**响应示例**:

```json
{
  "success": true,
  "msg": "query success",
  "tickets": [
    {
      "tid": "T20251206001",
      "submitted_at": "2025-12-06T10:00:00Z",
      "occur_at": "2025-12-06T09:30:00Z",
      "description": "宿舍 WIFI 突然无法连接，路由器灯正常。",
      "appointed_at": "2025-12-07T00:00:00Z",
      "notes": "明天下午都有空。",
      "priority": "medium",
      "category": "无法连接",
      "status": "scheduled",
      "last_updated_at": "2025-12-06T10:00:00Z",
      "issuer": {
        "sid": "20230001001",
        "name": "张三",
        "block": "A",
        "access": "user",
        "room": "101",
        "phone": "13800138000",
        "isp": "中国移动",
        "account": "123456789",
        "wx": "user_openid_123"
      }
    }
  ]
}
```

如果用户没有任何工单，`tickets` 数组将为空。

### 失败响应

#### 400 Bad Request (业务逻辑错误)

| `msg` 内容     | `errType` | 描述           |
| -------------- | --------- | -------------- |
| `no such user` | `logic`   | 目标用户不存在 |

#### 403 Forbidden (权限错误)

- 当非管理员用户尝试查看其他用户的工单时返回。
- 当非活跃用户（如 `unregistered`）尝试调用此接口时返回。

```json
{
  "success": false,
  "msg": "only admins can view other users' own tickets",
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

- 如果不提供 `who` 查询参数，API 将自动查询当前认证用户的所有工单。
- 返回的工单列表包含了每个工单的详细信息以及报修人的完整个人信息。
- 这是一个查询操作，不会修改任何数据。