// brokers := []string{"localhost:9092"} // Адреса брокеров Kafka
// topic := "orders"                     // Топик

// Настройка Sarama
// config := sarama.NewConfig()
// config.Consumer.Return.Errors = true

// // Подключение к Kafka
// client, err := sarama.NewClient(brokers, config)
// if err != nil {
// 	log.Fatalf("Не удалось подключиться к Kafka: %v", err)
// }
// defer client.Close()

// consumer, err := sarama.NewConsumerFromClient(client)
// if err != nil {
// 	log.Fatalf("Не удалось создать Kafka consumer: %v", err)
// }
// defer consumer.Close()

// partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
// if err != nil {
// 	log.Fatalf("Не удалось подписаться на топик: %v", err)
// }
// defer partitionConsumer.Close()

//Handler
// case msg := <-partitionConsumer.Messages():
// if err := json.Unmarshal(msg.Value, &order); err != nil {

// case err := <-partitionConsumer.Errors():
// 	log.Printf("Ошибка Kafka: %v", err)

// case msg := <-partitionConsumer.Messages():
// if err := json.Unmarshal(msg.Value, &order); err != nil {
// case msg := <-test_channel:

// 	// var order generator.Order
// 	// if err := json.Unmarshal(msg.Value, &order); err != nil {
// 	// 	log.Printf("Ошибка декодирования сообщения: %v", err)
// 	// 	continue
// 	// }
// 	log.Println("Получено сообщение %v", msg)

// 	// Сохранение в базу данных
// 	// if err := saveOrderToDB(ctx, db, msg); err != nil { //order
// 	// 	log.Printf("Ошибка сохранения заказа в базу данных: %v", err)
// 	// } else {
// 	// 	log.Printf("Заказ успешно сохранен: %+v", msg) //order
// 	// }
// // case err := <-partitionConsumer.Errors():
// // 	log.Printf("Ошибка Kafka: %v", err)