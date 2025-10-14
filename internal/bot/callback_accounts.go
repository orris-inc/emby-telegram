// Package bot 账号回调处理
package bot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"emby-telegram/internal/user"
	"emby-telegram/pkg/timeutil"
)

// handleAccountsCallback 处理账号列表相关回调
func (b *Bot) handleAccountsCallback(ctx context.Context, query *tgbotapi.CallbackQuery, parts []string, currentUser *user.User) CallbackResponse {
	if len(parts) < 2 {
		return CallbackResponse{Answer: "无效的操作", ShowAlert: true}
	}

	subAction := parts[1]

	switch subAction {
	case "list":
		page := 1
		if len(parts) >= 3 {
			page = strToInt(parts[2])
		}
		return b.showMyAccountsList(ctx, currentUser, page)
	default:
		return CallbackResponse{Answer: "未知操作", ShowAlert: true}
	}
}

// handleAccountCallback 处理单个账号操作回调
func (b *Bot) handleAccountCallback(ctx context.Context, query *tgbotapi.CallbackQuery, parts []string, currentUser *user.User) CallbackResponse {
	if len(parts) < 3 {
		return CallbackResponse{Answer: "无效的操作", ShowAlert: true}
	}

	subAction := parts[1]
	accountID := strToUint(parts[2])

	switch subAction {
	case "info":
		return b.showAccountInfo(ctx, currentUser, accountID)
	case "renew":
		return b.showRenewOptions(ctx, currentUser, accountID)
	case "pwd":
		return b.startChangePassword(ctx, currentUser, accountID)
	case "sync":
		return b.syncAccountStatus(ctx, currentUser, accountID)
	case "rating":
		return b.showRatingOptions(ctx, currentUser, accountID)
	case "del":
		return b.confirmDeleteAccount(ctx, currentUser, accountID)
	default:
		return CallbackResponse{Answer: "未知操作", ShowAlert: true}
	}
}

// showMyAccountsList 显示我的账号列表
func (b *Bot) showMyAccountsList(ctx context.Context, currentUser *user.User, page int) CallbackResponse {
	accounts, err := b.accountService.ListByUser(ctx, currentUser.ID)
	if err != nil {
		return CallbackResponse{
			Answer:    "获取账号列表失败",
			ShowAlert: true,
		}
	}

	if len(accounts) == 0 {
		text := `📋 <b>您的账号列表</b>

您还没有创建任何账号。

点击下方按钮创建新账号：`

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ 创建新账号", CallbackCreateAccount),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⬅️ 返回主菜单", CallbackMainMenu),
			),
		)

		return CallbackResponse{
			EditText:   text,
			EditMarkup: &keyboard,
		}
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📋 <b>您的账号列表</b> (共 %d 个)\n\n", len(accounts)))

	// 构建账号列表（每个账号显示为按钮）
	var rows [][]tgbotapi.InlineKeyboardButton

	for i, acc := range accounts {
		status := getStatusEmoji(string(acc.Status))
		expireInfo := timeutil.FormatExpireTime(acc.ExpireAt)

		builder.WriteString(fmt.Sprintf("%d. <b>%s</b> %s\n", i+1, acc.Username, status))
		builder.WriteString(fmt.Sprintf("   到期时间: %s\n", expireInfo))
		builder.WriteString(fmt.Sprintf("   最大设备数: %d\n\n", acc.MaxDevices))

		// 每个账号一个操作按钮行
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("📝 %s", acc.Username),
				CallbackAccountInfo+":"+uintToStr(acc.ID),
			),
		))
	}

	builder.WriteString("💡 点击账号名称查看详细信息")

	// 添加返回按钮
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ 返回主菜单", CallbackMainMenu),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return CallbackResponse{
		EditText:   builder.String(),
		EditMarkup: &keyboard,
	}
}

