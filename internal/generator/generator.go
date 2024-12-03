package generator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/EgorcaA/create_db/internal/config"
	"github.com/EgorcaA/create_db/internal/order_struct"
	"github.com/IBM/sarama"
	"github.com/brianvoe/gofakeit/v7"
)

func GenerateFakeOrder() order_struct.Order {
	return order_struct.Order{
		OrderUID:    gofakeit.UUID(),
		TrackNumber: gofakeit.UUID(),
		Entry:       gofakeit.Word(),
		Delivery: order_struct.Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Street(),
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		},
		Payment: order_struct.Payment{
			Transaction:  gofakeit.UUID(),
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       gofakeit.Number(100, 5000),
			PaymentDT:    gofakeit.Date().Unix(),
			Bank:         gofakeit.Company(),
			DeliveryCost: gofakeit.Number(100, 2000),
			GoodsTotal:   gofakeit.Number(50, 2000),
			CustomFee:    0,
		},
		Items: []order_struct.Item{
			{
				ChrtID:      gofakeit.Number(1000000, 9999999),
				TrackNumber: gofakeit.UUID(),
				Price:       gofakeit.Number(100, 1000),
				RID:         gofakeit.UUID(),
				Name:        gofakeit.Word(),
				Sale:        gofakeit.Number(0, 50),
				Size:        gofakeit.Word(),
				TotalPrice:  gofakeit.Number(50, 500),
				NmID:        gofakeit.Number(100000, 999999),
				Brand:       gofakeit.Company(),
				Status:      gofakeit.Number(1, 10),
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        gofakeit.UUID(),
		DeliveryService:   gofakeit.Company(),
		ShardKey:          gofakeit.Number(1, 10),
		SMID:              gofakeit.Number(1, 100),
		DateCreated:       gofakeit.Date(),
		OOFShard:          "1",
	}
}

// func main() {
// 	fakeOrder := generateFakeOrder()
// 	fmt.Printf("Generated Fake Order: %+v\n", fakeOrder)
// }

func Spam_channel(test_channel chan order_struct.Order) {
	for i := 1; i <= 3; i++ {
		order := GenerateFakeOrder()
		test_channel <- order
		// fmt.Printf("Отправлен заказ: %+v\n", order)
		fmt.Printf("Отправлен заказ \n")
		time.Sleep(300 * time.Millisecond) // Симуляция нагрузки
	}
}

func Spam_kafka(log *slog.Logger, kafka_conf config.KafkaConfig) {
	// Kafka broker address (adjust to match your setup)
	brokers := []string{kafka_conf.BootstrapServers}

	// Create a new Sarama configuration
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true // Wait for success response
	config.Producer.Return.Errors = true    // Capture errors

	// Create a new synchronous producer
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create producer: %v", err))
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Error(fmt.Sprintf("Failed to close producer: %v", err))
		}
	}()
	log.Info("Succeded to create kafka producer")

	// Create a message
	for a := 0; a < 5; a++ {
		orderJSON, err := json.Marshal(GenerateFakeOrder())
		if err != nil {
			log.Error(fmt.Sprintf("Failed to serialize order for kafka producer: %v", err))
		} else {
			message := &sarama.ProducerMessage{
				Topic: kafka_conf.Topic,
				Value: sarama.StringEncoder(orderJSON),
			}

			// Send the message
			partition, offset, err := producer.SendMessage(message)
			if err != nil {
				log.Warn(fmt.Sprintf("Failed to send message: %v", err))
			}

			// Success confirmation
			log.Debug(fmt.Sprintf("Message sent to partition %d with offset %d\n", partition, offset))
			time.Sleep(300 * time.Millisecond)

		}
	}
}
