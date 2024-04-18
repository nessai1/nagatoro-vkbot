package bot

import (
	"fmt"
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
