// Package bot 命令注册和解析
package bot

import (
	"strings"
)

// groupAllowedCommands 群组中允许执行的命令白名单
var groupAllowedCommands = map[string]bool{
	"grant":        true,
	"checkemby":    true,
	"stats":        true,
	"playingstats": true,
	"help":         true,
	"start":        true,
	"generatecode": true,
	"listcodes":    true,
}

// registerHandlers 注册所有命令处理器
func (b *Bot) registerHandlers() {
	// 基础命令
	b.handlers["start"] = b.handleStart
	b.handlers["help"] = b.handleHelp

	// 用户命令
	b.handlers["myaccounts"] = b.handleMyAccounts
	b.handlers["create"] = b.handleCreateAccount
	b.handlers["info"] = b.handleAccountInfo
	b.handlers["renew"] = b.handleRenewAccount
	b.handlers["changepassword"] = b.handleChangePassword
	b.handlers["quota"] = b.handleQuota

	// 管理员命令
	b.handlers["admin"] = b.handleAdmin
	b.handlers["grant"] = b.handleGrant
	b.handlers["users"] = b.handleListUsers
	b.handlers["accounts"] = b.handleListAccounts
	b.handlers["deleteaccount"] = b.handleDeleteAccount
	b.handlers["suspend"] = b.handleSuspendAccountCmd
	b.handlers["activate"] = b.handleActivateAccountCmd
	b.handlers["setrole"] = b.handleSetRole
	b.handlers["blockuser"] = b.handleBlockUser
	b.handlers["unblockuser"] = b.handleUnblockUser
	b.handlers["stats"] = b.handleStats
	b.handlers["playingstats"] = b.handlePlayingStats
	b.handlers["updatepolicies"] = b.handleUpdatePolicies

	// Emby 管理命令
	b.handlers["checkemby"] = b.handleCheckEmby
	b.handlers["syncstatus"] = b.handleSyncStatus
	b.handlers["syncaccount"] = b.handleSyncAccount
	b.handlers["embyusers"] = b.handleListEmbyUsers
	b.handlers["setdevicelimit"] = b.handleSetDeviceLimit

	// 邀请码管理命令
	b.handlers["generatecode"] = b.handleGenerateCode
	b.handlers["listcodes"] = b.handleListCodes
	b.handlers["codeinfo"] = b.handleCodeInfo
	b.handlers["revokecode"] = b.handleRevokeCode
}

// parseArgs 解析命令参数
func parseArgs(argsString string) []string {
	if argsString == "" {
		return []string{}
	}

	parts := strings.Fields(argsString)
	return parts
}

// getArg 安全获取参数
func getArg(args []string, index int) string {
	if index < len(args) {
		return args[index]
	}
	return ""
}

// hasArg 检查是否有足够的参数
func hasArg(args []string, count int) bool {
	return len(args) >= count
}
