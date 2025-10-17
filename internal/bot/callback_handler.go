// Package bot 按钮回调处理器
package bot

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"emby-telegram/internal/logger"
)

// handleCallbackQuery 处理按钮回调
func (b *Bot) handleCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) {
	if query.Message != nil && isGroupChatFromMessage(query.Message) {
		b.answerCallback(query.ID, "请在私聊中使用按钮功能", true)
		return
	}

	currentUser, err := b.userService.GetOrCreate(ctx, query.From)
	if err != nil {
		logger.Errorf("failed to get or create user: %v", err)
		b.answerCallback(query.ID, "系统错误，请稍后再试", true)
		return
	}

	if !currentUser.CanAccess() {
		b.answerCallback(query.ID, "您已被封禁，无法使用此 Bot", true)
		return
	}

	// 解析 callback data
	parts := strings.Split(query.Data, ":")
	if len(parts) == 0 {
		b.answerCallback(query.ID, "无效的操作", true)
		return
	}

	action := parts[0]

	logger.Infof("user clicked button: %s, data: %s", currentUser.DisplayName(), query.Data)

	// 路由到对应的处理函数
	var response CallbackResponse
	switch action {
	case "menu":
		response = b.handleMenuCallback(ctx, query, parts, currentUser)
	case "accounts":
		response = b.handleAccountsCallback(ctx, query, parts, currentUser)
	case "account":
		response = b.handleAccountCallback(ctx, query, parts, currentUser)
	case "create":
		response = b.handleCreateCallback(ctx, query, parts, currentUser)
	case "admin":
		response = b.handleAdminCallback(ctx, query, parts, currentUser)
	case "confirm":
		response = b.handleConfirmCallback(ctx, query, parts, currentUser)
	case "cancel":
		response = CallbackResponse{Answer: "操作已取消"}
	case "back":
		response = b.handleBackCallback(ctx, query, parts, currentUser)
	default:
		response = CallbackResponse{Answer: "未知操作", ShowAlert: true}
	}

	// 发送响应
	b.sendCallbackResponse(query, response)
}

// CallbackResponse 回调响应结构
type CallbackResponse struct {
	Answer    string // Callback answer 提示文本
	ShowAlert bool   // 是否显示为弹窗
	EditText  string // 要编辑的消息文本
	EditMarkup *tgbotapi.InlineKeyboardMarkup // 要编辑的按钮
	NewMessage string // 发送新消息
	NewMarkup *tgbotapi.InlineKeyboardMarkup // 新消息的按钮
}

// sendCallbackResponse 发送回调响应
func (b *Bot) sendCallbackResponse(query *tgbotapi.CallbackQuery, response CallbackResponse) {
	if !strings.HasPrefix(query.ID, "reply_keyboard_") {
		b.answerCallback(query.ID, response.Answer, response.ShowAlert)
	}

	isReplyKeyboard := strings.HasPrefix(query.ID, "reply_keyboard_")
	isGroup := query.Message != nil && isGroupChatFromMessage(query.Message)

	if isGroup {
		response.EditMarkup = nil
		response.NewMarkup = nil
	}

	// 3. 编辑或发送消息
	if response.EditText != "" {
		// 如果来自 Reply Keyboard，发送新消息而不是编辑
		if isReplyKeyboard {
			msg := tgbotapi.NewMessage(query.Message.Chat.ID, response.EditText)
			msg.ParseMode = "HTML"
			if response.EditMarkup != nil {
				msg.ReplyMarkup = response.EditMarkup
			}
			if _, err := b.api.Send(msg); err != nil {
				logger.Errorf("failed to send message: %v", err)
			}
		} else {
			// 编辑原消息
			edit := tgbotapi.NewEditMessageText(
				query.Message.Chat.ID,
				query.Message.MessageID,
				response.EditText,
			)
			edit.ParseMode = "HTML"
			if response.EditMarkup != nil {
				edit.ReplyMarkup = response.EditMarkup
			}
			if _, err := b.api.Send(edit); err != nil {
				logger.Errorf("failed to edit message: %v", err)
			}
		}
	}

	// 4. 发送新消息
	if response.NewMessage != "" {
		msg := tgbotapi.NewMessage(query.Message.Chat.ID, response.NewMessage)
		msg.ParseMode = "HTML"
		if response.NewMarkup != nil {
			msg.ReplyMarkup = response.NewMarkup
		}
		if _, err := b.api.Send(msg); err != nil {
			logger.Errorf("failed to send message: %v", err)
		}
	}
}

// answerCallback 应答回调查询
func (b *Bot) answerCallback(callbackID, text string, showAlert bool) {
	callback := tgbotapi.NewCallback(callbackID, text)
	callback.ShowAlert = showAlert
	if _, err := b.api.Request(callback); err != nil {
		logger.Errorf("failed to answer callback: %v", err)
	}
}

// 辅助函数：获取参数
func getCallbackParam(parts []string, index int) string {
	if len(parts) > index {
		return parts[index]
	}
	return ""
}

// 辅助函数：字符串转 int
func strToInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		logger.Warnf("failed to convert string to int: %s, error: %v", s, err)
		return 0
	}
	return n
}

// 辅助函数：字符串转 uint
func strToUint(s string) uint {
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		logger.Warnf("failed to convert string to uint: %s, error: %v", s, err)
		return 0
	}
	return uint(n)
}

// isGroupChatFromMessage 检查消息是否来自群组
func isGroupChatFromMessage(msg *tgbotapi.Message) bool {
	if msg == nil {
		return false
	}
	return msg.Chat.Type == "group" || msg.Chat.Type == "supergroup"
}
