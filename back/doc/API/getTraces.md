# 获取工单追踪记录 API

- **路径**: `/api/v3/get_traces`
- **方法**: `GET`
- **功能**: 获取指定工单的所有处理和状态变更记录。

## 描述

此接口用于查询一个特定工单从创建到当前状态的所有追踪记录（Traces）。用户可以查看自己工单的处理进度，网维人员则可以查看所有他们有权访问的工单的完整历史记录。

## 认证

- **需要 JWT**: 是
- **权限要求**:
  - 工单的创建者 (`user` 权限) 可以查看该工单的追踪记录。
  - 网维人员 (`operator` 或更高权限) 可以查看任何工单的追踪记录。

## 请求

### 请求头

| Header          | 类型   | 描述                      |
| --------------- | ------ | ------------------------- |
| `Authorization` | string | `Bearer <your_jwt_token>` |

### 查询参数

| 参数 | 类型   | 描述             | 是否必须 |
| ---- | ------ | ---------------- | -------- |
| `tid`  | string | 要查询的工单 ID。 | 是       |

**请求示例**:

```
GET /api/v3/get_traces?tid=123
```

## 响应

### 成功响应 (200 OK)

响应体包含一个 `traces` 数组，其中每个元素都是一条工单处理记录。

| 字段        | 类型    | 描述                                                         |
| ----------- | ------- | ------------------------------------------------------------ |
| `success`   | boolean | `true` 表示操作成功                                          |
| `msg`       | string  | 成功的提示信息                                               |
| `traces`    | array   | 追踪记录对象数组                                             |
| `trace.opid`          | integer | 操作记录的唯一 ID                                            |
| `trace.tid`           | integer | 所属工单的 ID                                                |
| `trace.updated_at`    | string  | 本次记录的更新时间                                           |
| `trace.op`            | string  | 操作人员的 ID (如果是用户自己取消，可能为特殊值如 "-1")      |
| `trace.op_name`       | string  | 操作人的姓名                                                |
| `trace.new_status`    | string  | 本次操作后工单的新状态                                       |
| `trace.new_priority`  | string  | 本次操作后工单的新优先级                                     |
| `trace.new_appointment`| string | 本次操作后工单的新预约时间                                   |
| `trace.new_category`   | string | 本次操作后工单的类型                                          |
| `trace.remark`        | string  | 本次操作的备注信息                                           |

**响应示例**:

```json
{
  "success": true,
  "msg": "query success",
  "traces": [
    {
      "opid": 1,
      "tid": 123,
      "updated_at": "2025-12-06T10:00:00Z",
      "op": "-1",
      "new_status": "fresh",
      "new_priority": "mainline",
      "new_appointment": "",
      "remark": "工单已创建"
    },
    {
      "opid": 3,
      "tid": 123,
      "updated_at": "2025-12-06T15:30:00Z",
      "op": "W007",
      "new_status": "solved",
      "new_priority": "",
      "new_appointment": "",
      "remark": "问题已解决，用户确认网络恢复正常。"
    }
  ]
}
```

如果工单不存在，`traces` 数组将为空（或返回 `ErrNoSuchTicket` 错误）。

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

#### 400 Bad Request (业务逻辑错误)

| `msg` 内容         | `errType` | 描述                          |
| ------------------ | --------- | ----------------------------- |
| "无法找到对应的工单"   | `logic`   | 提供的 `tid` 对应的工单不存在。 |

#### 403 Forbidden (权限错误)

- 当用户尝试查看不属于自己的工单，并且该用户不是网维人员时返回。
- 当非活跃用户（如 `unregistered`）尝试调用此接口时返回。

```json
{
  "success": false,
  "msg": "you can only view ticket traces of your own",
  "errType": 3
}
```

#### 500 Internal Server Error (服务器内部错误)

当发生未预料到的数据库查询错误或其他内部错误时返回。

```json
{
  "success": false,
  "msg": "system met a uncaught error,please view logs.",
  "errType": 1
}
```

## 注意事项

- 此 API 返回的记录是按时间顺序排列的，可以清晰地展示工单的整个生命周期。
- `op` 字段标识了执行该操作的人员。
- 这是一个只读操作，不会对工单或追踪记录进行任何修改。
