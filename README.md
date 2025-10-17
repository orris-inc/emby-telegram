# Emby Telegram Bot

一个用于管理 Emby 账号的 Telegram Bot，提供账号创建、续期、查询等功能。

## 特性

✅ **账号管理**
- 创建 Emby 账号（自动生成强密码）
- 查看账号列表和详情
- 续期账号
- 修改密码
- 设置家长控制评级
- 暂停/激活账号
- 自动同步到 Emby 服务器

✅ **用户管理**
- 自动用户注册
- 角色管理（管理员/普通用户）
- 用户封禁/解封
- 授权制账号创建（配额管理）
- 群组/私聊环境隔离

✅ **安全特性**
- bcrypt 密码加密
- 权限验证
- 所有权检查

✅ **Emby 集成**
- 自动同步账号到 Emby 服务器
- 检查 Emby 连接状态
- 查看账号同步状态
- 手动同步功能
- 离线模式支持（Emby 不可用时）

✅ **用户体验**
- 按钮式交互界面（Inline Keyboard + Reply Keyboard）
- 输入框固定快捷按钮
- 命令菜单支持（点击 / 快速选择命令）
- 对话式状态机（支持多步骤操作）

✅ **技术特性**
- 领域驱动设计（DDD）
- 遵循 Google Go 最佳实践
- 接口在消费端定义
- SQLite 数据库（可扩展至 MySQL/PostgreSQL）
- 结构化日志（Zap）
- 配置热加载（Viper）
- 优雅降级（Emby 离线模式）

## 快速开始

### 前置要求

