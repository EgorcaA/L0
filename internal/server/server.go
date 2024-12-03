package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/EgorcaA/create_db/internal/redisclient"
)

// Main HTML-form
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Найти заказ</title>
	</head>
	<body>
		<h1>Введите OrderUID</h1>
		<form action="/user" method="POST">
			<label for="id">OrderUID:</label>
			<input type="string" OrderUID="OrderUID" name="OrderUID" required>
			<button type="submit">Отправить</button>
		</form>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, tmpl)
}

// Order retrieve handler
func OrderHandler(ctx context.Context, rdb redisclient.CacheClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}

		OrderUID := r.FormValue("OrderUID")

		// getting order from cache
		order, err := rdb.GetOrder(ctx, OrderUID)
		if err != nil || order.OrderUID == "" {
			http.Error(w, "DB internal error", http.StatusInternalServerError)
			log.Printf("Cache search error: %v\n", err)
			return
		}

		// return
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	}
}
