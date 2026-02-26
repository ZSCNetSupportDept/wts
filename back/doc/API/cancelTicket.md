# 取消工单 API

- **路径**: `/api/v3/cancel_ticket`
- **方法**: `POST`
- **功能**: 用户取消自己提交的报修工单。

## 描述

此接口允许用户取消一个他们自己创建的、尚未完成的工单。取消操作会向该工单添加一条新的追踪记录，并将工单状态更新为 `canceled`。

## 认证

- **需要 JWT**: 是
- **权限要求**:
  - 任何已激活的用户 (`user` 或更高权限) 都可以取消自己的工单。
  - 管理员 (`admin` 或 `dev` 权限) 可以取消任何工单。

## 请求

### 请求头

| Header          | 类型   | 描述                      |
| --------------- | ------ | ------------------------- |
| `Authorization` | string | `Bearer <your_jwt_token>` |

### 查询参数

| 参数 | 类型   | 描述             | 是否必须 |
| ---- | ------ | ---------------- | -------- |
| `tid`  | string | 要取消的工单 ID。 | 是       |

**请求示例**:

```
POST /api/v3/cancel_ticket?tid=T20251206001
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
  "msg": "ticket canceled"
}
```

### 失败响应

#### 400 Bad Request (请求错误)

- `tid` 参数缺失或格式不正确。
- 无法获取工单信息（例如，工单不存在）。

```json
{
  "success": false,
  "msg": "missing required URL parameter: tid",
  "errType": 2
}
```

```json
{
  "success": false,
  "msg": "invalid ticket ID: ...",
  "errType": 2
}
```

#### 400 Bad Request (业务逻辑错误)

| `msg` 内容              | `errType` | 描述                                                         |
| ----------------------- | --------- | ------------------------------------------------------------ |
| `no such ticket`        | `logic`   | 提供的 `tid` 对应的工单不存在。                              |
| `new status is invalid` | `logic`   | 工单当前状态不允许被取消 (例如，已经完成或已经取消的工单)。 |

#### 403 Forbidden (权限错误)

- 当用户尝试取消不属于自己的工单时返回。
- 当非活跃用户（如 `unregistered`）尝试调用此接口时返回。

```json
{
  "success": false,
  "msg": "you can only cancel tickets of your own",
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

- 此操作是不可逆的。一旦工单被取消，通常不能再重新打开。
- 只有工单的创建者或管理员才能取消工单。
- 后台逻辑会检查工单的当前状态，确保只有在特定状态下的工单才能被取消。
- 取消操作实际上是调用了 `AppendTrace` 逻辑，添加了一条备注为“用户取消报修”并更新状态为 `canceled` 的记录。