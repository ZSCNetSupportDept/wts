# 注册 API

- **路径**: `/api/v3/register`
- **方法**: `POST`
- **功能**: 为新用户创建账户。

## 描述

此接口用于新用户进行注册。用户需要提供其个人信息和联系方式。服务器将验证所提供信息的有效性，并在数据库中创建一个新的用户记录。

## 认证

- **需要 JWT**: 是
- **权限要求**: 调用此接口的 JWT 负载中，用户的 `access` 级别必须是 `unregistered`。已注册用户无法调用此接口。

## 请求

### 请求头

| Header          | 类型   | 描述                               |
| --------------- | ------ | ---------------------------------- |
| `Authorization` | string | `Bearer <your_jwt_token>`          |
| `Content-Type`  | string | 必须是 `application/json`          |

### 请求体 (JSON)

| 字段      | 类型     | 描述                                       | 校验规则     |
| --------- | -------- | ------------------------------------------ | ------------ |
| `sid`     | string   | 用户的学号                                 | `required`   |
| `name`    | string   | 用户的真实姓名                             | `required`   |
| `block`   | string   | 宿舍楼                               | `required` ,`wts.block`枚举          |
| `room`    | string   | 房间号                                     | `required`   |
| `phone`   | string   | 手机号码      | `required` ，必须为有效的中国大陆11为手机号    |
| `isp`     | string   | 宽带运营商                    | `required`，`wts.isp`    |
| `account` | string   | 宽带账号                                   | `required`   |

### 命令行参数

如果后端开启跳过JWT模式的话，需要提供`op`参数作为OpenID

**请求示例**:

```json
{
  "sid": "20230001001",
  "name": "哈基米",
  "block": "ZH",
  "room": "1501",
  "phone": "13800138000",
  "isp": "mobile",
  "account": "12345678901"
}
```

## 响应

### 成功响应 (201 Created)

| 字段    | 类型    | 描述                 |
| ------- | ------- | -------------------- |
| `success` | boolean | `true` 表示操作成功  |
| `msg`     | string  | 成功的提示信息     |

**响应示例**:

```json
{
  "success": true,
  "msg": "register success~"
}
```

### 失败响应

#### 400 Bad Request (请求错误)

请求体绑定失败、格式错误或未通过验证。

```json
{
  "success": false,
  "msg": "invalid request body: Key: 'RegisterRequest.Sid' Error:Field validation for 'Sid' failed on the 'required' tag",
  "errType": 2
}
```

#### 400 Bad Request (业务逻辑错误)

由业务逻辑产生的已知错误。

| `msg` 内容                   | `errType` | 描述                               |
| ---------------------------- | --------- | ---------------------------------- |
| "抱歉，您输入的姓名或学号有误，如果确信所输入信息没有问题，请联系我们的工作人员。" | `logic`   | 提供的学号在学校记录中不存在     |
| "抱歉，您输入的姓名或学号有误，如果确信所输入信息没有问题，请联系我们的工作人员。"| `logic`   | 提供的学号与姓名不匹配           |
| "您已经注册了。如果您确信您还没有注册，请联系我们的工作人员。"  | `logic`   | 该学号已被注册                     |
|"抱歉，您所使用的联系电话已经被登记，请换一个不一样的电话号码。" | `logic`   | 该手机号已被其他用户注册           |
| "抱歉，您的微信已经注册过了，一个微信只能注册一个账号。" | `logic`   | 该微信账号已被其他用户注册         |

#### 403 Forbidden (权限错误)

当一个已经注册的用户尝试调用此接口时返回。

```json
{
  "success": false,
  "msg": "only unregistered users can access this API",
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

- 在调用此 API 前，用户必须已经通过微信授权流程，并获取了一个包含 `unregistered` 权限的 JWT。
- 提交的学号和姓名必须与学校数据库中的记录完全匹配，否则会注册失败。
- 手机号码和微信 OpenID 在系统中是唯一的，不能重复注册。