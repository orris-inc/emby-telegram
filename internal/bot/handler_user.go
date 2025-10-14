// Package bot 用户命令处理器
package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"emby-telegram/pkg/timeutil"
)

// handleStart 处理 /start 命令
func (b *Bot) handleStart(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	user, err := b.userService.GetByTelegramID(ctx, msg.From.ID)
	if err != nil {
		return "", fmt.Errorf("获取用户信息失败: %w", err)
	}

	text := fmt.Sprintf(`👋 <b>欢迎使用 Emby 账号管理 Bot！</b>

您好，%s！

请使用下方按钮或输入命令进行操作：`, user.DisplayName())

	// 发送带 Reply Keyboard 的消息（显示在输入框下方）
	keyboard := MainReplyKeyboard(user.IsAdmin())
	replyMsg := tgbotapi.NewMessage(msg.Chat.ID, text)
	replyMsg.ParseMode = "HTML"
	replyMsg.ReplyMarkup = keyboard

	if _, err := b.api.Send(replyMsg); err != nil {
		return "", fmt.Errorf("发送消息失败: %w", err)
	}

	return "", nil // 返回空，因为已经发送了消息
}

// handleHelp 处理 /help 命令
func (b *Bot) handleHelp(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	user, err := b.userService.GetByTelegramID(ctx, msg.From.ID)
	if err != nil {
		return "", err
	}

	help := `📚 <b>帮助信息</b>

<b>基础命令:</b>
/start - 开始使用
/help - 查看帮助

<b>账号管理:</b>
/myaccounts - 查看我的所有账号
/create &lt;用户名&gt; - 创建新账号
/info &lt;用户名&gt; - 查看账号详情
/renew &lt;用户名&gt; &lt;天数&gt; - 续期账号
/changepassword &lt;用户名&gt; &lt;新密码&gt; - 修改密码
/syncstatus &lt;用户名&gt; - 查看账号同步状态

<b>使用示例:</b>
<code>/create john</code> - 创建名为 john 的账号
<code>/info john</code> - 查看 john 的账号信息
<code>/renew john 30</code> - 为 john 续期 30 天
<code>/changepassword john newpass123</code> - 修改密码

<b>注意事项:</b>
• 创建账号时会自动生成强密码
• 账号过期后需要续期才能继续使用
• 账号会自动同步到 Emby 服务器`

	if user.IsAdmin() {
		help += "\n\n🔑 您是管理员，使用 /admin 查看管理命令"
	}

	return help, nil
}

// handleMyAccounts 处理 /myaccounts 命令
func (b *Bot) handleMyAccounts(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	user, err := b.userService.GetByTelegramID(ctx, msg.From.ID)
	if err != nil {
		return "", err
	}

	accounts, err := b.accountService.ListByUser(ctx, user.ID)
	if err != nil {
		return "", fmt.Errorf("获取账号列表失败: %w", err)
	}

	if len(accounts) == 0 {
		return "您还没有创建任何账号\n\n使用 /create &lt;用户名&gt; 创建新账号", nil
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📋 <b>您的账号列表</b> (共 %d 个)\n\n", len(accounts)))

	for i, acc := range accounts {
		status := getStatusEmoji(string(acc.Status))
		expireInfo := timeutil.FormatExpireTime(acc.ExpireAt)

		builder.WriteString(fmt.Sprintf("%d. <b>%s</b> %s\n", i+1, acc.Username, status))
		builder.WriteString(fmt.Sprintf("   到期时间: %s\n", expireInfo))
		builder.WriteString(fmt.Sprintf("   最大设备数: %d\n", acc.MaxDevices))
		builder.WriteString("\n")
	}

	builder.WriteString("💡 使用 <code>/info &lt;用户名&gt;</code> 查看详细信息")

	return builder.String(), nil
}

// handleCreateAccount 处理 /create 命令
func (b *Bot) handleCreateAccount(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if !hasArg(args, 1) {
		return "❌ 请提供用户名\n\n使用方法: <code>/create &lt;用户名&gt;</code>\n例如: <code>/create john</code>", nil
	}

	username := getArg(args, 0)

	user, err := b.userService.GetByTelegramID(ctx, msg.From.ID)
	if err != nil {
		return "", err
	}

	// 创建账号
	acc, plainPassword, err := b.accountService.Create(ctx, username, user.ID)
	if err != nil {
		return "", fmt.Errorf("创建账号失败: %w", err)
	}

	expireInfo := timeutil.FormatExpireTime(acc.ExpireAt)

	return fmt.Sprintf(`✅ <b>账号创建成功！</b>

<b>用户名:</b> <code>%s</code>
<b>密码:</b> <code>%s</code>
<b>有效期:</b> %s
<b>最大设备数:</b> %d

⚠️ <b>重要提示:</b>
• 请立即保存密码，此密码只显示一次
• 可使用 /changepassword 修改密码
• 使用 /info %s 查看详细信息
`,
		acc.Username,
		plainPassword,
		expireInfo,
		acc.MaxDevices,
		acc.Username,
	), nil
}

// handleAccountInfo 处理 /info 命令
func (b *Bot) handleAccountInfo(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if !hasArg(args, 1) {
		return "❌ 请提供用户名\n\n使用方法: <code>/info &lt;用户名&gt;</code>", nil
	}

	username := getArg(args, 0)

	user, err := b.userService.GetByTelegramID(ctx, msg.From.ID)
	if err != nil {
		return "", err
	}

	acc, err := b.accountService.GetByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("获取账号信息失败: %w", err)
	}

	// 检查所有权
	if err := b.accountService.CheckOwnership(ctx, acc.ID, user.ID); err != nil {
		return "❌ 您没有权限查看此账号", nil
	}

	status := getStatusEmoji(string(acc.Status))
	expireInfo := timeutil.FormatExpireTime(acc.ExpireAt)
	createdAt := timeutil.FormatDateTime(acc.CreatedAt)

	return fmt.Sprintf(`📝 <b>账号详情</b>

<b>用户名:</b> <code>%s</code>
<b>状态:</b> %s %s
<b>有效期:</b> %s
<b>最大设备数:</b> %d
<b>创建时间:</b> %s

💡 <b>可用操作:</b>
• /renew %s &lt;天数&gt; - 续期账号
• /changepassword %s &lt;新密码&gt; - 修改密码
`,
		acc.Username,
		status,
		acc.Status,
		expireInfo,
		acc.MaxDevices,
		createdAt,
		acc.Username,
		acc.Username,
	), nil
}

