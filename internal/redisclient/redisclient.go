package redisclient

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	_ "github.com/EgorcaA/create_db/internal/config"
	"github.com/EgorcaA/create_db/internal/order_struct"
	"github.com/EgorcaA/create_db/internal/storage"
	"github.com/redis/go-redis/v9"
)

func InitRedis() (rdb *redis.Client, err error) {

	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Update with your Redis address
		DB:   0,                // Default DB
	})

	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis")
	return
}

func RestoreCacheFromDB(ctx context.Context, rdb *redis.Client, db *storage.PostgresDB) {
	// Fetch all orders from the database
	orders, err := db.GetAllOrders()
	if err != nil {
		log.Printf("Ошибка восстановления данных из базы данных: %v", err)
		return
	}

	for _, order := range orders {
		cacheKey := "order:" + order.OrderUID
		err := SaveOrder(ctx, rdb, order)
		// err := rdb.Set(ctx, cacheKey, order, time.Hour*24).Err() // Cache for 24 hours
		if err != nil {
			log.Printf("Ошибка кэширования заказа %s: %v", order.OrderUID, err)
		} else {
			log.Printf("Заказ восстановлен в кэш: %s", cacheKey)
		}
	}
	log.Println("Кэш успешно восстановлен из базы данных")
}

// func GetOrder(ctx context.Context, rdb *redis.Client, orderUID string) (order_struct.Order, error) {
// 	// Try to get the order from Redis
// 	cacheKey := "order:" + orderUID
// 	cachedOrder, err := rdb.Get(ctx, cacheKey).Result()
// 	if err == nil {
// 		var order order_struct.Order
// 		err = json.Unmarshal([]byte(cachedOrder), &order)
// 		if err == nil {
// 			log.Printf("Заказ получен из кэша: %s", cacheKey)
// 			return order, nil
// 		}
// 	}

// 	// If not found in cache, fetch from the database
// 	order, err := storage.GetOrderByUID(ctx, db, orderUID)
// 	if err != nil {
// 		return order, err
// 	}

// 	// Cache the order for future requests
// 	err = rdb.Set(ctx, cacheKey, order, time.Hour*24).Err()
// 	if err != nil {
// 		log.Printf("Ошибка кэширования заказа %s: %v", orderUID, err)
// 	}

// 	return order, nil
// }

func SaveOrder(ctx context.Context, rdb *redis.Client, order order_struct.Order) error {
	// Save general order details
	orderKey := "order:" + order.OrderUID
	orderData := map[string]interface{}{
		"OrderUID":          order.OrderUID,
		"TrackNumber":       order.TrackNumber,
		"Entry":             order.Entry,
		"Locale":            order.Locale,
		"InternalSignature": order.InternalSignature,
		"CustomerID":        order.CustomerID,
		"DeliveryService":   order.DeliveryService,
		"ShardKey":          order.ShardKey,
		"SMID":              order.SMID,
		"DateCreated":       order.DateCreated.Unix(),
		"OOFShard":          order.OOFShard,
	}
	if err := rdb.HSet(ctx, orderKey, orderData).Err(); err != nil {
		return err
	}

	// Save delivery details
	deliveryKey := orderKey + ":delivery"
	deliveryData := map[string]interface{}{
		"Name":    order.Delivery.Name,
		"Phone":   order.Delivery.Phone,
		"Zip":     order.Delivery.Zip,
		"City":    order.Delivery.City,
		"Address": order.Delivery.Address,
		"Region":  order.Delivery.Region,
		"Email":   order.Delivery.Email,
	}
	if err := rdb.HSet(ctx, deliveryKey, deliveryData).Err(); err != nil {
		return err
	}

	// Save payment details
	paymentKey := orderKey + ":payment"
	paymentData := map[string]interface{}{
		"Transaction":  order.Payment.Transaction,
		"RequestID":    order.Payment.RequestID,
		"Currency":     order.Payment.Currency,
		"Provider":     order.Payment.Provider,
		"Amount":       order.Payment.Amount,
		"PaymentDT":    order.Payment.PaymentDT,
		"Bank":         order.Payment.Bank,
		"DeliveryCost": order.Payment.DeliveryCost,
		"GoodsTotal":   order.Payment.GoodsTotal,
		"CustomFee":    order.Payment.CustomFee,
	}
	if err := rdb.HSet(ctx, paymentKey, paymentData).Err(); err != nil {
		return err
	}

	// Save items
	itemsKey := orderKey + ":items"
	for _, item := range order.Items {
		itemJSON, err := json.Marshal(item)
		if err != nil {
			return err
		}
		if err := rdb.RPush(ctx, itemsKey, itemJSON).Err(); err != nil {
			return err
		}
	}

	// Add order to customer's order list
	customerOrdersKey := "customer:" + order.CustomerID + ":orders"
	if err := rdb.SAdd(ctx, customerOrdersKey, order.OrderUID).Err(); err != nil {
		return err
	}

	return nil
}

