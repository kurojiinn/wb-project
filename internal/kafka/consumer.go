package kafka

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"wb-project/internal/logger/sl"
	"wb-project/internal/metric"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type KafkaHeaderCarrier []*sarama.RecordHeader
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
			parCtx := otel.GetTextMapPropagator().Extract(ctx, KafkaHeaderCarrier(message.Headers))

			//трасировка
			tr := otel.Tracer("consumer")
			processCtx, span := tr.Start(parCtx, "Kafka.Consume",
				trace.WithSpanKind(trace.SpanKindConsumer)) //отмечаем что это консьюмер

			//логирование
			slog.Info("Сообщение из кафки прочитано: ",
				slog.String("topic", order.topic),
				slog.Int64("partition", int64(message.Partition)),
				slog.Int64("offset", message.Offset),
				sl.Traced(parCtx), // Связываем лог с трейсом
			)

			span.SetAttributes(
				attribute.String("message.kafka.topic", order.topic),
				attribute.Int("message.kafka.partition", 0),
				attribute.Int64("message.kafka.offset", message.Offset))

			if err := order.processor(processCtx, message.Value); err != nil {
				slog.Error("error processing message",
					slog.Any("error", err),
					sl.Traced(processCtx))
				span.RecordError(err)
				metric.KafkaMessagesTotal.WithLabelValues("error").Inc()
			} else {
				metric.KafkaMessagesTotal.WithLabelValues("success").Inc()
			}
			span.End()
		}
	}
}

func (order *OrderConsumer) Close() error {
	return order.consumer.Close()
}

func (c KafkaHeaderCarrier) Get(key string) string {
	for _, h := range c {
		if string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c KafkaHeaderCarrier) Set(key string, value string) {
}

func (c KafkaHeaderCarrier) Keys() []string {
	keys := make([]string, len(c))
	for i, h := range c {
		keys[i] = string(h.Key)
	}
	return keys
}