// handleRenewAccount 处理 /renew 命令
func (b *Bot) handleRenewAccount(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if !hasArg(args, 2) {
		return "❌ 参数不足\n\n使用方法: <code>/renew &lt;用户名&gt; &lt;天数&gt;</code>\n例如: <code>/renew john 30</code>", nil
	}

	username := getArg(args, 0)
	daysStr := getArg(args, 1)

	days, err := strconv.Atoi(daysStr)
	if err != nil {
		return "❌ 天数必须是有效的数字", nil
	}

	user, err := b.userService.GetByTelegramID(ctx, msg.From.ID)
	if err != nil {
		return "", err
	}

	acc, err := b.accountService.GetByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("获取账号信息失败: %w", err)
	}

	// 检查所有权
	if err := b.accountService.CheckOwnership(ctx, acc.ID, user.ID); err != nil {
		return "❌ 您没有权限操作此账号", nil
	}

	// 续期
	if err := b.accountService.Renew(ctx, acc.ID, days); err != nil {
		return "", fmt.Errorf("续期失败: %w", err)
	}

	// 重新获取更新后的账号信息
	acc, _ = b.accountService.Get(ctx, acc.ID)
	expireInfo := timeutil.FormatExpireTime(acc.ExpireAt)

	return fmt.Sprintf(`✅ <b>续期成功！</b>

账号 <b>%s</b> 已续期 %d 天
新的到期时间: %s
`,
		acc.Username,
		days,
		expireInfo,
	), nil
}

// handleChangePassword 处理 /changepassword 命令
func (b *Bot) handleChangePassword(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if !hasArg(args, 2) {
		return "❌ 参数不足\n\n使用方法: <code>/changepassword &lt;用户名&gt; &lt;新密码&gt;</code>\n例如: <code>/changepassword john newpass123</code>", nil
	}

	username := getArg(args, 0)
	newPassword := getArg(args, 1)

	user, err := b.userService.GetByTelegramID(ctx, msg.From.ID)
	if err != nil {
		return "", err
	}

	acc, err := b.accountService.GetByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("获取账号信息失败: %w", err)
	}

	// 检查所有权
	if err := b.accountService.CheckOwnership(ctx, acc.ID, user.ID); err != nil {
		return "❌ 您没有权限操作此账号", nil
	}

	// 修改密码
	if err := b.accountService.ChangePassword(ctx, acc.ID, newPassword); err != nil {
		return "", fmt.Errorf("修改密码失败: %w", err)
	}

	return fmt.Sprintf(`✅ <b>密码修改成功！</b>

账号 <b>%s</b> 的密码已更新
新密码: <code>%s</code>

⚠️ 请妥善保管新密码
`,
		acc.Username,
		newPassword,
	), nil
}

// getStatusEmoji 根据状态返回 emoji
func getStatusEmoji(status string) string {
	switch status {
	case "active":
		return "✅"
	case "suspended":
		return "⏸️"
	case "expired":
		return "❌"
	default:
		return "❓"
	}
}
