package kafka

import (
	"encoding/json"
	"fmt"
	"time"
	"wb-project/internal/models"

	"github.com/IBM/sarama"
	"github.com/brianvoe/gofakeit"
)

type OrderProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewProducer(broker []string, topic string) (*OrderProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll // Ждем подтверждения от всех брокеров

	producer, err := sarama.NewSyncProducer(broker, config)
	if err != nil {
		return &OrderProducer{}, fmt.Errorf("не удалось создать продюсера: %v", err)
	}
	return &OrderProducer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (pr *OrderProducer) Send() error {
	//1. Генерация данных
	order := generateFakeOrders()
	//2. Парсинг данных
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("ошибка при парсинге для отправки %v", err)
	}
	//3. Создания сообщения
	message := &sarama.ProducerMessage{
		Topic: pr.topic,
		Value: sarama.ByteEncoder(data),
	}
	//4. Отправка сообщений в кафку
	partition, offset, err := pr.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("ошибка при отправке данных в кафку")
	}
	fmt.Printf("Заказ отправлен успешно!\n")
	fmt.Printf("UID: %s\nPartition: %d\nOffset: %d\n", order.OrderUID, partition, offset)
	return nil
}

func generateFakeOrders() models.Order {
	return models.Order{
		OrderUID:    gofakeit.UUID(),
		TrackNumber: "WB-" + gofakeit.Numerify("##########"),
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    gofakeit.Name(),
			Phone:   "+79" + gofakeit.Numerify("#########"), // e164 формат
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Address().Address,
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		},
		Payment: models.Payment{
			Transaction:  gofakeit.UUID(),
			RequestID:    gofakeit.Numerify("##########"),
			Currency:     "RUB", // ровно 3 символа
			Provider:     "wbpay",
			Amount:       gofakeit.Number(100, 100000), // gt=0
			PaymentDt:    int(time.Now().Unix()),
			Bank:         "alpha",
			DeliveryCost: 500,
			GoodsTotal:   gofakeit.Number(100, 100000),
			CustomFee:    10,
		},
		Items: []models.Items{
			{
				ChrtID:      gofakeit.Number(1, 1000),
				TrackNumber: "TRK-" + gofakeit.Numerify("#####"),
				Price:       gofakeit.Number(100, 10000),
				Rid:         gofakeit.UUID(),
				Name:        gofakeit.Name(),
				Sale:        gofakeit.Number(0, 50),
				Size:        "XL",
				TotalPrice:  gofakeit.Number(100, 10000),
				NmID:        gofakeit.Number(1, 1000000),
				Brand:       gofakeit.Company(),
				Status:      202,
			},
		},
		Locale:            "ru",
		InternalSignature: "",
		CustomerID:        gofakeit.UUID(),
		DeliveryService:   "meest",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}
}

func (order *OrderProducer) Close() error {
	return order.producer.Close()
}
