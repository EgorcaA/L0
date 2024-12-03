package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/EgorcaA/create_db/internal/order_struct"
	"github.com/EgorcaA/create_db/internal/redisclient"
	"github.com/EgorcaA/create_db/internal/storage"
)

// main kafka messages handler
func Handle_message(log *slog.Logger, ctx context.Context,
	rdb redisclient.CacheClient, msg order_struct.Order, db storage.Database) {

	log.Debug("Got new message", slog.String("OrderUID", msg.OrderUID))

	// Saving to db
	err := db.InsertOrder(ctx, msg)
	if err != nil {
		log.Debug(fmt.Sprintf("Error saving order in DB: %v", err),
			slog.String("OrderUID", msg.OrderUID))
	} else {
		log.Debug("Order is saved in DB", slog.String("OrderUID", msg.OrderUID))

		// Cache the order in Redis
		err = rdb.SaveOrder(ctx, msg)
		if err != nil {
			log.Warn(fmt.Sprintf("Order is saved in Cache: %+v", msg),
				slog.String("OrderUID", msg.OrderUID))
		} else {
			log.Debug("Order is saved in Cache", slog.String("OrderUID", msg.OrderUID))
		}
	}
}
