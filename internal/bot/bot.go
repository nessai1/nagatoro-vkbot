package bot

import (
	"context"
	"fmt"
	"github.com/SevereCloud/vksdk/api"
	"github.com/SevereCloud/vksdk/callback"
	"github.com/SevereCloud/vksdk/object"
	"github.com/nessai1/nagatoro-vkbot/internal/ai"
	"github.com/nessai1/nagatoro-vkbot/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"
)

const groupChatOffset = 2000000000

func Run(cfg config.Config) error {
	logger := initLogger()

	defer func() {
		err := logger.Sync()
		if err != nil {
			log.Printf("[LOGGER ERROR] Cannot sync logger: %v", err)
		}
	}()

	service := Service{
		logger: logger,
		config: cfg,
	}
	if err := service.ListenAndServe(); err != nil {
		return fmt.Errorf("cannot start bot service listening: %w", err)
	}

	return nil
}

type Service struct {
	logger *zap.Logger
	config config.Config

	assistant ai.Assistant

	vk *api.VK
}

func (s *Service) ListenAndServe() error {
	s.logger.Info("Bot started!", zap.String("address", s.config.Address))

	cb := callback.NewCallback()
	cb.ConfirmationKey = "99a210d6"

	s.vk = api.NewVK(s.config.VK.GroupAPIKey)

	var err error
	s.assistant, err = ai.NewGPT4(s.config.OpenAI, "preprompts/ru")
	if err != nil {
		return fmt.Errorf("cannot init GPT4 assistant: %w", err)
	}

	cb.MessageNew(func(object object.MessageNewObject, i int) {
		if object.Message.PeerID != object.Message.FromID {
			go s.HandleChatMessage(object.Message.PeerID-groupChatOffset, object.Message)
		} else {
			go s.HandlePersonalMessage(object.Message)
		}
	})

	http.HandleFunc("/callback", cb.HandleFunc)
	if err = http.ListenAndServe(s.config.Address, nil); err != nil {
		s.logger.Error("Cannot listen callback server", zap.Error(err))
	}
	return nil
}

func (s *Service) HandlePersonalMessage(message object.MessagesMessage) {
	s.logger.Debug("Got personal message", zap.String("message", message.Text), zap.Int("from_id", message.FromID))

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3600*time.Second) // TODO: change timeout
	defer cancel()

	answer, err := s.assistant.AskPersonal(ctx, message.FromID, message.Text)
	if err != nil {
		s.logger.Error("Cannot ask assistant in personal chat", zap.Error(err))

		return
	}

	_, err = s.vk.MessagesSend(map[string]interface{}{
		"user_id":   message.FromID,
		"random_id": rand.Int32(),
		"message":   answer,
	})

	if err != nil {
		s.logger.Error("Cannot send message to user", zap.Error(err))
	}
}

func (s *Service) HandleChatMessage(chatID int, message object.MessagesMessage) {
	if !s.hasAssistantMention(message.Text) {
		return
	}

	s.logger.Debug("Got chat message", zap.String("message", message.Text), zap.Int("from_id", message.FromID), zap.Int("chat_id", chatID))

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 3600*time.Second) // TODO: change timeout
	defer cancel()

	answer, err := s.assistant.AskPersonal(ctx, groupChatOffset, message.Text)
	if err != nil {
		s.logger.Error("Cannot ask assistant in group chat", zap.Error(err))

		return
	}

	_, err = s.vk.MessagesSend(map[string]interface{}{
		"reply_to":  message.ID,
		"chat_id":   chatID,
		"random_id": rand.Int32(),
		"message":   answer,
	})

	if err != nil {
		s.logger.Error("Cannot send message to chat", zap.Error(err))
	}
}

func (s *Service) hasAssistantMention(messageText string) bool {
	// TODO: take mention from config
	return strings.Contains(messageText, "[club225584757|@nagatorotoro]")
}

func initLogger() *zap.Logger {
	atom := zap.NewAtomicLevel()

	atom.SetLevel(zapcore.DebugLevel)
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))

	return logger
}
