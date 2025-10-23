package invitecode

import "context"

type Store interface {
	Create(ctx context.Context, inviteCode *InviteCode) error
	Get(ctx context.Context, id uint) (*InviteCode, error)
	GetByCode(ctx context.Context, code string) (*InviteCode, error)
	GetWithUsage(ctx context.Context, code string) (*InviteCodeWithUsage, error)
	Update(ctx context.Context, inviteCode *InviteCode) error
	List(ctx context.Context, offset, limit int) ([]*InviteCode, error)
	Count(ctx context.Context) (int64, error)
	RecordUsage(ctx context.Context, usage *InviteCodeUsage) error
	GetUsageByUser(ctx context.Context, userID uint) (*InviteCodeUsage, error)
}
