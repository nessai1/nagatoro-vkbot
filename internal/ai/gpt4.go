package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nessai1/nagatoro-vkbot/internal/config"
	"github.com/nessai1/nagatoro-vkbot/internal/storage"
	"os"
	"path/filepath"
	"time"
)
import openai "github.com/sashabaranov/go-openai"

type GPT4 struct {
	client    *openai.Client
	assistant *openai.Assistant

	storage storage.Storage

	memoryThreads map[int]openai.Thread // chatID -> thread
}

func NewGPT4(cfg config.OpenAIConfig, prepromptsDir string) (*GPT4, error) {
	s, err := storage.CreateStorage()
	if err != nil {
		return nil, fmt.Errorf("cannot create storage: %w", err)
	}

	f, err := os.Open(filepath.Join(prepromptsDir, "init.txt"))
	if err != nil {
		return nil, fmt.Errorf("cannot open preprompt file 'init.txt': %w", err)
	}
	defer f.Close()

	b := bytes.Buffer{}
	_, err = b.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("cannot read preprompt file 'init.txt': %w", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	assistantName := "Хаясе Нагаторо"
	assistantPreprompt := string(b.Bytes())

	client := openai.NewClient(cfg.APIToken)

	assistant, err := initAssistant(client, assistantName, assistantPreprompt)
	if err != nil {
		return nil, fmt.Errorf("cannot init assistant: %w", err)
	}

	gpt := GPT4{
		client:    client,
		assistant: assistant,

		storage: s,
	}

	return &gpt, nil
}

func initAssistant(client *openai.Client, name string, preprompt string) (*openai.Assistant, error) {
	fileAssistant, err := initAssistantFromFile() // TODO: use client.RetrieveAssistant instead
	if err == nil {
		return fileAssistant, nil
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	assistant, err := client.CreateAssistant(ctx, openai.AssistantRequest{
		Model:        openai.GPT4Turbo,
		Name:         &name,
		Instructions: &preprompt,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create assistant: %w", err)
	}

	bs, err := json.Marshal(assistant)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal assistant: %w", err)
	}

	f, err := os.Create("assistant.json")
	if err != nil {
		return nil, fmt.Errorf("cannot create assistant.json file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(bs)
	if err != nil {
		return nil, fmt.Errorf("cannot write assistant.json file: %w", err)
	}

	return &assistant, nil
}

func initAssistantFromFile() (*openai.Assistant, error) {
	f, err := os.Open("assistant.json")
	if err != nil {
		return nil, fmt.Errorf("cannot open assistant.json file: %w", err)
	}
	defer f.Close()

	b := bytes.Buffer{}
	_, err = b.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("cannot read assistant.json file: %w", err)
	}

	assistant := openai.Assistant{}
	err = json.Unmarshal(b.Bytes(), &assistant)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal assistant.json file: %w", err)
	}

	return &assistant, nil
}

func (gpt *GPT4) AskPersonal(ctx context.Context, ownerID int, text string) (string, error) {
	thread, err := gpt.loadThread(ctx, ownerID)
	if err != nil {
		return "", fmt.Errorf("cannot load thread: %w", err)
	}

	_, err = gpt.client.CreateMessage(ctx, thread.ID, openai.MessageRequest{
		Role:    openai.ChatMessageRoleUser,
		Content: text,
	})

	if err != nil {
		return "", fmt.Errorf("cannot create message: %w", err)
	}

	run, err := gpt.client.CreateRun(ctx, thread.ID, openai.RunRequest{
		AssistantID: gpt.assistant.ID,
	})
	if err != nil {
		return "", fmt.Errorf("cannot create run: %w", err)
	}

	for {
		<-time.After(3 * time.Second)
		run, err = gpt.client.RetrieveRun(ctx, thread.ID, run.ID)
		if err != nil {
			return "Сенпай, отвали от меня пока", nil
		}
		if run.Status == openai.RunStatusCompleted || run.Status == openai.RunStatusFailed {
			break
		}
	}

	if run.Status != openai.RunStatusCompleted {
		return "", fmt.Errorf("run status is not completed: %s", run.Status)
	}

	l, err := gpt.client.ListMessage(ctx, thread.ID, nil, nil, nil, nil)

	return l.Messages[0].Content[0].Text.Value, nil
}

func (gpt *GPT4) loadThread(ctx context.Context, chatID int) (openai.Thread, error) {
	thread, ok := gpt.memoryThreads[chatID]
	if ok {
		return thread, nil
	}

	chat, err := gpt.storage.GetChat(ctx, chatID)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return openai.Thread{}, fmt.Errorf("cannot get thread ID: %w", err)
	}

	if errors.Is(err, storage.ErrNotFound) {
		thread, err := gpt.createThread(ctx)
		if err != nil {
			return openai.Thread{}, fmt.Errorf("cannot create thread: %w", err)
		}

		_, err = gpt.storage.CreateChat(ctx, chatID, thread.ID)
		if err != nil {
			return openai.Thread{}, fmt.Errorf("cannot create chat: %w", err)
		}

		return thread, nil
	}

	thread, err = gpt.getThread(ctx, chat.ThreadID)
	if err != nil {
		return openai.Thread{}, fmt.Errorf("cannot get thread: %w", err)
	}

	return thread, nil
}

func (gpt *GPT4) createThread(ctx context.Context) (openai.Thread, error) {
	thread, err := gpt.client.CreateThread(ctx, openai.ThreadRequest{})
	if err != nil {
		return openai.Thread{}, fmt.Errorf("cannot create thread: %w", err)
	}

	return thread, nil
}

func (gpt *GPT4) getThread(ctx context.Context, threadID string) (openai.Thread, error) {
	thread, err := gpt.client.RetrieveThread(ctx, threadID)
	if err != nil {
		return openai.Thread{}, fmt.Errorf("cannot retrieve thread: %w", err)
	}

	return thread, nil
}
