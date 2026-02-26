# 筛选用户 API

- **路径**: `/api/v3/filter_users`
- **方法**: `POST`
- **功能**: 根据指定条件筛选用户列表。

## 描述

此接口仅供管理员使用，用于根据一个或多个筛选条件查询和获取用户列表。可以根据姓名、电话、宿舍区、房间号、运营商和宽带账号进行组合查询。

## 认证

- **需要 JWT**: 是
- **权限要求**: 只有管理员 (`admin`) 权限的用户才能访问此 API。

## 请求

### 请求头

| Header          | 类型   | 描述                      |
| --------------- | ------ | ------------------------- |
| `Authorization` | string | `Bearer <your_jwt_token>` |
| `Content-Type`  | string | 必须是 `application/json` |

### 请求体 (JSON)

所有字段都是可选的，但至少需要提供一个字段进行筛选。

| 字段      | 类型     | 描述                                       |
| --------- | -------- | ------------------------------------------ |
| `name`    | string   | 用户姓名 (支持模糊匹配)                    |
| `phone`   | string   | 手机号码 (支持模糊匹配)                    |
| `block`   | string   | 宿舍区               |
| `room`    | string   | 房间号 (支持模糊匹配)                      |
| `isp`     | string   | 宽带运营商  |
| `account` | string   | 宽带账号 (支持模糊匹配)                    |

**请求示例**:

查询朝晖所有使用中国移动宽带的用户。

```json
{
  "block": "ZH",
  "isp": "mobile"
}
```

## 响应

### 成功响应 (200 OK)

响应体包含一个 `profiles` 数组，其中每个元素都是一个符合条件的用户信息对象。

| 字段       | 类型     | 描述                                       |
| ---------- | -------- | ------------------------------------------ |
| `success`  | boolean  | `true` 表示操作成功                        |
| `msg`      | string   | 成功的提示信息                             |
| `profiles` | array    | 用户信息对象数组                           |

`profiles` 数组中每个对象的结构与 `view_profile` API 返回的 `profile` 对象相同。

**响应示例**:

```json
{
  "success": true,
  "msg": "query success",
  "profiles": [
    {
      "sid": "20230002002",
      "name": "李四",
      "block": "QT",
      "access": "user",
      "room": "101",
      "phone": "13800138001",
      "isp": "中国移动",
      "account": "111222333",
      "wx": "user_openid_456"
    },
    {
      "sid": "20230002005",
      "name": "王五",
      "block": "ZH",
      "access": "operator",
      "room": "203",
      "phone": "13800138004",
      "isp": "中国移动",
      "account": "444555666",
      "wx": "operator_openid_789"
    }
  ]
}
```

如果找不到匹配的用户，`profiles` 数组将为空。

### 失败响应

#### 400 Bad Request (请求错误)

请求体绑定失败或格式错误。

```json
{
  "success": false,
  "msg": "cannot bind your request body: ...",
  "errType": 2
}
```

#### 403 Forbidden (权限错误)

当非管理员用户尝试调用此接口时返回。

```json
{
  "success": false,
  "msg": "only admin can access this API",
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

- 这是一个仅限管理员使用的功能。
- 所有筛选条件都是可选的，不提供的字段不会作为筛选依据。
- 多个筛选条件之间是 "与" (AND) 的关系。