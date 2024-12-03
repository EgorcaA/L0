package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EgorcaA/create_db/internal/config"
	"github.com/EgorcaA/create_db/internal/generator"
	"github.com/EgorcaA/create_db/internal/handler"
	"github.com/EgorcaA/create_db/internal/logger/handlers/slogpretty"
	"github.com/EgorcaA/create_db/internal/order_struct"
	"github.com/EgorcaA/create_db/internal/redisclient"
	"github.com/EgorcaA/create_db/internal/server"
	"github.com/EgorcaA/create_db/internal/storage"
	// "github.com/redis/go-redis/v9"
)

func main() {

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Current directory:", dir)

	cfg := config.MustLoad()

	log := setupLogger(cfg.App.Env)
	log.Info(
		"starting app",
		slog.String("env", cfg.App.Env),
		slog.String("version", "123"),
	)
	log.Debug("debug messages are enabled")

	rdb, _ := redisclient.InitRedis(cfg.Redis)

	db, err := storage.New(cfg.Postgres)
	if err != nil {
		// log.Fatalf("Failed to create storage instance: %v", err)
		log.Info(fmt.Sprintf("Failed to create storage instance: %v", err))

		os.Exit(1)
	}
	defer db.Conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Restore cache from database
	rdb.RestoreCacheFromDB(ctx, db)

	// Обработка системных сигналов
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// Чтение сообщений из Kafka
	test_channel := make(chan order_struct.Order, 10)

	go func() {
		for i := 1; i <= 3; i++ {
			order := generator.GenerateFakeOrder()
			test_channel <- order
			// fmt.Printf("Отправлен заказ: %+v\n", order)
			fmt.Printf("Отправлен заказ \n")
			time.Sleep(300 * time.Millisecond) // Симуляция нагрузки
		}
	}()

	go func() {
		for {
			select {
			case msg := <-test_channel:
				handler.Handle_message(log, ctx, rdb, msg, db)

			case <-signals:
				// log.Println("Получен сигнал завершения, выходим...")
				log.Info("Received termination signal...")
				cancel()
				return
			}
		}
	}()

	http.HandleFunc("/", server.IndexHandler)               // Главная страница с HTML-формой.
	http.HandleFunc("/user", server.OrderHandler(ctx, rdb)) // Обработчик формы.

	// Создание и запуск HTTP-сервера с возможностью корректного завершения
	srv := &http.Server{
		Addr: ":8080",
	}

	// Запуск сервера в горутине
	go func() {
		// log.Println("Сервер запущен на http://localhost:8080")
		log.Info("Server is up at http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// log.Fatalf("Ошибка при запуске сервера: %v", err)
			log.Error(fmt.Sprintf("Error starting server: %v", err))
		}
	}()

	// Ожидаем сигнала завершения
	<-signals
	log.Info("Received termination signal, finalizing...")
	// log.Println("Сигнал получен, начинаем завершение работы...")

	// Попытка корректно завершить работу HTTP-сервера
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Error(fmt.Sprintf("Error closing server: %v", err))
		// log.Fatalf("Ошибка при завершении работы сервера: %v", err)
	}
	rdb.Conn.Close()
	log.Info("Server is closed")
	// log.Println("Сервер успешно завершен.")

	// <-ctx.Done()
}

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		logFile, err := os.OpenFile("slog.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("Failed to open log file: " + err.Error())
		}
		// defer logFile.Close()

		log = slog.New(
			slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
