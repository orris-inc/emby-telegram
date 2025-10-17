// Package bot 状态输入处理
package bot

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"emby-telegram/internal/account"
	"emby-telegram/internal/user"
	"emby-telegram/pkg/timeutil"
)

// handleStateInput 处理用户在状态机中的输入
func (b *Bot) handleStateInput(ctx context.Context, msg *tgbotapi.Message, currentUser *user.User, state UserState, stateData map[string]interface{}) {
	// 检查是否取消
	if msg.Text == "/cancel" {
		b.stateMachine.ClearState(currentUser.TelegramID)
		b.reply(msg.Chat.ID, "操作已取消")
		return
	}

	switch state {
	case StateWaitingUsername:
		b.handleUsernameInput(ctx, msg, currentUser)
	case StateWaitingPassword:
		b.handlePasswordInput(ctx, msg, currentUser, stateData)
	case StateWaitingDays:
		b.handleDaysInput(ctx, msg, currentUser, stateData)
	default:
		b.stateMachine.ClearState(currentUser.TelegramID)
		b.reply(msg.Chat.ID, "会话已过期，请重新开始")
	}
}

// handleUsernameInput 处理用户名输入
func (b *Bot) handleUsernameInput(ctx context.Context, msg *tgbotapi.Message, currentUser *user.User) {
	username := strings.TrimSpace(msg.Text)

	// 验证用户名格式
	if !isValidUsername(username) {
		b.reply(msg.Chat.ID, `❌ 用户名格式不正确

<b>用户名要求：</b>
• 只能包含字母、数字和下划线
• 长度 3-20 个字符

请重新输入，或发送 /cancel 取消：`)
		return
	}

	// 检查用户名是否已存在
	if _, err := b.accountService.GetByUsername(ctx, username); err == nil {
		b.reply(msg.Chat.ID, fmt.Sprintf("❌ 用户名 <code>%s</code> 已存在，请使用其他用户名：", username))
		return
	}

	// 创建账号
	acc, plainPassword, err := b.accountService.Create(ctx, username, currentUser.ID)
	if err != nil {
		var errMsg string
		if errors.Is(err, account.ErrNotAuthorized) {
			errMsg = "❌ 您尚未获得创建账号的授权\n\n请在管理群组联系管理员申请"
		} else if errors.Is(err, account.ErrAccountLimitExceeded) {
			errMsg = fmt.Sprintf("❌ %v\n\n如需更多配额，请联系管理员", err)
		} else {
			errMsg = fmt.Sprintf("❌ 创建账号失败: %v", err)
		}
		b.reply(msg.Chat.ID, errMsg)
		b.stateMachine.ClearState(currentUser.TelegramID)
		return
	}

	// 清除状态
	b.stateMachine.ClearState(currentUser.TelegramID)

	expireInfo := timeutil.FormatExpireTime(acc.ExpireAt)

	text := fmt.Sprintf(`✅ <b>账号创建成功！</b>

<b>用户名:</b> <code>%s</code>
<b>密码:</b> <code>%s</code>
<b>有效期:</b> %s
<b>最大设备数:</b> %d

⚠️ <b>重要提示:</b>
• 请立即保存密码，此密码只显示一次
• 可通过账号详情页面修改密码`,
		acc.Username,
		plainPassword,
		expireInfo,
		acc.MaxDevices,
	)

	// 添加操作按钮
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 查看详情", CallbackAccountInfo+":"+uintToStr(acc.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 我的账号", CallbackMyAccounts+":1"),
			tgbotapi.NewInlineKeyboardButtonData("⬅️ 主菜单", CallbackMainMenu),
		),
	)

	replyMsg := tgbotapi.NewMessage(msg.Chat.ID, text)
	replyMsg.ParseMode = "HTML"
	replyMsg.ReplyMarkup = keyboard

	if _, err := b.api.Send(replyMsg); err != nil {
		b.reply(msg.Chat.ID, text)
	}
}

