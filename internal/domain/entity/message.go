package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID       string
	Role     string
	Content  string
	Tokens   int
	Model    *Model
	CreateAt time.Time
}

func NewMessage(role, content string, model *Model) (*Message, error) {

	codec := NewCodec()

	totalTokens, err := codec.Count(content)
	if err != nil {
		return nil, errors.New("failed to count tokens: " + err.Error())
	}

	msg := &Message{
		ID:       uuid.New().String(),
		Role:     role,
		Content:  content,
		Tokens:   totalTokens,
		Model:    model,
		CreateAt: time.Now(),
	}

	// Valida a mensagem
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	return msg, nil
}

func (m *Message) Validate() error {
	if m.Role != "user" && m.Role != "system" && m.Role != "assistant" {
		return errors.New("invalid role")
	}

	if m.Content == "" {
		return errors.New("content is empty")
	}

	if m.CreateAt.IsZero() {
		return errors.New("created_at is empty")
	}
	return nil
}

func (m *Message) GetQtdTokens() int {
	return m.Tokens
}
