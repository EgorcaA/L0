package generator

import (
	"github.com/EgorcaA/create_db/internal/order_struct"
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
