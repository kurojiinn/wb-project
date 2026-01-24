package kafka

import (
	"context"
	"fmt"
	"log"
	"wb-project/internal/metric"

	"github.com/IBM/sarama"
)

type MessageProcessor func(context.Context, []byte) error
type OrderConsumer struct {
	consumer sarama.Consumer
	topic    string
	// Это может быть сервис, который умеет валидировать и сохранять.
	processor MessageProcessor
}

func NewOrderConsumer(broker []string, topic string, processor MessageProcessor) (*OrderConsumer, error) {
	conf := sarama.NewConfig()
	// Указываем, откуда будет читать наш консьюмер
	conf.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumer(broker, conf)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании консьюмера: %w", err)
	}
	return &OrderConsumer{consumer: consumer, topic: topic, processor: processor}, nil
}

//Подключиться и подписаться на канал сообщений: настроить получение данных из брокера сообщений (Kafka).

func (order *OrderConsumer) Start(ctx context.Context) error {
	//подключение к партициям(test-new), номер партиции(0), откуда начинаем читать(с новых сообщенией)
	partitionConsumer, err := order.consumer.ConsumePartition(order.topic, 0, sarama.OffsetNewest)
	if err != nil {
		panic("Не удалось устроить работу с партициями")
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Printf("ошибка при закрытии partitionConsumer: %v", err)
		}
	}()

	//Messages() возвращает сообщения из партиций.
	for {
		select {
		case <-ctx.Done(): //1. шаг 1 graceful shutdown
			log.Println("Kafka consumer stopping...")
			return ctx.Err()
		case message := <-partitionConsumer.Messages():
			if err := order.processor(ctx, message.Value); err != nil {
				log.Printf("Error processing message: %v", err)
				metric.KafkaMessagesTotal.WithLabelValues("error").Inc()
			} else {
				metric.KafkaMessagesTotal.WithLabelValues("success").Inc()
			}
			//логируем сообщения, которые читаем
			fmt.Printf("Сообщение: %s ", string(message.Value))
		}
	}
}

func (order *OrderConsumer) Close() error {
	return order.consumer.Close()
}
