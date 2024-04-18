package bot

import (
	"fmt"
	"github.com/SevereCloud/vksdk/api"
	"github.com/SevereCloud/vksdk/longpoll-user"
	wrapper "github.com/SevereCloud/vksdk/longpoll-user/v3"
	"github.com/nessai1/nagatoro-vkbot/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
)

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
}

func (s *Service) ListenAndServe() error {
	s.logger.Info("Bot started!", zap.String("address", s.config.Address))

	vk := api.NewVK(s.config.VK.GroupAPIKey)
	lp, err := longpoll.NewLongpoll(vk, 12)
	if err != nil {
		return fmt.Errorf("cannot create longpoll: %w", err)
	}

	w := wrapper.NewWrapper(lp)
	w.OnNewMessage(func(m wrapper.NewMessage) {
		s.logger.Info("Got new message", zap.String("msg", m.Text))
	})

	err = lp.Run()
	if err != nil {
		s.logger.Error("Error while run longpoll", zap.Error(err))
	}

	return nil
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
