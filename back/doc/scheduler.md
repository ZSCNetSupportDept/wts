# 排班系统

## 部署

先备份数据库，然后在现有应用数据库中以数据库所有者执行：

```sh
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f back/src/db/migrations/001_scheduler.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f back/src/db/migrations/003_scheduler_notifications.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f back/src/db/migrations/004_weekly_schedule.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f back/src/db/migrations/005_remove_scheduler_notifications.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f back/src/db/migrations/006_monthly_schedule.sql
```

微信聊天命令沿用原仓库实现：`/auth` 查询已有微信账号的成员身份并更新菜单；`/deauth` 移除微信菜单标签、恢复默认菜单。两条命令均不修改 OpenID 与工号的绑定，也不改变成员权限。

排班不发送微信通知。组长在换班大厅查看和确认请求，蹭班仍按原权限直接添加。迁移 003 保留为历史记录；迁移 005 删除排班通知队列、模板读取函数和排班模板配置，不影响原有工单通知。

新数据库先按现有初始化流程建立 `data`、`wts`、`scheduler` schema 和 `app` 角色，再执行上述迁移。
迁移在一个事务中创建表、索引和权限。重复执行会报错，已执行成功的数据库不要再次运行。
迁移只调整排班结构，不导入图片中的姓名，不修改任何成员权限。
前端页面分别为 `/op/scheduler/`、`/admin/scheduler/`，换班大厅位于 `/op/scheduler/swaps/`。

回退应用前，先导出 `scheduler.duty_days`、`duty_members`、`availability_exceptions`、`swap_requests`、`audit_log`。
旧版本不使用新增表，回退程序时可保留这些表；不要删除既有 `scheduler.freeday` 或成员数据。

## 权限

小组、片区和发布状态按星期保存为模板，仅在所选日期起一个日历月内沿用（结束日不含）。保存时生成窗口内同星期班次，历史日期不因后续模板编辑改变。一次性换班或蹭班的日期保持独立，再次保存片区、小组或发布也不写入模板。新的生效日期建立后，修改更早的模板不会覆盖较新的版本。迁移 006 为现有模板设置结束日期，清理范围外的未来自动班次；有换班引用的班次保留为未发布记录，单次例外、手动配置和历史班次保留。

每个请求先验证 JWT 中的 OpenID，再从数据库读取当前 `users.op`、`operators.access`。
旧 JWT 内的角色不参与授权。已离职、普通用户、前成员和未注册用户无法访问排班业务。
排班接口不支持 `Debug.SkipJWTAuth`；本地开发和测试也需要有效 JWT。

| 操作 | 正式/实习成员 | 组长 | 科长/dev/api |
| --- | --- | --- | --- |
| 查看全组已发布排班、自己的班次 | 允许 | 允许 | 允许 |
| 填写本人的固定空闲日、日期例外 | 允许 | 允许 | 允许 |
| 查看草稿、候选成员空闲情况 | 禁止 | 仅本人负责日期 | 全部日期 |
| 配置日期、小组成员、值班组长 | 禁止 | 禁止 | 允许 |
| 安排片区、沿用、随机、发布、蹭班、查看日期审计 | 禁止 | 仅本人负责日期 | 全部日期 |
| 申请一次性换班 | 本人正式班次 | 本人非值班组长的正式班次 | 本人非值班组长的正式班次 |
| 确认/拒绝换班 | 禁止 | 对应日期的值班组长 | 禁止代替组长确认 |

值班组长必须是当前 `group-leader`；真实成员工号用于排班，系统特殊工号 `-1`、`-2` 禁止参与。
片区负责人是当天的职责标记，不会授予管理员权限。图例背景为组长橙、API/dev 浅蓝、负责人黄、正式成员绿、实习成员白、科长蓝；同时显示文字标签。

## 业务规则

一个日期对应一个 16:30–18:00 的值班班次，日期统一按北京时间解释。
科长/dev/api 先配置当天小组，值班组长再为小组成员通过片区下拉框安排值班。
值班组长自动加入小组。同一人同一天只能有一条记录，每片区至多一名负责人。
修改小组或片区会转为草稿，重新发布后成员可见；蹭班和已确认换班直接更新所选班次。
修改已发布排班需填写原因。空闲日期不满足时，发布需要确认特殊安排并填写原因。
成员调整空闲时间只影响可用性提示，不自动取消已有值班。

