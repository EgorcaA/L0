package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/EgorcaA/create_db/internal/redisclient"
)

// Главная страница с HTML-формой.
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

// Обработчик формы для получения пользователя.
func OrderHandler(ctx context.Context, rdb redisclient.CacheClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}

		// Извлекаем ID из формы.
		OrderUID := r.FormValue("OrderUID")
		// if err != nil {
		// 	http.Error(w, "некорректный формат ID", http.StatusBadRequest)
		// 	return
		// }

		// Получаем данные пользователя из базы.
		order, err := rdb.GetOrder(ctx, OrderUID)
		if err != nil {
			http.Error(w, "ошибка базы данных", http.StatusInternalServerError)
			log.Printf("ошибка: %v\n", err)
			return
		}
		// if order == nil {
		// 	http.Error(w, "пользователь не найден", http.StatusNotFound)
		// 	return
		// }

		// Возвращаем данные пользователя.
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	}
}

// // Основная функция.
// func main() {
// 	// Инициализируем базу данных.
// 	db, err := initDB()
// 	if err != nil {
// 		log.Fatalf("не удалось инициализировать базу данных: %v\n", err)
// 	}
// 	defer db.Close()

// 	// Регистрируем обработчики.
// 	http.HandleFunc("/", IndexHandler)        // Главная страница с HTML-формой.
// 	http.HandleFunc("/user", OrderHandler(db)) // Обработчик формы.

// 	// Запускаем HTTP-сервер.
// 	fmt.Println("Сервер запущен на http://localhost:8080")
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }
