package chatcompletionstream

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/lucaslimasilvafoligem/chatservice/internal/domain/entity"
	"github.com/lucaslimasilvafoligem/chatservice/internal/domain/geteway"
	openai "github.com/sashabaranov/go-openai"
)

type ChatCompletionConfigInputDTO struct {
	Model                 string
	ModelMaxTokens        int
	Temperature           float32
	TopP                  float32
	N                     int
	Stop                  []string
	MaxTokens             int
	PresencePenalty       float32
	FrequencyPenalty      float32
	iInitialSystemMessage string
}

type ChatCompletionInputDTO struct {
	ChatID      string
	UserID      string
	UserMessage string
	Config      ChatCompletionConfigInputDTO
}

type ChatCompletionOutputDTO struct {
	ChatID  string
	UserID  string
	Content string
}

type ChatcompletionUseCase struct {
	ChatGateway  geteway.ChatGateway
	OpenAiClient *openai.Client
	Stream       chan ChatCompletionOutputDTO
}

func NewChatCompletUseCase(chatGateWay geteway.ChatGateway, opeAiClient *openai.Client, stream chan ChatCompletionOutputDTO) *ChatcompletionUseCase {
	return &ChatcompletionUseCase{
		ChatGateway:  chatGateWay,
		OpenAiClient: opeAiClient,
		Stream:       stream,
	}
}

func (uc *ChatcompletionUseCase) Execute(ctx context.Context, input ChatCompletionInputDTO) (*ChatCompletionOutputDTO, error) {
	chat, err := uc.ChatGateway.FindChatById(ctx, input.ChatID)
	if err != nil {
		if err.Error() == "chat not found" {
			chat, err = createNewChat(input)
			if err != nil {
				return nil, errors.New("error creating chat: " + err.Error())
			}
			err = uc.ChatGateway.CreateChat(ctx, chat)
			if err != nil {
				return nil, errors.New("error persisting chat: " + err.Error())
			}
		} else {
			return nil, errors.New("error fetching existing chat: " + err.Error())
		}
	}

	userMessage, err := entity.NewMessage("user", input.UserMessage, chat.Config.Model)
	if err != nil {
		return nil, errors.New("error creating user message: " + err.Error())
	}

	err = chat.AddMessage(userMessage)
	if err != nil {
		return nil, errors.New("error adding user message: " + err.Error())
	}

	messages := []openai.ChatCompletionMessage{}
	for _, msg := range chat.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	resp, err := uc.OpenAiClient.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:            chat.Config.Model.Name,
			Messages:         messages,
			MaxTokens:        chat.Config.MaxTokens,
			Temperature:      chat.Config.Temperature,
			TopP:             chat.Config.TopP,
			PresencePenalty:  chat.Config.PresencePenalty,
			FrequencyPenalty: chat.Config.FrequencyPenalty,
			Stop:             chat.Config.Stop,
			Stream:           true,
		},
	)
	if err != nil {
		return nil, errors.New("error creating chat completion: " + err.Error())
	}

	var fullResponse strings.Builder

	for {
		response, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, errors.New("error reading chat completion response: " + err.Error())
		}

		fullResponse.WriteString(response.Choices[0].Delta.Content)
		r := ChatCompletionOutputDTO{
			ChatID:  chat.ID,
			UserID:  input.UserID,
			Content: fullResponse.String(),
		}
		uc.Stream <- r
	}

	assistant, err := entity.NewMessage("assistant", fullResponse.String(), chat.Config.Model)
	if err != nil {
		return nil, errors.New("error creating assistant messasge: " + err.Error())
	}

	err = chat.AddMessage(assistant)
	if err != nil {
		return nil, errors.New("error adding assistant messasge: " + err.Error())
	}

	err = uc.ChatGateway.SaveChat(ctx, chat)
	if err != nil {
		return nil, errors.New("error saving chat: " + err.Error())
	}

	return &ChatCompletionOutputDTO{
		ChatID:  chat.ID,
		UserID:  input.UserID,
		Content: fullResponse.String(),
	}, nil
}

func createNewChat(input ChatCompletionInputDTO) (*entity.Chat, error) {
	model := entity.NewModel(input.Config.Model, input.Config.ModelMaxTokens)
	chatConfig := &entity.ChatConfig{
		Temperature:      input.Config.Temperature,
		TopP:             input.Config.TopP,
		N:                input.Config.N,
		Stop:             input.Config.Stop,
		MaxTokens:        input.Config.MaxTokens,
		PresencePenalty:  input.Config.PresencePenalty,
		FrequencyPenalty: input.Config.FrequencyPenalty,
		Model:            model,
	}
	initialMessage, err := entity.NewMessage("system", input.Config.iInitialSystemMessage, model)
	if err != nil {
		return nil, errors.New("error creting initial message: " + err.Error())
	}
	chat, err := entity.NewChat(input.UserID, initialMessage, chatConfig)
	if err != nil {
		return nil, errors.New("error creating new chat: " + err.Error())
	}
	return chat, nil
}
