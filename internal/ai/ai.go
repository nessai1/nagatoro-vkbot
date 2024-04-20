package ai

import "context"

type Assistant interface {
	AskPersonal(ctx context.Context, ownerID int, text string) (string, error)
}
