// Package bot 提供 Telegram Bot 功能
package bot

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"emby-telegram/internal/account"
	"emby-telegram/internal/emby"
	"emby-telegram/internal/logger"
	"emby-telegram/internal/user"
)

// Bot Telegram Bot 实例
type Bot struct {
	api            *tgbotapi.BotAPI
	accountService *account.Service
	userService    *user.Service
	embyClient     *emby.Client
	adminIDs       map[int64]bool
	handlers       map[string]CommandHandler
	stateMachine   *StateMachine // 状态机
}

// CommandHandler 命令处理函数类型
type CommandHandler func(ctx context.Context, msg *tgbotapi.Message, args []string) (string, error)

// New 创建 Bot 实例
func New(token string, adminIDs []int64, accountSvc *account.Service, userSvc *user.Service, embyClient *emby.Client) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("create bot api: %w", err)
	}

	// 创建管理员映射
	admins := make(map[int64]bool, len(adminIDs))
	for _, id := range adminIDs {
		admins[id] = true
	}

	b := &Bot{
		api:            api,
		accountService: accountSvc,
		userService:    userSvc,
		embyClient:     embyClient,
		adminIDs:       admins,
		handlers:       make(map[string]CommandHandler),
		stateMachine:   NewStateMachine(),
	}

	// 注册命令处理器
	b.registerHandlers()

	// 设置 Bot 命令菜单
	if err := b.setupBotCommands(); err != nil {
		logger.Warnf("failed to setup bot commands: %v", err)
	}

	logger.Infof("bot authorized: @%s", api.Self.UserName)
	return b, nil
}

// Start 启动 Bot
func (b *Bot) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	logger.Info("bot listening for messages")

	for {
		select {
		case <-ctx.Done():
			logger.Info("bot received stop signal")
			return ctx.Err()
		case update := <-updates:
			// 处理回调查询（按钮点击）
			if update.CallbackQuery != nil {
				go b.handleCallbackQuery(ctx, update.CallbackQuery)
				continue
			}

			// 处理消息
			if update.Message != nil {
				go b.handleUpdate(ctx, update.Message)
			}
		}
	}
}

// Stop 停止 Bot
func (b *Bot) Stop() {
	b.api.StopReceivingUpdates()
	b.stateMachine.Stop()
	logger.Info("bot stopped receiving updates")
}

// handleUpdate 处理消息更新
func (b *Bot) handleUpdate(ctx context.Context, msg *tgbotapi.Message) {
	currentUser, err := b.userService.GetOrCreate(ctx, msg.From)
	if err != nil {
		logger.Errorf("failed to get or create user: %v", err)
		if isGroupChat(msg) {
			b.replyWithAutoDelete(msg.Chat.ID, "系统错误，请稍后再试", msg.MessageID)
		} else {
			b.reply(msg.Chat.ID, "系统错误，请稍后再试")
		}
		return
	}

	if !currentUser.CanAccess() {
		if isGroupChat(msg) {
			b.replyWithAutoDelete(msg.Chat.ID, "您已被封禁，无法使用此 Bot", msg.MessageID)
		} else {
			b.reply(msg.Chat.ID, "您已被封禁，无法使用此 Bot")
		}
		return
	}

	// 处理命令
	if msg.IsCommand() {
		b.handleCommand(ctx, msg, currentUser)
		return
	}

	// 检查用户是否处于对话状态
	state, stateData := b.stateMachine.GetState(currentUser.TelegramID)
	if state != StateIdle {
		b.handleStateInput(ctx, msg, currentUser, state, stateData)
		return
	}

	// 处理 Reply Keyboard 按钮点击
	if b.handleReplyKeyboardButton(ctx, msg, currentUser) {
		return
	}

	// 处理普通消息
	if isGroupChat(msg) {
		b.replyWithAutoDelete(msg.Chat.ID, "请使用命令进行操作，使用 /help 查看帮助", msg.MessageID)
	} else {
		b.reply(msg.Chat.ID, "请点击下方按钮进行操作，或使用 /start 查看主菜单")
	}
}

// handleCommand 处理命令
func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message, currentUser *user.User) {
	cmd := msg.Command()
	args := parseArgs(msg.CommandArguments())

	logger.Infof("user executed command: %s, command: %s", currentUser.DisplayName(), cmd)

	handler, ok := b.handlers[cmd]
	if !ok {
		if isGroupChat(msg) {
			b.replyWithAutoDelete(msg.Chat.ID, "未知命令，使用 /help 查看帮助", msg.MessageID)
		} else {
			b.reply(msg.Chat.ID, "未知命令，使用 /help 查看帮助")
		}
		return
	}

	if isGroupChat(msg) && !groupAllowedCommands[cmd] {
		b.replyWithAutoDelete(msg.Chat.ID, "❌ 此命令仅支持私聊使用\n\n请点击 Bot 头像进入私聊", msg.MessageID)
		return
	}

	reply, err := handler(ctx, msg, args)
	if err != nil {
		logger.Errorf("command execution failed: %s, error: %v", cmd, err)
		errMsg := fmt.Sprintf("❌ 错误: %v", err)
		if isGroupChat(msg) {
			b.replyWithAutoDelete(msg.Chat.ID, errMsg, msg.MessageID)
		} else {
			b.reply(msg.Chat.ID, errMsg)
		}
		return
	}

	if reply != "" {
		if isGroupChat(msg) {
			b.replyWithAutoDelete(msg.Chat.ID, reply, msg.MessageID)
		} else {
			b.reply(msg.Chat.ID, reply)
		}
	}
}

