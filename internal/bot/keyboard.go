// Package bot 按钮菜单定义
package bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Callback Data 格式常量
const (
	// 主菜单
	CallbackMainMenu = "menu:main"
	CallbackHelp     = "menu:help"

	// 账号列表
	CallbackMyAccounts = "accounts:list"     // accounts:list:page
	CallbackAccountInfo = "account:info"     // account:info:accountID
	CallbackAccountRenew = "account:renew"   // account:renew:accountID
	CallbackAccountPassword = "account:pwd"    // account:pwd:accountID
	CallbackAccountDelete = "account:del"      // account:del:accountID
	CallbackAccountSync = "account:sync"       // account:sync:accountID
	CallbackAccountRating = "account:rating"   // account:rating:accountID

	// 创建账号
	CallbackCreateAccount = "create:start"

	// 管理员菜单
	CallbackAdminMenu = "admin:menu"
	CallbackAdminUsers = "admin:users"       // admin:users:page
	CallbackAdminUserDetail = "admin:user"   // admin:user:userID:page
	CallbackAdminAccounts = "admin:accounts" // admin:accounts:page
	CallbackAdminAccountDetail = "admin:account" // admin:account:accountID
	CallbackAdminAccountSuspend = "admin:suspend" // admin:suspend:accountID
	CallbackAdminAccountActivate = "admin:activate" // admin:activate:accountID
	CallbackAdminStats = "admin:stats"
	CallbackAdminEmby = "admin:emby"
	CallbackAdminPlayingStats = "admin:playing"
	CallbackAdminUpdatePolicies = "admin:updatepolicies"

	// 通用操作
	CallbackConfirm = "confirm" // confirm:action:param
	CallbackCancel  = "cancel"
	CallbackBack    = "back" // back:to:menu
)

// MainMenuKeyboard 主菜单键盘
func MainMenuKeyboard(isAdmin bool) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("📋 我的账号", CallbackMyAccounts+":1"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("➕ 创建新账号", CallbackCreateAccount),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("❓ 帮助", CallbackHelp),
		},
	}

	// 管理员额外按钮
	if isAdmin {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("🔑 管理员菜单", CallbackAdminMenu),
		})
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// AdminMenuKeyboard 管理员菜单键盘
func AdminMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👥 用户管理", CallbackAdminUsers+":1"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 账号管理", CallbackAdminAccounts+":1"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎬 Emby 管理", CallbackAdminEmby),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 系统统计", CallbackAdminStats),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ 返回主菜单", CallbackMainMenu),
		),
	)
}

// EmbyMenuKeyboard Emby 管理子菜单键盘
func EmbyMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 播放统计", CallbackAdminPlayingStats),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 批量更新策略", CallbackAdminUpdatePolicies),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ 返回管理员菜单", CallbackAdminMenu),
		),
	)
}

// AccountActionsKeyboard 单个账号操作键盘
func AccountActionsKeyboard(accountID uint, isAdmin bool) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("🔄 续期", CallbackAccountRenew+":"+uintToStr(accountID)),
			tgbotapi.NewInlineKeyboardButtonData("🔑 改密", CallbackAccountPassword+":"+uintToStr(accountID)),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("🔞 设置评级", CallbackAccountRating+":"+uintToStr(accountID)),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("🔄 同步状态", CallbackAccountSync+":"+uintToStr(accountID)),
		},
	}

	// 管理员可以删除账号
	if isAdmin {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("❌ 删除账号", CallbackAccountDelete+":"+uintToStr(accountID)),
		})
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("⬅️ 返回列表", CallbackMyAccounts+":1"),
	})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// RenewDaysKeyboard 续期天数选择键盘
func RenewDaysKeyboard(accountID uint) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("7天", CallbackConfirm+":renew:"+uintToStr(accountID)+":7"),
			tgbotapi.NewInlineKeyboardButtonData("30天", CallbackConfirm+":renew:"+uintToStr(accountID)+":30"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("90天", CallbackConfirm+":renew:"+uintToStr(accountID)+":90"),
			tgbotapi.NewInlineKeyboardButtonData("365天", CallbackConfirm+":renew:"+uintToStr(accountID)+":365"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ 取消", CallbackAccountInfo+":"+uintToStr(accountID)),
		),
	)
}