// showAccountInfo 显示账号详情
func (b *Bot) showAccountInfo(ctx context.Context, currentUser *user.User, accountID uint) CallbackResponse {
	acc, err := b.accountService.Get(ctx, accountID)
	if err != nil {
		return CallbackResponse{
			Answer:    "获取账号信息失败",
			ShowAlert: true,
		}
	}

	// 检查所有权
	if err := b.accountService.CheckOwnership(ctx, acc.ID, currentUser.ID); err != nil {
		return CallbackResponse{
			Answer:    "您没有权限查看此账号",
			ShowAlert: true,
		}
	}

	status := getStatusEmoji(string(acc.Status))
	expireInfo := timeutil.FormatExpireTime(acc.ExpireAt)
	createdAt := timeutil.FormatDateTime(acc.CreatedAt)

	syncStatus := "❌ 未同步"
	if acc.EmbyUserID != "" {
		syncStatus = fmt.Sprintf("✅ 已同步 (ID: <code>%s</code>)", acc.EmbyUserID)
	}

	text := fmt.Sprintf(`📝 <b>账号详情</b>

<b>用户名:</b> <code>%s</code>
<b>状态:</b> %s %s
<b>有效期:</b> %s
<b>最大设备数:</b> %d
<b>创建时间:</b> %s
<b>Emby 同步:</b> %s

请选择要执行的操作：`,
		acc.Username,
		status,
		acc.Status,
		expireInfo,
		acc.MaxDevices,
		createdAt,
		syncStatus,
	)

	keyboard := AccountActionsKeyboard(acc.ID, currentUser.IsAdmin())

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// showRenewOptions 显示续期选项
func (b *Bot) showRenewOptions(ctx context.Context, currentUser *user.User, accountID uint) CallbackResponse {
	acc, err := b.accountService.Get(ctx, accountID)
	if err != nil {
		return CallbackResponse{
			Answer:    "获取账号信息失败",
			ShowAlert: true,
		}
	}

	// 检查所有权
	if err := b.accountService.CheckOwnership(ctx, acc.ID, currentUser.ID); err != nil {
		return CallbackResponse{
			Answer:    "您没有权限操作此账号",
			ShowAlert: true,
		}
	}

	text := fmt.Sprintf(`🔄 <b>续期账号: %s</b>

当前到期时间: %s

请选择续期天数：`,
		acc.Username,
		timeutil.FormatExpireTime(acc.ExpireAt),
	)

	keyboard := RenewDaysKeyboard(acc.ID)

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// syncAccountStatus 同步账号状态
func (b *Bot) syncAccountStatus(ctx context.Context, currentUser *user.User, accountID uint) CallbackResponse {
	acc, err := b.accountService.Get(ctx, accountID)
	if err != nil {
		return CallbackResponse{
			Answer:    "获取账号信息失败",
			ShowAlert: true,
		}
	}

	// 检查所有权
	if err := b.accountService.CheckOwnership(ctx, acc.ID, currentUser.ID); err != nil {
		return CallbackResponse{
			Answer:    "您没有权限操作此账号",
			ShowAlert: true,
		}
	}

	// 账号在创建和修改时会自动同步，这里只是刷新显示
	// 重新显示账号详情
	response := b.showAccountInfo(ctx, currentUser, accountID)
	response.Answer = "✅ 已刷新账号信息"
	return response
}

// confirmDeleteAccount 确认删除账号
func (b *Bot) confirmDeleteAccount(ctx context.Context, currentUser *user.User, accountID uint) CallbackResponse {
	// 只有管理员可以删除
	if !currentUser.IsAdmin() {
		return CallbackResponse{
			Answer:    "只有管理员可以删除账号",
			ShowAlert: true,
		}
	}

	acc, err := b.accountService.Get(ctx, accountID)
	if err != nil {
		return CallbackResponse{
			Answer:    "获取账号信息失败",
			ShowAlert: true,
		}
	}

	text := fmt.Sprintf(`⚠️ <b>确认删除账号</b>

您确定要删除账号 <b>%s</b> 吗？

此操作<b>不可撤销</b>，账号将从数据库和 Emby 服务器上删除。`,
		acc.Username,
	)

	keyboard := ConfirmKeyboard("delete", uintToStr(accountID))

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// startChangePassword 开始修改密码流程
func (b *Bot) startChangePassword(ctx context.Context, currentUser *user.User, accountID uint) CallbackResponse {
	acc, err := b.accountService.Get(ctx, accountID)
	if err != nil {
		return CallbackResponse{
			Answer:    "获取账号信息失败",
			ShowAlert: true,
		}
	}

	// 检查所有权
	if err := b.accountService.CheckOwnership(ctx, acc.ID, currentUser.ID); err != nil {
		return CallbackResponse{
			Answer:    "您没有权限操作此账号",
			ShowAlert: true,
		}
	}

	// 设置用户状态为等待输入密码
	b.stateMachine.SetState(currentUser.TelegramID, StateWaitingPassword, map[string]interface{}{
		"account_id": accountID,
	})

	text := fmt.Sprintf(`🔑 <b>修改密码: %s</b>

请输入新密码（或输入 /cancel 取消）：

<b>密码要求：</b>
• 长度至少 6 个字符
• 建议包含字母、数字和特殊字符`,
		acc.Username,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消", CallbackAccountInfo+":"+uintToStr(accountID)),
		),
	)

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// showRatingOptions 显示评级选项
func (b *Bot) showRatingOptions(ctx context.Context, currentUser *user.User, accountID uint) CallbackResponse {
	acc, err := b.accountService.Get(ctx, accountID)
	if err != nil {
		return CallbackResponse{
			Answer:    "获取账号信息失败",
			ShowAlert: true,
		}
	}

	// 检查所有权
	if err := b.accountService.CheckOwnership(ctx, acc.ID, currentUser.ID); err != nil {
		return CallbackResponse{
			Answer:    "您没有权限操作此账号",
			ShowAlert: true,
		}
	}

	// 获取当前 Emby 用户策略
	var currentRating string = "未设置"
	if acc.EmbyUserID != "" {
		policy, err := b.embyClient.GetUserPolicy(ctx, acc.EmbyUserID)
		if err == nil && policy.MaxParentalRating > 0 {
			currentRating = fmt.Sprintf("%d", policy.MaxParentalRating)
		}
	}

	text := fmt.Sprintf(`🔞 <b>设置家长控制评级: %s</b>

<b>当前评级：</b>%s

<b>评级说明：</b>
• <b>TV-Y7(3)</b> - 适合7岁及以上儿童
• <b>TV-Y7-FV(4)</b> - 7岁+含幻想暴力
• <b>TV-PG(5)</b> - 建议家长指导观看
• <b>PG-13(7)</b> - 建议13岁以上观看
• <b>TV-14(8)</b> - 适合14岁及以上（推荐）
• <b>TV-MA(9)</b> - 仅限成人观看
• <b>NC-17(10)</b> - 17岁以下禁止
• <b>AO(15)</b> - 仅限成人

请选择评级等级：`,
		acc.Username,
		currentRating,
	)

	keyboard := ParentalRatingKeyboard(acc.ID)

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}
