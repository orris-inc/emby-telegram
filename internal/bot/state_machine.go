// Package bot 用户状态机
package bot

import (
	"sync"
	"time"
)

// UserState 用户状态类型
type UserState string

const (
	StateIdle             UserState = "idle"              // 空闲状态
	StateWaitingUsername  UserState = "waiting_username"  // 等待输入用户名
	StateWaitingPassword  UserState = "waiting_password"  // 等待输入密码
	StateWaitingDays      UserState = "waiting_days"      // 等待输入天数
)

// StateData 状态数据
type StateData struct {
	State     UserState              // 当前状态
	Data      map[string]interface{} // 附加数据
	UpdatedAt time.Time              // 更新时间
}

// StateMachine 状态机
type StateMachine struct {
	mu     sync.RWMutex
	states map[int64]*StateData // userID -> StateData
	stopCh chan struct{}        // 停止信号
}

// NewStateMachine 创建状态机
func NewStateMachine() *StateMachine {
	sm := &StateMachine{
		states: make(map[int64]*StateData),
		stopCh: make(chan struct{}),
	}

	// 启动清理 goroutine
	go sm.cleanupExpiredStates()

	return sm
}

// Stop 停止状态机
func (sm *StateMachine) Stop() {
	close(sm.stopCh)
}

// SetState 设置用户状态
func (sm *StateMachine) SetState(userID int64, state UserState, data map[string]interface{}) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if data == nil {
		data = make(map[string]interface{})
	}

	sm.states[userID] = &StateData{
		State:     state,
		Data:      data,
		UpdatedAt: time.Now(),
	}
}

// GetState 获取用户状态
func (sm *StateMachine) GetState(userID int64) (UserState, map[string]interface{}) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if stateData, ok := sm.states[userID]; ok {
		// 检查是否过期（10分钟）
		if time.Since(stateData.UpdatedAt) > 10*time.Minute {
			return StateIdle, nil
		}
		return stateData.State, stateData.Data
	}

	return StateIdle, nil
}

// ClearState 清除用户状态
func (sm *StateMachine) ClearState(userID int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.states, userID)
}

// cleanupExpiredStates 清理过期状态
func (sm *StateMachine) cleanupExpiredStates() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sm.stopCh:
			return // 优雅退出
		case <-ticker.C:
			sm.mu.Lock()
			now := time.Now()
			for userID, stateData := range sm.states {
				if now.Sub(stateData.UpdatedAt) > 10*time.Minute {
					delete(sm.states, userID)
				}
			}
			sm.mu.Unlock()
		}
	}
}