// reply 回复消息
func (b *Bot) reply(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	if _, err := b.api.Send(msg); err != nil {
		logger.Errorf("failed to send message: %v", err)
	}
}

// replyWithAutoDelete 回复消息并在群组中自动删除
func (b *Bot) replyWithAutoDelete(chatID int64, text string, userMsgID int) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	if chatID < 0 {
		msg.ReplyMarkup = RemoveReplyKeyboard()
	}

	sentMsg, err := b.api.Send(msg)
	if err != nil {
		logger.Errorf("failed to send message: %v", err)
		return
	}

	if chatID < 0 {
		go func() {
			time.Sleep(30 * time.Second)

			deleteMsg := tgbotapi.NewDeleteMessage(chatID, sentMsg.MessageID)
			if _, err := b.api.Request(deleteMsg); err != nil {
				logger.Debugf("failed to delete bot message: %v", err)
			}

			if userMsgID != 0 {
				deleteUserMsg := tgbotapi.NewDeleteMessage(chatID, userMsgID)
				if _, err := b.api.Request(deleteUserMsg); err != nil {
					logger.Debugf("failed to delete user message: %v", err)
				}
			}
		}()
	}
}

// isAdmin 检查用户是否为管理员
func (b *Bot) isAdmin(telegramID int64) bool {
	return b.adminIDs[telegramID]
}

// requireAdmin 检查是否为管理员，如果不是则返回错误
func (b *Bot) requireAdmin(telegramID int64) error {
	if !b.isAdmin(telegramID) {
		return user.UnauthorizedError("此命令需要管理员权限")
	}
	return nil
}

// isPrivateChat 检查是否为私聊
func isPrivateChat(msg *tgbotapi.Message) bool {
	return msg.Chat.Type == "private"
}

// isGroupChat 检查是否为群组或超级群组
func isGroupChat(msg *tgbotapi.Message) bool {
	return msg.Chat.Type == "group" || msg.Chat.Type == "supergroup"
}

// handleReplyKeyboardButton 处理 Reply Keyboard 按钮点击
// 返回 true 表示已处理，false 表示不是按钮文本
func (b *Bot) handleReplyKeyboardButton(ctx context.Context, msg *tgbotapi.Message, currentUser *user.User) bool {
	var callbackData string

	switch msg.Text {
	case "📋 我的账号":
		callbackData = CallbackMyAccounts + ":1"
	case "➕ 创建账号":
		callbackData = CallbackCreateAccount
	case "❓ 帮助":
		callbackData = CallbackHelp
	case "🔑 管理员菜单":
		if !currentUser.IsAdmin() {
			if isGroupChat(msg) {
				b.replyWithAutoDelete(msg.Chat.ID, "❌ 您没有管理员权限", msg.MessageID)
			} else {
				b.reply(msg.Chat.ID, "❌ 您没有管理员权限")
			}
			return true
		}
		callbackData = CallbackAdminMenu
	default:
		return false
	}

	if isGroupChat(msg) {
		b.replyWithAutoDelete(msg.Chat.ID, "请在私聊中使用按钮功能", msg.MessageID)
		return true
	}

	query := &tgbotapi.CallbackQuery{
		ID:   fmt.Sprintf("reply_keyboard_%d", msg.MessageID),
		From: msg.From,
		Message: &tgbotapi.Message{
			MessageID: msg.MessageID,
			Chat:      msg.Chat,
			Text:      msg.Text,
		},
		Data: callbackData,
	}

	b.handleCallbackQuery(ctx, query)
	return true
}

// setupBotCommands 设置 Bot 命令菜单（显示在输入框的 / 按钮中）
func (b *Bot) setupBotCommands() error {
	// 群组命令（仅白名单命令）
	groupCommands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "开始使用 Bot",
		},
		{
			Command:     "help",
			Description: "查看帮助信息",
		},
		{
			Command:     "grant",
			Description: "授权用户创建账号（管理员）",
		},
		{
			Command:     "stats",
			Description: "查看系统统计（管理员）",
		},
		{
			Command:     "checkemby",
			Description: "检查 Emby 服务器状态",
		},
		{
			Command:     "playingstats",
			Description: "查看播放统计（管理员）",
		},
	}

	// 私聊命令（全部命令）
	privateCommands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "开始使用 Bot",
		},
		{
			Command:     "help",
			Description: "查看帮助信息",
		},
		{
			Command:     "myaccounts",
			Description: "查看我的账号列表",
		},
		{
			Command:     "create",
			Description: "创建新账号",
		},
		{
			Command:     "info",
			Description: "查看账号详情",
		},
		{
			Command:     "renew",
			Description: "续期账号",
		},
		{
			Command:     "changepassword",
			Description: "修改账号密码",
		},
		{
			Command:     "admin",
			Description: "管理员菜单（仅管理员）",
		},
	}

	// 为群组设置命令
	groupScope := tgbotapi.BotCommandScope{
		Type: "all_group_chats",
	}
	groupCfg := tgbotapi.SetMyCommandsConfig{
		Commands: groupCommands,
		Scope:    &groupScope,
	}
	if _, err := b.api.Request(groupCfg); err != nil {
		logger.Warnf("failed to set group commands: %v", err)
	}

	// 为私聊设置命令
	privateScope := tgbotapi.BotCommandScope{
		Type: "all_private_chats",
	}
	privateCfg := tgbotapi.SetMyCommandsConfig{
		Commands: privateCommands,
		Scope:    &privateScope,
	}
	if _, err := b.api.Request(privateCfg); err != nil {
		logger.Warnf("failed to set private commands: %v", err)
	}

	logger.Info("bot commands configured for different chat types")
	return nil
}
