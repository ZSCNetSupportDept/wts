# wtstool

构建工具：

```sh
make -C back tool
make -C back tool VERSION=v1.0.0
```

生成的程序位于 `back/build/wtstool`。以下示例假定它已加入 `PATH`。`--version` 显示构建版本，未指定版本时为 `dev`。

业务命令需要指定与服务程序相同的 YAML 配置文件，使用 `--config`、`-c` 或 `WTS_CONFIG`：

```sh
wtstool -c /path/to/config.yaml wx tag list
export WTS_CONFIG=/path/to/config.yaml
wtstool wx tag list
```

`--config` 优先于 `WTS_CONFIG`，工具不会自动搜索配置文件。配置项优先级为 `WTS_*` 环境变量、YAML 文件、原有默认值，例如 `WTS_WX_APPID`、`WTS_WX_APPSECRET` 和 `WTS_JWTKEY`。微信命令要求 AppID 和 AppSecret，JWT 命令要求非空 JWTKey。密钥不应通过命令行传入；环境变量和配置文件也应限制访问权限。

## 命令

```sh
wtstool wx menu set default menu.json
wtstool wx menu set operator menu.json
wtstool wx menu set admin menu.json
wtstool wx tag create "tag name"
wtstool wx tag list
wtstool wx tag list --output json
wtstool wx user tag OPENID operator
wtstool jwt issue --openid OPENID --access user --expire-minutes 30 \
 --sid STUDENT_ID --username USERNAME --avatar AVATAR_URL --name NAME
wtstool jwt parse TOKEN
```

`jwt issue` 的 `openid`、`access` 和 `expire-minutes` 必须显式指定；其他字段默认为空。有效期单位是分钟，可以使用零或负数生成过期令牌。JWT 以单行输出到 stdout，请妥善保管。

`jwt parse TOKEN` 将 JWT 头部和载荷解码为 JSON，并使用配置中的 `JWTKey` 验证 HS256 签名。输出的 JSON 包含 `header`、`payload` 和布尔值 `verified`。签名不匹配、算法不受支持或签名格式错误时，仍会输出解码结果和 `verified: false`，并返回非零退出码；配置错误或头部、载荷无法解码时不输出 JSON。

该验证只检查签名，不检查有效期或其他声明；已过期或尚未生效的令牌也可能显示 `verified: true`。这只表明签名与配置密钥匹配，不能证明密钥由谁持有。未验证的内容不可信。令牌可能包含敏感信息，命令行参数可能出现在 shell 历史或进程列表中，应注意保护。

`wx tag list -o json` 输出 JSON 数组，每个元素包含数值 `id`、字符串 `name` 和数值 `count`；没有标签时输出 `[]`。错误和提示不混入 JSON。

设置 `default` 菜单会替换公众号默认菜单；`operator` 和 `admin` 会新增条件菜单，匹配规则由 JSON 文件中的 `matchrule` 决定，命令不会自动修改该规则。重复执行条件菜单设置可能创建多个菜单。

用户标签切换沿用服务端逻辑：先移除旧标签再添加新标签；操作不是事务，中途失败不会自动回滚。`wx menu get` 目前尚未实现，会返回错误。

旧版 `-a "动作 参数..."` 已移除，迁移关系如下：

| 旧动作          | 新命令                  |
| --------------- | ----------------------- |
| `set_wx_menu`   | `wx menu set`           |
| `set_wx_tags`   | `wx tag create`         |
| `get_wx_tags`   | `wx tag list`           |
| `change_wx_tag` | `wx user tag`           |
| `gen_jwt_key`   | `jwt issue`             |
| `get_wx_menu`   | `wx menu get`，尚未实现 |

工具不再接受 `--JWTKey`、`--WX.AppID` 等服务端配置覆盖参数，改用配置文件或 `WTS_*` 环境变量；服务程序的参数保持不变。

## 帮助与补全

```sh
wtstool --help
wtstool wx menu set --help
source <(wtstool completion bash)  # bash
source <(wtstool completion zsh)   # zsh
wtstool completion fish | source   # fish
wtstool completion powershell      # PowerShell 补全脚本
```

帮助、版本和补全无需配置文件。开发配置存在时，也可以通过 `make -C back start-tool ARGS="wx tag list"` 执行命令。
