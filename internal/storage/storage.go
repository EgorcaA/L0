package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"

	"github.com/EgorcaA/create_db/internal/config"
	"github.com/EgorcaA/create_db/internal/order_struct"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Check if a database exists
func databaseExists(db *sql.DB, dbName string) (bool, error) {
	query := "SELECT 1 FROM pg_database WHERE datname = $1"
	var exists int
	err := db.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func checkTableExists(db *sql.DB, tableName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.tables 
			WHERE table_schema = 'public' AND table_name = $1
		)
	`

	var exists bool
	err := db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

type PostgresDB struct {
	Conn *sql.DB
}

//go:generate go run github.com/vektra/mockery/v2@v2.49.1 --name=Database --outpkg=mocks --dir=.
type Database interface {
	InsertOrder(ctx context.Context, order order_struct.Order) error
	GetAllOrders() ([]order_struct.Order, error)
}

func New(log *slog.Logger, Postgres_conf config.PostgresConfig) (*PostgresDB, error) {

	// Open a database connection
	// db, err := dbConn(db_name)
	// connStr := fmt.Sprintf("host=localhost port=5432 user=egor password=resu dbname=%s  sslmode=disable", db_name)

	serverConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable dbname=postgres",
		Postgres_conf.Host,
		Postgres_conf.Port,
		Postgres_conf.User,
		Postgres_conf.Password)
	serverDb, err := sql.Open("postgres", serverConnStr)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to connect to PostgreSQL server: %v", err))
	}
	log.Info("PostgreSQL server opened successfully!: %v", err)

	// Check if the database exists
	exists, err := databaseExists(serverDb, Postgres_conf.Name)
	if err != nil {
		log.Fatalf("Failed to check database existence: %v", err)
	}

	// Create the database if it does not exist
	if !exists {
		log.Println("creating tables")

		query := fmt.Sprintf("CREATE DATABASE %s;", Postgres_conf.Name)
		_, err = serverDb.Exec(query)
		if err != nil {
			log.Printf("Error creating database %s: %v (it might already exist)", Postgres_conf.Name, err)
		} else {
			log.Printf("Database %s created successfully!", Postgres_conf.Name)
		}
		serverConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable dbname=%s",
			Postgres_conf.Host,
			Postgres_conf.Port,
			Postgres_conf.User,
			Postgres_conf.Password,
			Postgres_conf.Name)
		serverDb, err = sql.Open("postgres", serverConnStr)
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL db %s: %v", Postgres_conf.Name, err)
		} else {
			log.Printf("Successfully connected to PostgreSQL db %s", Postgres_conf.Name)
		}

		query = `
		CREATE TABLE orders (
			order_uid VARCHAR PRIMARY KEY,
			track_number VARCHAR NOT NULL,
			entry VARCHAR NOT NULL,
			locale VARCHAR NOT NULL,
			internal_signature VARCHAR,
			customer_id VARCHAR NOT NULL,
			delivery_service VARCHAR NOT NULL,
			shardkey VARCHAR NOT NULL,
			sm_id INT NOT NULL,
			date_created TIMESTAMP NOT NULL,
			oof_shard VARCHAR NOT NULL
		);
		
		-- Таблица доставки
		CREATE TABLE delivery (
			order_uid VARCHAR PRIMARY KEY REFERENCES orders(order_uid),
			name VARCHAR NOT NULL,
			phone VARCHAR NOT NULL,
			zip VARCHAR NOT NULL,
			city VARCHAR NOT NULL,
			address VARCHAR NOT NULL,
			region VARCHAR NOT NULL,
			email VARCHAR NOT NULL
		);
		
		-- Таблица оплаты
		CREATE TABLE payment (
			transaction VARCHAR PRIMARY KEY,
			order_uid VARCHAR NOT NULL REFERENCES orders(order_uid),
			request_id VARCHAR,
			currency VARCHAR NOT NULL,
			provider VARCHAR NOT NULL,
			amount INT NOT NULL,
			payment_dt BIGINT NOT NULL,
			bank VARCHAR NOT NULL,
			delivery_cost INT NOT NULL,
			goods_total INT NOT NULL,
			custom_fee INT NOT NULL
		);
		
		-- Таблица товаров в заказе
		CREATE TABLE items (
			id SERIAL PRIMARY KEY,
			order_uid VARCHAR NOT NULL REFERENCES orders(order_uid),
			chrt_id BIGINT NOT NULL,
			track_number VARCHAR NOT NULL,
			price INT NOT NULL,
			rid VARCHAR NOT NULL,
			name VARCHAR NOT NULL,
			sale INT NOT NULL,
			size VARCHAR NOT NULL,
			total_price INT NOT NULL,
			nm_id BIGINT NOT NULL,
			brand VARCHAR NOT NULL,
			status INT NOT NULL
		);`
		_, err = serverDb.Exec(query)
		if err != nil {
			log.Fatalf("Failed to create tables: %v", err)
		} else {
			log.Printf("Created tables")
		}

		// log.Println("Migrations applied successfully!")

		// driver, _ := postgres.WithInstance(serverDb, &postgres.Config{})

		// // Load migrations from the migrations directory
		// mirgator, err := migrate.NewWithDatabaseInstance(
		// 	"file://migrations", // Path to the migrations folder
		// 	"postgres",          // Database driver name
		// 	driver,
		// )
		// if err != nil {
		// 	log.Fatalf("Failed to initialize migrations: %v", err)
		// }

		// // Apply all migrations
		// if err := mirgator.Up(); err != nil && err != migrate.ErrNoChange {
		// 	log.Fatalf("Failed to apply migrations: %v", err)
		// }
		// err = serverDb.Ping()
		// if err != nil {
		// 	log.Fatalf("Failed to ping the original database connection: %v", err)
		// }

		// log.Println("Migrations applied successfully!")

	} else {
		log.Printf("%s is already here", Postgres_conf.Name)
		serverConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable dbname=%s",
			Postgres_conf.Host,
			Postgres_conf.Port,
			Postgres_conf.User,
			Postgres_conf.Password,
			Postgres_conf.Name)
		serverDb, err = sql.Open("postgres", serverConnStr)
		if err != nil {
			log.Fatalf("Failed to open %s connection: %v", Postgres_conf.Name, err)
		}
	}
	return &PostgresDB{Conn: serverDb}, nil

}

func (db *PostgresDB) InsertOrder(ctx context.Context, order order_struct.Order) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	// connStr := "host=localhost port=5433 user=postgres password=qwerty sslmode=disable"
	// db, err := sql.Open("postgres", connStr)

	// func saveOrderToDB(ctx context.Context, db *sql.DB, order generator.Order) error {
	// query := `INSERT INTO orders (
	// 	order_uid, track_number, entry, locale, internal_signature,
	// 	customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
	// ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	// _, err := db.ExecContext(ctx, query, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
	// 	order.CustomerID, order.DeliveryService, order.ShardKey, order.SMID, order.DateCreated, order.OOFShard)

	// Insert into orders table
	_, err = tx.Exec(`
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.ShardKey, order.SMID, order.DateCreated, order.OOFShard)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert into deliveries table
	_, err = tx.Exec(`
		INSERT INTO delivery (
			order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City,
		order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert into payments table
	_, err = tx.Exec(`
		INSERT INTO payment (
			order_uid, transaction, request_id, currency, provider, amount,
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert into items table
	for _, item := range order.Items {
		_, err := tx.Exec(`
			INSERT INTO items (
				order_uid, chrt_id, track_number, price, rid, name, sale, size,
				total_price, nm_id, brand, status
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// GetAllOrders retrieves all orders along with their associated delivery, payment, and items.
func (db *PostgresDB) GetAllOrders() ([]order_struct.Order, error) {
	// SQL query to join the orders table with delivery, payment, and items
	query := `
		SELECT 
			o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, 
			o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
			d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
			p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, 
			p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
			i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, i.size, i.total_price, 
			i.nm_id, i.brand, i.status
		FROM orders o
		LEFT JOIN delivery d ON o.order_uid = d.order_uid
		LEFT JOIN payment p ON o.order_uid = p.order_uid
		LEFT JOIN items i ON o.order_uid = i.order_uid
	`

	// Execute the query
	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	// Slice to hold all orders
	var orders []order_struct.Order

	// Iterate over the rows and build the orders slice
	for rows.Next() {
		var order order_struct.Order
		var item order_struct.Item
		var delivery order_struct.Delivery
		var payment order_struct.Payment

		// Scan the row into the corresponding fields
		err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
			&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SMID, &order.DateCreated, &order.OOFShard,
			&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
			&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDT,
			&payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale, &item.Size, &item.TotalPrice,
			&item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}

		// Add delivery and payment data to the order
		order.Delivery = delivery
		order.Payment = payment

		// Add the item to the order's items slice
		order.Items = append(order.Items, item)

		// Append the order to the orders slice
		orders = append(orders, order)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return orders, nil
}
