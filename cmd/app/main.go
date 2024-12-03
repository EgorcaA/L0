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
)

func main() {

	//setting up env
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	cfg := config.MustLoad()

	log := sl.SetupLogger(cfg.App.Env)
	log.Info(
		"starting app",
		slog.String("env", cfg.App.Env),
		slog.String("version", "123"),
	)
	log.Debug("debug messages are enabled")
	log.Debug(fmt.Sprintf("Current directory: %s", dir))

	//redis init
	rdb, _ := redisclient.InitRedis(cfg.Redis, log)

	//db init
	db, err := storage.New(log, cfg.Postgres)
	if err != nil {
		log.Info(fmt.Sprintf("Failed to create storage instance: %v", err))
		os.Exit(1)
	}
	defer db.Conn.Close()

	//kafka
	brokers := []string{cfg.Kafka.BootstrapServers} // Kafka brockers
	topic := cfg.Kafka.Topic                        // def "orders"

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

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

	// System signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	//in case of debug use test_channel
	// test_channel := make(chan order_struct.Order, 10)
	// go generator.Spam_channel(test_channel)
	//smaple script to spam into kafka channel
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
				log.Info("Received termination signal...")
				cancel()
				return
			case err := <-partitionConsumer.Errors():
				log.Warn(fmt.Sprintf("Kafka error: %v", err))
			}
		}
	}()

	http.HandleFunc("/", server.IndexHandler)
	http.HandleFunc("/user", server.OrderHandler(ctx, rdb))

	srv := &http.Server{
		Addr: ":8080",
		// ReadTimeout:  cfg.HTTPServer.Timeout,
		// WriteTimeout: cfg.HTTPServer.Timeout,
		// IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	// server run
	go func() {
		log.Info("Server is up at http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(fmt.Sprintf("Error starting server: %v", err))
		}
	}()

	// Waiting termination signal
	<-signals
	log.Info("Received termination signal, finalizing...")

	// Ending HTTP-server
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Error(fmt.Sprintf("Error closing server: %v", err))
	}
	rdb.Conn.Close()
	log.Info("Server is closed")
}