func GetOrder(ctx context.Context, rdb *redis.Client, orderUID string) (order_struct.Order, error) {
	order := order_struct.Order{}

	// Retrieve general order details
	orderKey := "order:" + orderUID
	orderData, err := rdb.HGetAll(ctx, orderKey).Result()
	if err != nil {
		return order, err
	}
	order.OrderUID = orderData["OrderUID"]
	order.TrackNumber = orderData["TrackNumber"]
	order.Entry = orderData["Entry"]
	order.Locale = orderData["Locale"]
	order.InternalSignature = orderData["InternalSignature"]
	order.CustomerID = orderData["CustomerID"]
	order.DeliveryService = orderData["DeliveryService"]
	order.ShardKey, _ = strconv.Atoi(orderData["ShardKey"])
	order.SMID, _ = strconv.Atoi(orderData["SMID"])
	dateCreated, _ := strconv.ParseInt(orderData["DateCreated"], 10, 64)
	order.DateCreated = time.Unix(dateCreated, 0)
	order.OOFShard = orderData["OOFShard"]

	// Retrieve delivery details
	deliveryKey := orderKey + ":delivery"
	deliveryData, err := rdb.HGetAll(ctx, deliveryKey).Result()
	if err != nil {
		return order, err
	}
	order.Delivery = order_struct.Delivery{
		Name:    deliveryData["Name"],
		Phone:   deliveryData["Phone"],
		Zip:     deliveryData["Zip"],
		City:    deliveryData["City"],
		Address: deliveryData["Address"],
		Region:  deliveryData["Region"],
		Email:   deliveryData["Email"],
	}

	// Retrieve payment details
	paymentKey := orderKey + ":payment"
	paymentData, err := rdb.HGetAll(ctx, paymentKey).Result()
	if err != nil {
		return order, err
	}

	tmp_Amount, _ := strconv.Atoi(paymentData["Amount"])
	tmp_PaymentDT, _ := strconv.ParseInt(paymentData["PaymentDT"], 10, 64)
	tmp_DeliveryCost, _ := strconv.Atoi(paymentData["DeliveryCost"])
	tmp_GoodsTotal, _ := strconv.Atoi(paymentData["GoodsTotal"])
	tmp_CustomFee, _ := strconv.Atoi(paymentData["CustomFee"])

	order.Payment = order_struct.Payment{
		Transaction:  paymentData["Transaction"],
		RequestID:    paymentData["RequestID"],
		Currency:     paymentData["Currency"],
		Provider:     paymentData["Provider"],
		Amount:       tmp_Amount,
		PaymentDT:    tmp_PaymentDT,
		Bank:         paymentData["Bank"],
		DeliveryCost: tmp_DeliveryCost,
		GoodsTotal:   tmp_GoodsTotal,
		CustomFee:    tmp_CustomFee}

	// Retrieve items
	itemsKey := orderKey + ":items"
	itemsData, err := rdb.LRange(ctx, itemsKey, 0, -1).Result()
	if err != nil {
		return order, err
	}
	for _, itemJSON := range itemsData {
		var item order_struct.Item
		if err := json.Unmarshal([]byte(itemJSON), &item); err != nil {
			return order, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}
