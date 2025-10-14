// Package bot 创建账号回调处理
package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"emby-telegram/internal/user"
)

// handleCreateCallback 处理创建账号回调
func (b *Bot) handleCreateCallback(ctx context.Context, query *tgbotapi.CallbackQuery, parts []string, currentUser *user.User) CallbackResponse {
	if len(parts) < 2 {
		return CallbackResponse{Answer: "无效的操作", ShowAlert: true}
	}

	subAction := parts[1]

	switch subAction {
	case "start":
		return b.startCreateAccount(ctx, currentUser)
	default:
		return CallbackResponse{Answer: "未知操作", ShowAlert: true}
	}
}

// startCreateAccount 开始创建账号流程
func (b *Bot) startCreateAccount(ctx context.Context, currentUser *user.User) CallbackResponse {
	// 设置状态为等待输入用户名
	b.stateMachine.SetState(currentUser.TelegramID, StateWaitingUsername, nil)

	text := `➕ <b>创建新账号</b>

请输入新账号的用户名：

<b>用户名要求：</b>
• 只能包含字母、数字和下划线
• 长度 3-20 个字符
• 不能与现有账号重复

输入 /cancel 取消创建`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消", CallbackMainMenu),
		),
	)

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// handleConfirmCallback 处理确认操作回调
func (b *Bot) handleConfirmCallback(ctx context.Context, query *tgbotapi.CallbackQuery, parts []string, currentUser *user.User) CallbackResponse {
	if len(parts) < 3 {
		return CallbackResponse{Answer: "无效的操作", ShowAlert: true}
	}

	action := parts[1]
	param := parts[2]

	switch action {
	case "renew":
		// confirm:renew:accountID:days
		if len(parts) < 4 {
			return CallbackResponse{Answer: "参数错误", ShowAlert: true}
		}
		accountID := strToUint(param)
		days := strToInt(parts[3])
		return b.executeRenew(ctx, currentUser, accountID, days)

	case "delete":
		// confirm:delete:accountID
		accountID := strToUint(param)
		return b.executeDelete(ctx, currentUser, accountID)

	case "rating":
		// confirm:rating:accountID:ratingValue
		if len(parts) < 4 {
			return CallbackResponse{Answer: "参数错误", ShowAlert: true}
		}
		accountID := strToUint(param)
		rating := strToInt(parts[3])
		return b.executeRatingUpdate(ctx, currentUser, accountID, rating)

	default:
		return CallbackResponse{Answer: "未知操作", ShowAlert: true}
	}
}

// executeRenew 执行续期
func (b *Bot) executeRenew(ctx context.Context, currentUser *user.User, accountID uint, days int) CallbackResponse {
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

	// 续期
	if err := b.accountService.Renew(ctx, acc.ID, days); err != nil {
		return CallbackResponse{
			Answer:    fmt.Sprintf("续期失败: %v", err),
			ShowAlert: true,
		}
	}

	// 显示成功消息并返回账号详情
	return CallbackResponse{
		Answer: fmt.Sprintf("✅ 已成功续期 %d 天", days),
		// 刷新账号详情页面
		EditText: "正在刷新...",
	}
}

// executeDelete 执行删除
func (b *Bot) executeDelete(ctx context.Context, currentUser *user.User, accountID uint) CallbackResponse {
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

	username := acc.Username

	// 删除账号
	if err := b.accountService.Delete(ctx, accountID); err != nil {
		return CallbackResponse{
			Answer:    fmt.Sprintf("删除失败: %v", err),
			ShowAlert: true,
		}
	}

	text := fmt.Sprintf(`✅ <b>账号已删除</b>

账号 <b>%s</b> 已成功删除。`, username)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 我的账号", CallbackMyAccounts+":1"),
			tgbotapi.NewInlineKeyboardButtonData("⬅️ 主菜单", CallbackMainMenu),
		),
	)

	return CallbackResponse{
		EditText:   text,
		EditMarkup: &keyboard,
	}
}

// executeRatingUpdate 执行评级更新
func (b *Bot) executeRatingUpdate(ctx context.Context, currentUser *user.User, accountID uint, rating int) CallbackResponse {
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

	// 检查账号是否已同步到 Emby
	if acc.EmbyUserID == "" {
		return CallbackResponse{
			Answer:    "账号尚未同步到 Emby，无法设置评级",
			ShowAlert: true,
		}
	}

	// 更新 Emby 用户策略中的家长控制评级
	policy, err := b.embyClient.GetUserPolicy(ctx, acc.EmbyUserID)
	if err != nil {
		return CallbackResponse{
			Answer:    fmt.Sprintf("获取用户策略失败: %v", err),
			ShowAlert: true,
		}
	}

	policy.MaxParentalRating = int32(rating)

	if err := b.embyClient.UpdateUserPolicy(ctx, acc.EmbyUserID, policy); err != nil {
		return CallbackResponse{
			Answer:    fmt.Sprintf("更新评级失败: %v", err),
			ShowAlert: true,
		}
	}

	// 返回账号详情页面，显示成功消息
	response := b.showAccountInfo(ctx, currentUser, accountID)
	response.Answer = fmt.Sprintf("✅ 评级已设置为 %d", rating)
	return response
}
