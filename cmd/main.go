package main

import (
	"context"
	"encoding/json"
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
	"github.com/EgorcaA/create_db/internal/logger/sl"
	"github.com/EgorcaA/create_db/internal/order_struct"
	"github.com/EgorcaA/create_db/internal/redisclient"
	"github.com/EgorcaA/create_db/internal/server"
	"github.com/EgorcaA/create_db/internal/storage"
	"github.com/IBM/sarama"
	// "github.com/redis/go-redis/v9"
)

func main() {

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Current directory:", dir)

	cfg := config.MustLoad()

	log := sl.SetupLogger(cfg.App.Env)
	log.Info(
		"starting app",
		slog.String("env", cfg.App.Env),
		slog.String("version", "123"),
	)
	log.Debug("debug messages are enabled")

	rdb, _ := redisclient.InitRedis(cfg.Redis)

	db, err := storage.New(log, cfg.Postgres)
	if err != nil {
		// log.Fatalf("Failed to create storage instance: %v", err)
		log.Info(fmt.Sprintf("Failed to create storage instance: %v", err))

		os.Exit(1)
	}
	defer db.Conn.Close()

	//kafka
	brokers := []string{cfg.Kafka.BootstrapServers} // Kafka brockers
	topic := cfg.Kafka.Topic                        // def "orders"

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// Kafka connect
	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to connect to Kafka: %v", err))
	}
	defer client.Close()

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		log.Error(fmt.Sprintf("Failed creating Kafka consumer: %v", err))
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Error(fmt.Sprintf("Failed subribing the topic: %v", err))
	} else {
		log.Info("Succeded subribing Kafka topic")
	}
	defer partitionConsumer.Close()
	// kafka end

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Restore cache from database
	rdb.RestoreCacheFromDB(ctx, log, db)

	// Обработка системных сигналов
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// test_channel := make(chan order_struct.Order, 10)
	// go generator.Spam_channel(test_channel)
	go generator.Spam_kafka(log, cfg.Kafka)

	go func() {
		for {
			select {
			// case msg := <-test_channel:
			case msg, ok := <-partitionConsumer.Messages():
				if !ok {
					log.Warn("Messages channel closed, exiting goroutine")
					return
				}

				if msg.Value == nil {
					log.Warn("Message value is nil, skipping")
					continue
				}
				var order order_struct.Order
				if err := json.Unmarshal(msg.Value, &order); err != nil {
					log.Warn(fmt.Sprintf("Error unmarshal the message: %v", err))
					continue
				}
				handler.Handle_message(log, ctx, rdb, order, db)

			case <-signals:
				// log.Println("Получен сигнал завершения, выходим...")
				log.Info("Received termination signal...")
				cancel()
				return
			case err := <-partitionConsumer.Errors():
				log.Warn(fmt.Sprintf("Kafka error: %v", err))

				// case <-ctx.Done(): // Optional: Handle context cancellation
				// 	log.Info("Context canceled, exiting goroutine")
				// 	return
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
