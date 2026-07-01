// Package scheduler contains background job scheduling for the AI module.
package scheduler

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/go-co-op/gocron/v2"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	notifRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
)

// FactScheduler delivers daily fun facts via Telegram
type FactScheduler struct {
	scheduler          gocron.Scheduler
	funFactUseCase     *usecases.FunFactUseCase
	moodUseCase        *usecases.MoodUseCase
	personalityService *services.TelegramPersonalityService
	telegramRepo       notifRepos.TelegramRepository
	logger             *slog.Logger
	serverCtx          context.Context
}

// NewFactScheduler creates a new FactScheduler. serverCtx is the
// application lifecycle context — when canceled (SIGTERM), the periodic
// deliverDailyFact task derives its own ctx from it and short-circuits
// rather than continuing with context.Background() (issue #263 ADR-4).
// Pass context.Background() in tests for no-cancelation semantics.
func NewFactScheduler(
	serverCtx context.Context,
	funFactUseCase *usecases.FunFactUseCase,
	moodUseCase *usecases.MoodUseCase,
	personalityService *services.TelegramPersonalityService,
	telegramRepo notifRepos.TelegramRepository,
	logger *slog.Logger,
) (*FactScheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &FactScheduler{
		scheduler:          s,
		funFactUseCase:     funFactUseCase,
		moodUseCase:        moodUseCase,
		personalityService: personalityService,
		telegramRepo:       telegramRepo,
		logger:             logger,
		serverCtx:          serverCtx,
	}, nil
}

// Start starts the fact delivery scheduler
func (fs *FactScheduler) Start() error {
	// Daily fact delivery at 9:00 AM
	_, err := fs.scheduler.NewJob(
		gocron.CronJob("0 9 * * *", false),
		gocron.NewTask(fs.deliverDailyFact),
	)
	if err != nil {
		return err
	}

	fs.scheduler.Start()
	fs.logger.Info("Fact scheduler started - daily delivery at 9:00 AM")
	return nil
}

// Stop stops the scheduler
func (fs *FactScheduler) Stop() error {
	return fs.scheduler.Shutdown()
}

func (fs *FactScheduler) deliverDailyFact() {
	// Derive task ctx from the application lifecycle ctx so a SIGTERM
	// during a daily tick cancels in-flight work rather than running к
	// completion on context.Background().
	ctx := fs.serverCtx
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		fs.logger.Info("fact scheduler tick skipped — server shutting down", "error", err)
		return
	}

	// Get a random fact
	fact, err := fs.funFactUseCase.GetRandomFact(ctx)
	if err != nil {
		fs.logger.Error("failed to get random fact for delivery", "error", err)
		return
	}

	if fact == nil {
		fs.logger.Warn("no facts available for delivery")
		return
	}

	// Get current mood
	mood := entities.MoodContext{State: entities.MoodContent, TimeOfDay: "morning"}
	if fs.moodUseCase != nil {
		moodResponse, err := fs.moodUseCase.GetCurrentMood(ctx)
		if err == nil && moodResponse != nil {
			mood.State = entities.MoodState(moodResponse.State)
		}
	}

	// Get all active Telegram connections
	connections, err := fs.telegramRepo.GetActiveConnections(ctx)
	if err != nil {
		fs.logger.Error("failed to get active Telegram connections", "error", err)
		return
	}

	// Send fact to each connected user with rate limiting
	sent := 0
	for _, conn := range connections {
		chatID := strconv.FormatInt(conn.TelegramChatID, 10)
		if err := fs.personalityService.SendFactMessage(ctx, chatID, fact.Content, mood); err != nil {
			fs.logger.Error("failed to send fact to user",
				"chat_id", chatID,
				"error", err,
			)
			continue
		}
		sent++

		// Rate limit: ~30 msg/sec (Telegram limit). Use select so SIGTERM
		// during the 1s sleep cuts the batch short instead of running к
		// completion on a canceled ctx (issue #263 ADR-4 carry-forward).
		if sent%25 == 0 {
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				fs.logger.Info("fact batch interrupted — server shutting down",
					"sent_so_far", sent,
					"total_connections", len(connections),
				)
				return
			}
		}
	}

	fs.logger.Info("Daily fact delivered", "fact_id", fact.ID, "sent_to", sent, "total_connections", len(connections))
}