沿用：读取该值班组长在所选日期之前最近一次已发布排班，按工号匹配当前小组，排除蹭班人员。
新成员没有历史安排时保留其当前片区，否则待手动选择；旧成员不会被自动加回小组。
随机：以同一历史班次各片区的总人数为配额，固定非实习成员，只随机分配当前小组的实习成员。
人数或固定成员片区与历史配额不兼容时拒绝生成。生成只返回预览，需要手动保存。

蹭班：输入有效工号和片区加入一个具体日期，独立标记为蹭班，不参与下一次沿用和随机配额。
不能重复添加已经在当天班次中的成员；移除接口只允许移除蹭班记录。

换班：申请人填写本人的原日期和预期换班日期，不指定对象工号。请求以 `open` 状态进入大厅。
预期日期的正式班次组员可查看并接下请求，接单身份由后端登录身份确定；其他日期成员不能接单。
首位成功接单者锁定为换班对象，请求转为 `pending`，此时才交给双方组长确认。并发接单只有一人成功。
双方必须在不同的未来/当天已发布正式班次中。
不能交换值班组长或蹭班；不能造成同一天重复值班。换班不会修改每周空闲日、日期例外或其他日期。
请求绑定两天的版本。两位对应组长分别确认后，在一个事务内交换两人的日期片区岗位；若两天组长相同，一次确认覆盖两方。
任一对应组长可以拒绝，申请人可以撤销。原排班任何修改会使待确认申请失效，需要重新申请。
成员背景按其实际角色显示，岗位的片区负责人标记随岗位交换。

所有写入使用日期锁和版本号，冲突返回 409，不覆盖他人的修改。日期变更记录保存前后状态、操作工号及原因。

## API

统一前缀 `/api/v3/scheduler`，请求带 `Authorization: Bearer <JWT>`。
成功为 `{ "success": true, "data": ... }`；失败使用 HTTP 状态码和 `{ "success": false, "error_type": 2, "msg": "..." }`。
403 的 `error_type` 为 3，500 为 1。请求体上限 64 KiB。

| 方法与路径 | 内容 |
| --- | --- |
| `GET /?month=YYYY-MM` | 当前成员、权限、可见月排班、本人空闲设置 |
| `GET /staff?date=YYYY-MM-DD` | 管理所选日期所需的候选成员，不返回联系方式或 OpenID |
| `GET /audit?date=YYYY-MM-DD` | 所选日期最近 100 条变更记录 |
| `PUT /availability` | `{free_days:[1,2], date:"YYYY-MM-DD", available:true/false/null}`，null 清除日期例外 |
| `PUT /group` | `{date, revision, leader_wid, wids:[...], reason}`；新日期 revision 为 0 |
| `PUT /assignments` | `{date, revision, assignments:[{wid,zone,area_leader}], reason}`，覆盖全部正式班次成员的片区 |
| `POST /publish` | `{date, revision, override_unavailable, reason}` |
| `POST /suggest` | `{date, revision, mode:"reuse"/"random"}` |
| `POST /extra` | `{date, revision, wid, zone, reason}` |
| `DELETE /extra` | 请求体 `{date, revision, wid, reason}` |
| `GET /swaps?month=YYYY-MM` | 本人参与、预期日期本组员可查看的待接单请求、组长相关或管理人员可见的申请；任一日期在本月即显示 |
| `POST /swaps` | `{date_a, date_b, reason}`，申请人由登录身份确定 |
| `POST /swaps/review` | `{id, action:"claim"/"approve"/"reject"/"cancel"}`；claim 为预期日期组员接单 |

换班状态：`open`（待接单）、`pending`（已接单待组长确认）、`approved`、`rejected`、`canceled`、`stale`。
版本号使用服务器返回值，不假设首次创建的具体版本。

## 验证

常规后端检查：在 `back/src` 执行 `go test ./...`。
独立数据库测试脚本 `back/scripts/test-scheduler.mjs` 使用临时目录、随机本机端口和专用 `wts_scheduler_test` 数据库。
它会执行原始初始化脚本、排班迁移，并以 `app` 用户测试真实查询、权限与双组长换班。
需在独立临时目录安装 `embedded-postgres`，通过 `EMBEDDED_POSTGRES_MODULE` 指定其 `dist/index.js`；可用 `GO_BINARY` 指定 Go。
脚本结束时停止数据库，保留临时数据便于诊断。测试代码拒绝非本机或非专用测试数据库，禁止指向生产环境。

前端在 `front` 执行 `npm run check` 和 `npm run test:e2e -- e2e/scheduler.test.ts`。
浏览器测试拦截所有 API，不访问真实后端，验证日历标记、角色界面、预览不自动保存、换班确认和图例底色。
