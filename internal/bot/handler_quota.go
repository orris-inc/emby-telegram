// Package bot 配额管理命令处理器
package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"emby-telegram/internal/user"
)

// handleGrant 处理 /grant 命令（管理员在群组中授权用户）
func (b *Bot) handleGrant(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if !isGroupChat(msg) {
		return "此命令请在管理群组中使用", nil
	}

	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	targetUser, quota, err := b.parseGrantArgs(ctx, msg, args)
	if err != nil {
		return "❌ " + err.Error(), nil
	}

	if err := b.userService.SetQuota(ctx, targetUser.ID, quota); err != nil {
		return "", fmt.Errorf("设置配额失败: %w", err)
	}

	count, _ := b.accountService.CountByUser(ctx, targetUser.ID)

	var text string
	if quota == 0 {
		text = fmt.Sprintf("⚠️ 已收回 %s 的创建权限", targetUser.DisplayName())
		if count > 0 {
			text += fmt.Sprintf("\n当前账号: %d (保留)", count)
		}
		text += "\n\n用户将无法创建新账号"
	} else {
		if count == 0 {
			text = fmt.Sprintf("✅ 已授权 %s 创建账号\n配额: %d 个",
				targetUser.DisplayName(), quota)
		} else {
			text = fmt.Sprintf("✅ 已将 %s 的账号配额调整为 %d",
				targetUser.DisplayName(), quota)
			text += fmt.Sprintf("\n当前账号: %d", count)

			if int(count) > quota {
				text += fmt.Sprintf("\n\n⚠️ 用户需要删除 %d 个账号才能继续使用", int(count)-quota)
			} else if int(count) < quota {
				text += fmt.Sprintf("\n可继续创建: %d 个", quota-int(count))
			}
		}
	}

	return text, nil
}

// parseGrantArgs 解析 grant 命令参数
func (b *Bot) parseGrantArgs(ctx context.Context, msg *tgbotapi.Message, args []string) (*user.User, int, error) {
	var targetUser *user.User
	var quotaStr string

	if msg.ReplyToMessage != nil {
		var err error
		targetUser, err = b.userService.GetByTelegramID(ctx, msg.ReplyToMessage.From.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("获取用户信息失败: %w", err)
		}

		if hasArg(args, 1) {
			quotaStr = getArg(args, 0)
		} else {
			quotaStr = "1"
		}
	} else if hasArg(args, 1) {
		userArg := getArg(args, 0)

		if hasArg(args, 2) {
			quotaStr = getArg(args, 1)
		} else {
			quotaStr = "1"
		}

		if strings.HasPrefix(userArg, "@") {
			username := strings.TrimPrefix(userArg, "@")
			var err error
			targetUser, err = b.userService.GetByUsername(ctx, username)
			if err != nil {
				return nil, 0, fmt.Errorf("用户 %s 不存在", userArg)
			}
		} else if telegramID, err := strconv.ParseInt(userArg, 10, 64); err == nil {
			targetUser, err = b.userService.GetByTelegramID(ctx, telegramID)
			if err != nil {
				return nil, 0, fmt.Errorf("用户 ID %d 不存在", telegramID)
			}
		} else {
			return nil, 0, errors.New("无效的用户标识\n\n支持的格式:\n• [回复消息] /grant [配额]\n• /grant @username [配额]\n• /grant <telegram_id> [配额]\n\n配额默认为 1")
		}
	} else {
		return nil, 0, errors.New("参数不足\n\n使用方法:\n• [回复目标消息] /grant [配额]\n• /grant @username [配额]\n• /grant <telegram_id> [配额]\n\n配额默认为 1")
	}

	quota, err := strconv.Atoi(quotaStr)
	if err != nil || quota < 0 {
		return nil, 0, errors.New("配额必须是非负整数")
	}

	return targetUser, quota, nil
}

// handleQuota 处理 /quota 命令（用户查询自己的配额）
func (b *Bot) handleQuota(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if !isPrivateChat(msg) {
		return "请在私聊中使用此命令", nil
	}

	user, err := b.userService.GetByTelegramID(ctx, msg.From.ID)
	if err != nil {
		return "", err
	}

	accounts, _ := b.accountService.ListByUser(ctx, user.ID)
	quota := user.AccountQuota
	count := len(accounts)

	if quota == 0 {
		return fmt.Sprintf(`📊 <b>账号授权状态:</b> ❌ 未授权

已创建账号: %d

💡 如需创建账号，请在管理群组联系管理员申请`, count), nil
	}

	remaining := quota - count
	status := "✅ 已授权"
	if count >= quota {
		status = "⚠️ 已满额"
	}

	return fmt.Sprintf(`📊 <b>账号授权状态:</b> %s

<b>账号配额:</b> %d
<b>已创建:</b> %d
<b>剩余:</b> %d

💡 配额不足？请在管理群组联系管理员`,
		status, quota, count, remaining), nil
}