// ParentalRatingKeyboard 家长控制评级选择键盘
// Emby 实际评级映射:
// 3=TV-Y7, 4=TV-Y7-FV, 5=TV-PG, 7=PG-13, 8=TV-14, 9=TV-MA, 10=NC-17, 15=AO
func ParentalRatingKeyboard(accountID uint) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("TV-Y7(3)", CallbackConfirm+":rating:"+uintToStr(accountID)+":3"),
			tgbotapi.NewInlineKeyboardButtonData("TV-Y7-FV(4)", CallbackConfirm+":rating:"+uintToStr(accountID)+":4"),
			tgbotapi.NewInlineKeyboardButtonData("TV-PG(5)", CallbackConfirm+":rating:"+uintToStr(accountID)+":5"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("PG-13(7)", CallbackConfirm+":rating:"+uintToStr(accountID)+":7"),
			tgbotapi.NewInlineKeyboardButtonData("TV-14(8)", CallbackConfirm+":rating:"+uintToStr(accountID)+":8"),
			tgbotapi.NewInlineKeyboardButtonData("TV-MA(9)", CallbackConfirm+":rating:"+uintToStr(accountID)+":9"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("NC-17(10)", CallbackConfirm+":rating:"+uintToStr(accountID)+":10"),
			tgbotapi.NewInlineKeyboardButtonData("AO(15)", CallbackConfirm+":rating:"+uintToStr(accountID)+":15"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ 取消", CallbackAccountInfo+":"+uintToStr(accountID)),
		),
	)
}

// ConfirmKeyboard 确认操作键盘
func ConfirmKeyboard(action string, param string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ 确认", CallbackConfirm+":"+action+":"+param),
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消", CallbackCancel),
		),
	)
}

// PaginationKeyboard 分页键盘
func PaginationKeyboard(prefix string, currentPage, totalPages int, backCallback string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// 分页按钮
	if totalPages > 1 {
		var pageRow []tgbotapi.InlineKeyboardButton

		if currentPage > 1 {
			pageRow = append(pageRow, tgbotapi.NewInlineKeyboardButtonData("⬅️ 上一页", prefix+":"+intToStr(currentPage-1)))
		}

		pageRow = append(pageRow, tgbotapi.NewInlineKeyboardButtonData(
			intToStr(currentPage)+"/"+intToStr(totalPages),
			"page:current", // 不可点击
		))

		if currentPage < totalPages {
			pageRow = append(pageRow, tgbotapi.NewInlineKeyboardButtonData("➡️ 下一页", prefix+":"+intToStr(currentPage+1)))
		}

		rows = append(rows, pageRow)
	}

	// 返回按钮
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("⬅️ 返回", backCallback),
	})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// BackButton 返回按钮
func BackButton(callback string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ 返回", callback),
		),
	)
}

// 辅助函数：uint 转字符串
func uintToStr(n uint) string {
	return intToStr(int(n))
}

// 辅助函数：int 转字符串
func intToStr(n int) string {
	return fmt.Sprintf("%d", n)
}

// MainReplyKeyboard 主菜单回复键盘（显示在输入框下方）
func MainReplyKeyboard(isAdmin bool) tgbotapi.ReplyKeyboardMarkup {
	rows := [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.NewKeyboardButton("📋 我的账号"),
			tgbotapi.NewKeyboardButton("➕ 创建账号"),
		},
		{
			tgbotapi.NewKeyboardButton("❓ 帮助"),
		},
	}

	if isAdmin {
		rows = append(rows, []tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton("🔑 管理员菜单"),
		})
	}

	return tgbotapi.NewReplyKeyboard(rows...)
}

// RemoveReplyKeyboard 移除回复键盘
func RemoveReplyKeyboard() tgbotapi.ReplyKeyboardRemove {
	return tgbotapi.NewRemoveKeyboard(true)
}

// AdminAccountActionsKeyboard 管理员账号操作键盘
func AdminAccountActionsKeyboard(accountID uint, status string, page int) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("🔄 续期", CallbackAccountRenew+":"+uintToStr(accountID)),
			tgbotapi.NewInlineKeyboardButtonData("🔑 改密", CallbackAccountPassword+":"+uintToStr(accountID)),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("🔞 设置评级", CallbackAccountRating+":"+uintToStr(accountID)),
		},
	}

	if status == "active" {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("⏸️ 停用", CallbackAdminAccountSuspend+":"+uintToStr(accountID)),
		})
	} else if status == "suspended" {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("✅ 激活", CallbackAdminAccountActivate+":"+uintToStr(accountID)),
		})
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ 删除账号", CallbackAccountDelete+":"+uintToStr(accountID)),
	})

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("⬅️ 返回列表", CallbackAdminAccounts+":"+intToStr(page)),
	})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
