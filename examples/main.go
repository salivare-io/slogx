package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/salivare-io/slogx"
)

func main() {
	// Initialization using Builder-options and Variadic functions
	log := slogx.New(
		slogx.WithLevel(slogx.LevelTrace),
		slogx.WithFormat(slogx.FormatText),

		// Use Builder to mask
		slogx.WithMaskRules(
			slogx.NewMaskRules().
				Add("user_email", slogx.MaskEmail).
				Add("user_phone", slogx.MaskPhone).
				Add("card_number", slogx.MaskCard),
		),

		//Just list the keys to delete with a comma (Variadic)
		slogx.WithRemoval(
			slogx.NewRemovalSet().
				Add("password").
				Add("session_token").
				Add("auth_cookie"),
		),

		// Auto-context extraction
		slogx.WithContextKeys("trace_id", "request_id"),
	)

	// Set as global (optional, but useful for libraries)
	slog.SetDefault(log.Logger)

	// Working with context
	ctx := context.Background()
	ctx = context.WithValue(ctx, "trace_id", "tid_999")
	ctx = context.WithValue(ctx, "request_id", "req_777")

	fmt.Println("Текстовый формат и маскирование (Builder)")
	log.InfoContext(
		ctx, "user data processed",
		slog.String("email", "test@gmail.com"),
		slog.String("phone", "+79111234567"),
		slog.String("secret", "this_will_be_removed"),
	)

	// Dynamic configuration change of the entire system
	fmt.Println("\n Динамическая смена конфига на JSON")
	log.UpdateConfig(
		func(c *slogx.Config) {
			c.Format = slogx.FormatJSON
			c.Level = slog.LevelInfo
		},
	)

	log.TraceContext(ctx, "Этот лог уровня TRACE будет пропущен")
	log.InfoContext(ctx, "Этот лог теперь в JSON формате", "event", "config_updated")

	// Error logging via Helper Err()
	fmt.Println("\n Логирование ошибок")
	err := fmt.Errorf("timeout: соединение с БД разорвано")
	log.ErrorContext(ctx, "Сбой операции", slogx.Err(err))

	// Working through context (Dependency Injection)
	fmt.Println("\n Логгер внутри контекста ")
	ctxWithLogger := slogx.ToContext(ctx, log)
	doWork(ctxWithLogger)
}

func doWork(ctx context.Context) {
	// Извлекаем логгер. Если его нет, FromContext вернет безопасный дефолтный логгер
	log := slogx.FromContext(ctx)
	log.InfoContext(ctx, "Функция doWork выполнена")
}
