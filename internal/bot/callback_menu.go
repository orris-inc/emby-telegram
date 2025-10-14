// Package bot 菜单回调处理
package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"emby-telegram/internal/user"
)

// handleMenuCallback 处理菜单相关回调
func (b *Bot) handleMenuCallback(ctx context.Context, query *tgbotapi.CallbackQuery, parts []string, currentUser *user.User) CallbackResponse {
	if len(parts) < 2 {
		return CallbackResponse{Answer: "无效的菜单操作", ShowAlert: true}
	}

	subAction := parts[1]

	switch subAction {
	case "main":
		return b.showMainMenu(ctx, currentUser)
	case "help":
		return b.showHelp(ctx, currentUser)
	default:
		return CallbackResponse{Answer: "未知菜单", ShowAlert: true}
	}
}

// showMainMenu 显示主菜单
func (b *Bot) showMainMenu(ctx context.Context, currentUser *user.User) CallbackResponse {
	text := fmt.Sprintf(`👋 <b>欢迎使用 Emby 账号管理 Bot！</b>

您好，%s！

请选择要执行的操作：`, currentUser.DisplayName())

	keyboard := MainMenuKeyboard(currentUser.IsAdmin())

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// showHelp 显示帮助信息
func (b *Bot) showHelp(ctx context.Context, currentUser *user.User) CallbackResponse {
	help := `📚 <b>帮助信息</b>

<b>📋 我的账号</b>
查看您创建的所有 Emby 账号，包括账号状态、到期时间等信息。

<b>➕ 创建新账号</b>
创建一个新的 Emby 账号。系统会自动生成强密码，您可以选择账号的有效期。

<b>账号操作</b>
• <b>续期</b> - 延长账号有效期
• <b>改密</b> - 修改账号密码
• <b>同步</b> - 手动同步账号到 Emby 服务器
• <b>删除</b> - 删除账号（仅管理员）

<b>注意事项</b>
• 创建账号时会自动生成强密码，请妥善保管
• 账号过期后需要续期才能继续使用
• 账号会自动同步到 Emby 服务器
• 修改密码后需要使用新密码登录`

	if currentUser.IsAdmin() {
		help += `

<b>🔑 管理员功能</b>
• <b>用户管理</b> - 管理所有用户，设置角色和权限
• <b>账号管理</b> - 管理所有 Emby 账号
• <b>Emby 管理</b> - 管理 Emby 服务器连接
• <b>系统统计</b> - 查看系统使用统计`
	}

	keyboard := BackButton(CallbackMainMenu)

	return CallbackResponse{
		EditText:   help,
		EditMarkup: &keyboard,
	}
}

// handleBackCallback 处理返回按钮
func (b *Bot) handleBackCallback(ctx context.Context, query *tgbotapi.CallbackQuery, parts []string, currentUser *user.User) CallbackResponse {
	if len(parts) < 3 {
		return b.showMainMenu(ctx, currentUser)
	}

	target := parts[2]

	switch target {
	case "menu":
		return b.showMainMenu(ctx, currentUser)
	default:
		return b.showMainMenu(ctx, currentUser)
	}
}
