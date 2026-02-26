# 筛选工单 API

- **路径**: `/api/v3/filter_tickets`
- **方法**: `POST`
- **功能**: 根据多种条件筛选工单列表。

## 描述

此接口是为网维人员设计的强大工具，允许他们根据各种条件（如状态、报修人、时间范围、问题分类等）来查询和筛选工单。管理员拥有更广泛的查询范围。

## 认证

- **需要 JWT**: 是
- **权限要求**:
  - 必须是网维人员 (`operator` 或更高权限) 才能调用此 API。
  - 只有管理员 (`admin`) 才能使用 `scope: "all"` 来查询所有（包括已关闭的）工单。

## 请求

### 请求头

| Header          | 类型   | 描述                      |
| --------------- | ------ | ------------------------- |
| `Authorization` | string | `Bearer <your_jwt_token>` |
| `Content-Type`  | string | 必须是 `application/json` |

### 请求体 (JSON)

| 字段        | 类型     | 描述                                                                 | 默认值/规则  |
| ----------- | -------- | -------------------------------------------------------------------- | ------------ |
| `scope`     | string   | 查询范围: `active` (活动工单) 或 `all` (所有工单, **仅管理员**)。      | `active`     |
| `block`     | array of strings | 宿舍区 (Block) 列表。                                        | (可选)       |
| `issuer`    | string   | 报修人学号 (支持模糊匹配)。                                          | (可选)       |
| `category`  | string   | 问题分类。                                                           | (可选)       |
| `isp`       | string   | 宽带运营商。                                                         | (可选)       |
| `status`    | string   | 工单状态。                                                           | (可选)       |
| `newer_than`| string   | 时间范围下限 (RFC3339 格式)。                                        | (可选)       |
| `older_than`| string   | 时间范围上限 (RFC3339 格式)。                                        | `time.Now()` |

**请求示例 (查询所有活动中的、特定宿舍区、新装工单)**:

```json
{
  "scope": "active",
  "block": ["QT", "ZH"],
  "category": "first-install"
}
```

**请求示例 (管理员查询2025年11月之后创建的所有已完成工单)**:

```json
{
  "scope": "all",
  "status": "solved",
  "newer_than": "2025-11-01T00:00:00Z"
}
```

## 响应

### 成功响应 (200 OK)

响应体包含一个 `tickets` 数组，其中每个元素都是一个符合条件的工单对象。

| 字段      | 类型    | 描述             |
| --------- | ------- | ---------------- |
| `success` | boolean | `true` 表示操作成功 |
| `msg`     | string  | 成功的提示信息   |
| `tickets` | array   | 工单对象数组     |

`tickets` 数组中每个对象的结构与 `get_ticket` API 返回的工单对象结构相同。

**响应示例**:

```json
{
  "success": true,
  "msg": "query success",
  "tickets": [
    // ... 符合条件的工单对象列表 ...
  ]
}
```

如果找不到匹配的工单，`tickets` 数组将为空。

### 失败响应

#### 400 Bad Request (请求错误)

- 请求体绑定失败、格式错误或未通过验证。
- `scope` 值无效 (不是 "active" 或 "all")。
- `newer_than` 时间晚于 `older_than` 时间。

```json
{
  "success": false,
  "msg": "invalid scope value",
  "errType": 2
}
```

```json
{
  "success": false,
  "msg": "newerThan cannot be after olderThan",
  "errType": 2
}
```

#### 400 Bad Request (业务逻辑错误)

| `msg` 内容        | `errType` | 描述             |
| ----------------- | --------- | ---------------- |
| "无效的片区参数"   | `logic`   | `block` 参数无效。 |
| "Scope参数无效"  | `logic`   | `scope` 参数无效。|

#### 403 Forbidden (权限错误)

- 当非网维人员尝试调用此接口时。
- 当非管理员用户尝试使用 `scope: "all"` 时。

```json
{
  "success": false,
  "msg": "only admin can view all tickets",
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

- 这是网维人员和管理员的核心查询功能。
- 普通 `operator` 只能查询 `active` (活动) 工单，而 `admin` 可以查询 `all` (所有) 工单。
- `block` 参数用于按宿舍区进行筛选，可以传入一个包含多个宿舍区代码的数组。
- 所有筛选条件都是可选的，不提供的字段不会作为筛选依据。多个条件之间是 "与" (AND) 关系。
- 时间参数 `older_than` 如果不提供，默认会使用当前时间 (`time.Now()`)。
- `newer_than` 的时间不能晚于 `older_than` 的时间，否则会返回 400 错误。
- 时间范围查询是基于工单的创建时间。