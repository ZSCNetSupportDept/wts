# 添加维修记录 (New Repair Trace) API

- **路径**: `/api/v3/new_repair_trace`
- **方法**: `POST`
- **功能**: 网维人员为工单添加处理记录、更新状态或修改其他属性。

## 描述

此接口是网维人员的核心工具，用于记录对报修工单的每一次操作。无论是状态变更（如“改日修”、“已完成”）、优先级调整，还是添加备注，都通过此接口完成。每一次调用都会在工单下生成一条新的追踪记录 (trace)。

## 认证

- **需要 JWT**: 是
- **权限要求**:
  - 必须是网维人员 (`operator` 或更高权限) 才能调用此 API。
  - 只有管理员 (`admin`) 才能修改工单的 `priority` (优先级)。

## 请求

### 请求头

| Header          | 类型   | 描述                      |
| --------------- | ------ | ------------------------- |
| `Authorization` | string | `Bearer <your_jwt_token>` |
| `Content-Type`  | string | 必须是 `application/json` |

### 请求体 (JSON)

| 字段             | 类型     | 描述                                                                 | 校验规则     |
| ---------------- | -------- | -------------------------------------------------------------------- | ------------ |
| `tid`            | integer  | 要更新的工单 ID                                                      | `required`   |
| `remark`         | string   | 本次操作的备注信息，例如                     | `required`   |
| `new_status`     | string   | 工单的新状态          | (可选)       |
| `new_priority`   | string   | **(仅管理员)** 工单的新优先级        | (可选)       |
| `new_appointment`| string   | 新的预约上门日期 (RFC3339 日期格式, 例如: "2025-12-08T00:00:00Z")     | (可选)       |
| `new_category`   | string   | 新工单类型                                                          | （可选）      |

**请求示例 (更新状态和备注)**:

```json
{
  "tid": 123,
  "remark": "已电话联系用户，指导其重启路由器后网络恢复正常。工单完成。",
  "new_status": "solved"
}
```

**请求示例 (重新安排上门时间)**:

```json
{
  "tid": 124,
  "remark": "用户今天没空，改约明天下午。",
  "new_status": "scheduled",
  "new_appointment": "2025-12-08T00:00:00Z"
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
  "msg": "new trace added"
}
```

### 失败响应

#### 400 Bad Request (请求错误)

- 请求体绑定失败、格式错误或未通过验证。
- 当设置了 `new_appointment` 但 `new_status` 不是 `scheduled` 时。

```json
{
  "success": false,
  "msg": "invalid request body: ...",
  "errType": 2
}
```

```json
{
  "success": false,
  "msg": "only appointed status can set appointment time",
  "errType": 2
}
```

#### 400 Bad Request (业务逻辑错误)

| `msg` 内容              | `errType` | 描述                                       |
| ----------------------- | --------- | ------------------------------------------ |
|"无法找到对应的工单"       | `logic`   | 提供的 `tid` 对应的工单不存在。            |
| "无法找到对应的网维成员"       | `logic`   | 操作者（网维人员）信息不存在。             |
|"您的工单状态更新请求不符合逻辑"| `logic`   | 新状态不符合工单状态流转规则 |

#### 403 Forbidden (权限错误)

- 当非网维人员尝试调用此接口时。
- 当非管理员尝试修改 `new_priority` 时。

```json
{
  "success": false,
  "msg": "only Network Support staff can access this API",
  "errType": 3
}
```

```json
{
  "success": false,
  "msg": "only admin can change ticket priority",
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

- 这是网维人员更新工单的主要方式。
- `remark` 字段是必填的，用于记录每次操作的内容。
- 只有在 `new_status` 设置为 `scheduled` 时，才能提供 `new_appointment` 字段。
- 工单状态的变更必须遵循预定义的逻辑流程。
- 只有管理员有权限调整工单的优先级。
- 这个 API 的底层实现是 `logic.AppendTrace`，与 `cancelTicket` API 共享部分逻辑。