// handlePasswordInput 处理密码输入
func (b *Bot) handlePasswordInput(ctx context.Context, msg *tgbotapi.Message, currentUser *user.User, stateData map[string]interface{}) {
	password := strings.TrimSpace(msg.Text)

	// 验证密码格式
	if len(password) < 6 {
		b.reply(msg.Chat.ID, `❌ 密码长度至少 6 个字符

请重新输入，或发送 /cancel 取消：`)
		return
	}

	// 获取账号 ID
	accountID, ok := stateData["account_id"].(uint)
	if !ok {
		b.stateMachine.ClearState(currentUser.TelegramID)
		b.reply(msg.Chat.ID, "❌ 会话已过期，请重新开始")
		return
	}

	// 获取账号信息
	acc, err := b.accountService.Get(ctx, accountID)
	if err != nil {
		b.stateMachine.ClearState(currentUser.TelegramID)
		b.reply(msg.Chat.ID, fmt.Sprintf("❌ 获取账号信息失败: %v", err))
		return
	}

	// 检查所有权
	if err := b.accountService.CheckOwnership(ctx, acc.ID, currentUser.ID); err != nil {
		b.stateMachine.ClearState(currentUser.TelegramID)
		b.reply(msg.Chat.ID, "❌ 您没有权限操作此账号")
		return
	}

	// 修改密码
	if err := b.accountService.ChangePassword(ctx, acc.ID, password); err != nil {
		b.stateMachine.ClearState(currentUser.TelegramID)
		b.reply(msg.Chat.ID, fmt.Sprintf("❌ 修改密码失败: %v", err))
		return
	}

	// 清除状态
	b.stateMachine.ClearState(currentUser.TelegramID)

	text := fmt.Sprintf(`✅ <b>密码修改成功！</b>

账号 <b>%s</b> 的密码已更新
新密码: <code>%s</code>

⚠️ 请妥善保管新密码`,
		acc.Username,
		password,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 返回账号详情", CallbackAccountInfo+":"+uintToStr(acc.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 我的账号", CallbackMyAccounts+":1"),
		),
	)

	replyMsg := tgbotapi.NewMessage(msg.Chat.ID, text)
	replyMsg.ParseMode = "HTML"
	replyMsg.ReplyMarkup = keyboard

	if _, err := b.api.Send(replyMsg); err != nil {
		b.reply(msg.Chat.ID, text)
	}
}

// handleDaysInput 处理天数输入
func (b *Bot) handleDaysInput(ctx context.Context, msg *tgbotapi.Message, currentUser *user.User, stateData map[string]interface{}) {
	daysStr := strings.TrimSpace(msg.Text)

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 || days > 3650 {
		b.reply(msg.Chat.ID, `❌ 请输入有效的天数（1-3650）

请重新输入，或发送 /cancel 取消：`)
		return
	}

	// 获取账号 ID
	accountID, ok := stateData["account_id"].(uint)
	if !ok {
		b.stateMachine.ClearState(currentUser.TelegramID)
		b.reply(msg.Chat.ID, "❌ 会话已过期，请重新开始")
		return
	}

	// 获取账号信息
	acc, err := b.accountService.Get(ctx, accountID)
	if err != nil {
		b.stateMachine.ClearState(currentUser.TelegramID)
		b.reply(msg.Chat.ID, fmt.Sprintf("❌ 获取账号信息失败: %v", err))
		return
	}

	// 检查所有权
	if err := b.accountService.CheckOwnership(ctx, acc.ID, currentUser.ID); err != nil {
		b.stateMachine.ClearState(currentUser.TelegramID)
		b.reply(msg.Chat.ID, "❌ 您没有权限操作此账号")
		return
	}

	// 续期
	if err := b.accountService.Renew(ctx, acc.ID, days); err != nil {
		b.stateMachine.ClearState(currentUser.TelegramID)
		b.reply(msg.Chat.ID, fmt.Sprintf("❌ 续期失败: %v", err))
		return
	}

	// 清除状态
	b.stateMachine.ClearState(currentUser.TelegramID)

	// 重新获取更新后的账号信息
	acc, _ = b.accountService.Get(ctx, acc.ID)
	expireInfo := timeutil.FormatExpireTime(acc.ExpireAt)

	text := fmt.Sprintf(`✅ <b>续期成功！</b>

账号 <b>%s</b> 已续期 %d 天
新的到期时间: %s`,
		acc.Username,
		days,
		expireInfo,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 返回账号详情", CallbackAccountInfo+":"+uintToStr(acc.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 我的账号", CallbackMyAccounts+":1"),
		),
	)

	replyMsg := tgbotapi.NewMessage(msg.Chat.ID, text)
	replyMsg.ParseMode = "HTML"
	replyMsg.ReplyMarkup = keyboard

	if _, err := b.api.Send(replyMsg); err != nil {
		b.reply(msg.Chat.ID, text)
	}
}

// isValidUsername 验证用户名格式
func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	match, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	return match
}
