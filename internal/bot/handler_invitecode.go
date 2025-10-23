package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"emby-telegram/internal/invitecode"
	"emby-telegram/pkg/timeutil"
)

func (b *Bot) handleGenerateCode(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	if !hasArg(args, 1) {
		return `❌ 参数不足

<b>使用方法:</b>
<code>/generatecode &lt;使用次数&gt; [有效期天数] [备注]</code>

<b>示例:</b>
<code>/generatecode 10 30 "推广活动A"</code> - 10次使用，30天有效
<code>/generatecode 1</code> - 单次使用，永久有效
<code>/generatecode -1 90</code> - 无限次使用，90天有效

<b>说明:</b>
• 使用次数: -1 表示无限次
• 有效期: 不填或0表示永久有效
• 备注: 可选，便于管理`, nil
	}

	maxUsesStr := getArg(args, 0)
	maxUses, err := strconv.Atoi(maxUsesStr)
	if err != nil || (maxUses != -1 && maxUses <= 0) {
		return "❌ 使用次数必须是 -1（无限）或正整数", nil
	}

	expireDays := 0
	if hasArg(args, 2) {
		expireDaysStr := getArg(args, 1)
		if expireDaysStr != "" {
			expireDays, err = strconv.Atoi(expireDaysStr)
			if err != nil || expireDays < 0 {
				return "❌ 有效期必须是非负整数", nil
			}
		}
	}

	description := ""
	if hasArg(args, 3) {
		description = strings.Join(args[2:], " ")
		description = strings.Trim(description, "\"' ")
	}

	code, err := b.inviteCodeService.Generate(ctx, maxUses, expireDays, description, msg.From.ID)
	if err != nil {
		return "", fmt.Errorf("生成邀请码失败: %w", err)
	}

	maxUsesText := "无限次"
	if maxUses != -1 {
		maxUsesText = fmt.Sprintf("%d 次", maxUses)
	}

	expireText := "永久有效"
	if expireDays > 0 {
		expireText = timeutil.FormatDateTime(*code.ExpireAt)
	}

	result := fmt.Sprintf(`✅ <b>邀请码已生成</b>

<b>邀请码:</b> <code>%s</code>
<b>使用次数:</b> 0/%s
<b>有效期:</b> %s`,
		code.Code,
		maxUsesText,
		expireText,
	)

	if description != "" {
		result += fmt.Sprintf("\n<b>备注:</b> %s", description)
	}

	result += "\n\n💡 每个邀请码激活后给用户 <b>1 个账号配额</b>"

	return result, nil
}