- Go 1.24 或更高版本
- Telegram Bot Token（从 [@BotFather](https://t.me/BotFather) 获取）
- Emby 服务器（可选，用于账号同步）
- Emby API Key（可选，从 Emby 控制台获取）

### 安装

1. **克隆项目**

```bash
git clone https://github.com/yourusername/emby-telegram.git
cd emby-telegram
```

2. **安装依赖**

```bash
make deps
```

或者

```bash
go mod download
```

3. **配置 Bot**

创建配置文件或设置环境变量：

```bash
cp .env.example .env
```

编辑 `.env` 文件，设置你的配置：

```env
# Telegram Bot 配置
TELEGRAM_BOT_TOKEN=your_bot_token_here

# Emby 服务器配置（可选）
EMBY_SERVER_URL=http://localhost:8096
EMBY_API_KEY=your_emby_api_key_here
```

或者复制并编辑配置文件：

```bash
cp configs/config.example.yaml configs/config.yaml
```

编辑 `configs/config.yaml`：

```yaml
telegram:
  token: "your_bot_token_here"
  admin_ids:
    - 123456789  # 替换为你的 Telegram ID

emby:
  server_url: "http://localhost:8096"
  api_key: "your_emby_api_key_here"
  enable_sync: true
  sync_on_create: true
  sync_on_delete: true
```

### 运行

**开发模式:**

```bash
make run
```

**编译后运行:**

```bash
make build
./bin/emby-bot
```

## 项目结构

```
emby-telegram/
├── cmd/
│   └── server/          # 应用入口
├── internal/
│   ├── account/         # 账号领域
│   ├── user/            # 用户领域
│   ├── storage/         # 存储实现
│   │   └── sqlite/      # SQLite 实现
│   ├── bot/             # Telegram Bot
│   ├── config/          # 配置管理
│   └── logger/          # 日志封装
├── pkg/                 # 公共工具包
│   ├── crypto/          # 加密工具
│   ├── timeutil/        # 时间工具
│   └── validator/       # 验证工具
├── configs/             # 配置文件
├── data/                # 数据库文件
└── Makefile             # 构建脚本
```

## 使用指南

### 三种交互方式

本 Bot 支持三种操作方式，选择你喜欢的：

1. **固定按钮**（推荐）: 输入框下方有固定的快捷按钮，点击即可操作
   - 📋 我的账号
   - ➕ 创建账号
   - ❓ 帮助
   - 🔑 管理员菜单（仅管理员）

2. **命令菜单**: 点击输入框左侧的 `/` 按钮，快速选择命令

3. **手动输入**: 直接输入命令和参数

### 基础命令

- `/start` - 开始使用，显示主菜单
- `/help` - 查看帮助信息

### 账号管理

通过点击按钮或使用以下命令：

- `/myaccounts` - 查看我的所有账号
- `/quota` - 查看授权状态和配额信息（私聊）
- `/create <用户名>` - 创建新账号（需要授权）
- `/info <用户名>` - 查看账号详情
- `/renew <用户名> <天数>` - 续期账号
- `/changepassword <用户名> <新密码>` - 修改密码
- `/syncstatus <用户名>` - 查看账号同步状态

**按钮操作**：
- 点击 "📋 我的账号" 查看账号列表
- 点击账号名称查看详情
- 在账号详情页可以：续期、改密、设置评级、同步状态、删除（管理员）

### 管理员命令

**配额管理**（群组）：
- `/grant <用户> [配额]` - 授权用户创建账号（配额默认为 1）
  - 支持回复消息：`[回复目标消息] /grant` 或 `/grant 3`
  - 支持 @mention：`/grant @username` 或 `/grant @username 3`
  - 支持 ID：`/grant 123456789` 或 `/grant 123456789 3`
  - 设置为 0 可收回权限
  - 群组消息将在 30 秒后自动删除

**账号管理**：
- `/admin` - 显示管理员面板
- `/users [页码]` - 列出所有用户
- `/accounts [页码]` - 列出所有账号
- `/deleteaccount <用户名>` - 删除账号
- `/suspend <用户名>` - 暂停账号
- `/activate <用户名>` - 激活账号
- `/setrole <telegram_id> <admin|user>` - 设置用户角色
- `/blockuser <telegram_id>` - 封禁用户
- `/unblockuser <telegram_id>` - 解封用户
- `/stats` - 查看系统统计

### Emby 管理命令

- `/checkemby` - 检查 Emby 服务器连接状态
- `/syncaccount <用户名> <密码>` - 手动同步账号到 Emby
- `/embyusers` - 列出 Emby 服务器上的所有用户

### 使用示例

**授权流程（管理员在群组）**：
```
# 用户在群里申请
用户：我想要创建账号

# 管理员回复该消息授权（默认 1 个配额）
管理员：[回复该消息] /grant
Bot：✅ 已授权 @username 创建账号，配额: 1 个

# 或者授权指定配额
管理员：[回复该消息] /grant 3
Bot：✅ 已授权 @username 创建账号，配额: 3 个

# 或者使用 @mention
管理员：/grant @username
管理员：/grant @username 3

注意：群组消息将在 30 秒后自动删除
```

**用户创建账号（私聊）**：
```
# 查看配额
/quota

# 创建账号
/create john

# 查看账号信息
/info john

# 续期30天
/renew john 30

# 修改密码
/changepassword john newpassword123
```

## 配置说明

### 配置文件 (configs/config.yaml)

```yaml
app:
  name: "Emby Telegram Bot"
  version: "1.0.0"
  debug: false

telegram:
  token: ""  # Bot Token
  timeout: 60
  admin_ids:  # 管理员 Telegram ID
    - 123456789

database:
  driver: "sqlite"
  dsn: "data/emby.db"

account:
  default_expire_days: 30
  default_max_devices: 3
  username_prefix: ""
  password_length: 12
  max_accounts_per_user: 3
  max_accounts_per_admin: -1

emby:
  server_url: "http://localhost:8096"
  api_key: ""  # Emby API Key
  enable_sync: true
  sync_on_create: true
  sync_on_delete: true
  timeout: 30
  retry_count: 3

log:
  level: "info"
  output: "stdout"
```

### 环境变量

- `TELEGRAM_BOT_TOKEN` - Bot Token（必需）
- `EMBY_SERVER_URL` - Emby 服务器地址（可选）
- `EMBY_API_KEY` - Emby API Key（可选）
- `DB_DRIVER` - 数据库驱动（可选）
- `DB_DSN` - 数据库连接字符串（可选）
- `APP_ENV` - 应用环境（可选）
- `LOG_LEVEL` - 日志级别（可选）

### 账号配置说明

- `default_expire_days`: 默认账号有效期（天，默认 30）
- `default_max_devices`: 默认最大设备数（默认 3）
- `username_prefix`: 账号用户名前缀（留空则不添加前缀）
- `password_length`: 自动生成密码长度（默认 12）
- `default_quota`: 新用户默认配额（默认 0，需要管理员授权）
- `max_accounts_per_user`: 普通用户最大账号数量（技术上限，默认 99）
- `max_accounts_per_admin`: 管理员最大账号数量（默认 -1，表示无限制）

**授权制度**:
- 新用户注册后默认配额为 0，无法创建账号
- 管理员通过 `/grant` 命令在群组中授权用户
- 配额表示用户可以持有的账号总数（上限）
- 已创建的账号不受配额调整影响（收回配额不会删除已有账号）
- 超过配额后无法创建新账号，需删除现有账号或申请更多配额

### Emby 配置说明

- `enable_sync`: 是否启用 Emby 同步（默认 true）
- `sync_on_create`: 创建账号时自动同步到 Emby（默认 true）
- `sync_on_delete`: 删除账号时从 Emby 删除（默认 true）
- `timeout`: HTTP 请求超时时间（秒，默认 30）
- `retry_count`: 失败重试次数（默认 3）

**离线模式**: 如果 Emby 服务器不可用或配置未设置，Bot 会自动降级为离线模式，只在本地数据库管理账号。

## 开发

### Makefile 命令

```bash
make help          # 显示帮助
make build         # 编译项目
make run           # 运行项目
make test          # 运行测试
make test-cover    # 测试覆盖率
make clean         # 清理编译文件
make deps          # 下载依赖
make fmt           # 格式化代码
make vet           # 代码检查
make lint          # Lint 检查（需要 golangci-lint）
make all           # 完整构建流程
```

### 代码规范

本项目遵循 [Google Go Style Guide](https://google.github.io/styleguide/go/)：

- Context 作为第一个参数
- 接口定义在消费端
- 错误处理带上下文信息
- 使用 `%w` 包装错误
- 简洁的命名规范
- 表驱动测试

### 添加新功能

1. 在对应的领域包中定义实体和接口
2. 实现业务逻辑（Service 层）
3. 实现存储层（如需要）
4. 添加 Bot 命令处理器
5. 注册命令到 `internal/bot/command.go`

## 架构设计

### 领域驱动设计

- **领域层**: 实体、接口、领域服务 (`internal/account/`, `internal/user/`)
- **基础设施层**: 数据库、存储实现 (`internal/storage/`)
- **接口层**: Telegram Bot 命令处理 (`internal/bot/`)
- **横切关注点**: 日志、配置、工具 (`internal/logger/`, `internal/config/`, `pkg/`)

### 关键设计原则

✅ 按业务域组织代码
✅ 接口在消费端定义
✅ 依赖倒置原则
✅ 单一职责原则
✅ 错误处理规范化

## 安全建议

1. **保护 Bot Token**: 不要将 Token 提交到版本控制
2. **保护 Emby API Key**: 不要将 API Key 提交到版本控制
3. **数据库备份**: 定期备份 SQLite 数据库文件
4. **管理员设置**: 在配置文件中正确设置管理员 ID
5. **日志安全**: 不要在日志中记录密码等敏感信息
6. **网络安全**: 如果 Emby 服务器在公网，建议使用 HTTPS

## 故障排除

### 常见问题

**Q: Bot 无法启动？**

A: 检查配置文件中的 Bot Token 是否正确，确保网络可以访问 Telegram API。

**Q: 数据库迁移失败？**

A: 确保 `data/` 目录存在且有写入权限，检查 SQLite 是否正确安装。

**Q: 命令无响应？**

A: 检查日志输出，确认用户是否被封禁，命令参数是否正确。

**Q: Emby 同步失败？**

A: 使用 `/checkemby` 检查 Emby 服务器连接状态，确认 API Key 是否正确，使用 `/syncstatus <用户名>` 查看详细错误信息。

**Q: Bot 启动时 Emby 连接失败？**

A: Bot 会自动降级为离线模式，只在本地数据库管理账号。修复 Emby 配置后重启 Bot 即可恢复同步功能。

## 许可证

[MIT License](LICENSE)

## 贡献

欢迎提交 Issue 和 Pull Request！

## 致谢

- [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [GORM](https://gorm.io/)
- [Viper](https://github.com/spf13/viper)
- [Zap](https://github.com/uber-go/zap)