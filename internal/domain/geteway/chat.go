package geteway

import (
	"context"

	"github.com/lucaslimasilvafoligem/chatservice/internal/domain/entity"
)

type ChatGateway interface {
	CreateChat(ctx context.Context, chat *entity.Chat) error
	FindChatById(ctx context.Context, chatID string) (*entity.Chat, error)
	SaveChat(ctx context.Context, chat *entity.Chat) error
}