func (b *Bot) handleListCodes(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	page := 1
	if hasArg(args, 1) {
		if p, err := strconv.Atoi(getArg(args, 0)); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	offset := (page - 1) * limit

	codes, err := b.inviteCodeService.List(ctx, offset, limit)
	if err != nil {
		return "", fmt.Errorf("获取邀请码列表失败: %w", err)
	}

	totalCount, _ := b.inviteCodeService.Count(ctx)

	if len(codes) == 0 {
		return "没有找到邀请码\n\n使用 <code>/generatecode</code> 生成新邀请码", nil
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("🎟️ <b>邀请码列表</b> (第 %d 页，共 %d 个)\n\n", page, totalCount))

	for i, code := range codes {
		statusEmoji := "✅"
		statusText := "有效"
		if code.Status == invitecode.StatusRevoked {
			statusEmoji = "🚫"
			statusText = "已撤销"
		} else if code.IsExpired() {
			statusEmoji = "⏰"
			statusText = "已过期"
		} else if code.IsExhausted() {
			statusEmoji = "📛"
			statusText = "已用完"
		}

		maxUsesText := "∞"
		if code.MaxUses != -1 {
			maxUsesText = fmt.Sprintf("%d", code.MaxUses)
		}

		builder.WriteString(fmt.Sprintf("%d. <code>%s</code> %s %s\n",
			offset+i+1,
			code.Code,
			statusEmoji,
			statusText,
		))

		builder.WriteString(fmt.Sprintf("   使用: %d/%s",
			code.CurrentUses,
			maxUsesText,
		))

		if code.ExpireAt != nil {
			builder.WriteString(fmt.Sprintf(" | 到期: %s", timeutil.FormatDate(*code.ExpireAt)))
		}

		if code.Description != "" {
			builder.WriteString(fmt.Sprintf("\n   备注: %s", code.Description))
		}

		builder.WriteString("\n\n")
	}

	totalPages := (int(totalCount) + limit - 1) / limit
	if totalPages > 1 {
		builder.WriteString(fmt.Sprintf("📄 第 %d/%d 页\n", page, totalPages))
		if page < totalPages {
			builder.WriteString(fmt.Sprintf("使用 <code>/listcodes %d</code> 查看下一页\n", page+1))
		}
	}

	builder.WriteString("\n💡 使用 <code>/codeinfo &lt;邀请码&gt;</code> 查看详情")

	return builder.String(), nil
}

func (b *Bot) handleCodeInfo(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	if !hasArg(args, 1) {
		return "❌ 请提供邀请码\n\n使用方法: <code>/codeinfo &lt;邀请码&gt;</code>", nil
	}

	codeStr := getArg(args, 0)

	codeWithUsage, err := b.inviteCodeService.GetWithUsage(ctx, codeStr)
	if err != nil {
		if errors.Is(err, invitecode.ErrNotFound) {
			return fmt.Sprintf("❌ 邀请码 <code>%s</code> 不存在", codeStr), nil
		}
		return "", fmt.Errorf("获取邀请码信息失败: %w", err)
	}

	code := codeWithUsage.InviteCode

	statusEmoji := "✅"
	statusText := "有效"
	if code.Status == invitecode.StatusRevoked {
		statusEmoji = "🚫"
		statusText = "已撤销"
	} else if code.IsExpired() {
		statusEmoji = "⏰"
		statusText = "已过期"
	} else if code.IsExhausted() {
		statusEmoji = "📛"
		statusText = "已用完"
	}

	maxUsesText := "无限次"
	if code.MaxUses != -1 {
		maxUsesText = fmt.Sprintf("%d 次", code.MaxUses)
	}

	expireText := "永久有效"
	if code.ExpireAt != nil {
		expireText = timeutil.FormatDateTime(*code.ExpireAt)
	}

	result := fmt.Sprintf(`🎟️ <b>邀请码详情</b>

<b>邀请码:</b> <code>%s</code>
<b>状态:</b> %s %s
<b>使用次数:</b> %d/%s
<b>有效期:</b> %s
<b>创建时间:</b> %s`,
		code.Code,
		statusEmoji,
		statusText,
		code.CurrentUses,
		maxUsesText,
		expireText,
		timeutil.FormatDateTime(code.CreatedAt),
	)

	if code.Description != "" {
		result += fmt.Sprintf("\n<b>备注:</b> %s", code.Description)
	}

	if len(codeWithUsage.UsageRecords) > 0 {
		result += "\n\n<b>使用记录:</b>"
		for i, usage := range codeWithUsage.UsageRecords {
			if i >= 10 {
				result += fmt.Sprintf("\n... 还有 %d 条记录", len(codeWithUsage.UsageRecords)-10)
				break
			}
			result += fmt.Sprintf("\n• 用户ID: %d | %s",
				usage.UserID,
				timeutil.FormatDateTime(usage.UsedAt),
			)
		}
	} else {
		result += "\n\n<b>使用记录:</b> 暂无"
	}

	return result, nil
}

func (b *Bot) handleRevokeCode(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error) {
	if err := b.requireAdmin(msg.From.ID); err != nil {
		return "❌ 此命令需要管理员权限", nil
	}

	if !hasArg(args, 1) {
		return "❌ 请提供邀请码\n\n使用方法: <code>/revokecode &lt;邀请码&gt;</code>", nil
	}

	codeStr := getArg(args, 0)

	if err := b.inviteCodeService.Revoke(ctx, codeStr); err != nil {
		if errors.Is(err, invitecode.ErrNotFound) {
			return fmt.Sprintf("❌ 邀请码 <code>%s</code> 不存在", codeStr), nil
		}
		return "", fmt.Errorf("撤销邀请码失败: %w", err)
	}

	return fmt.Sprintf("✅ 已撤销邀请码 <code>%s</code>", codeStr), nil
}
