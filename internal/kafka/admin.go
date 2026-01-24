package kafka

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

func EnsureTopicExists(broker []string, topic string) error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0

	//создаем клиента для управления кластером
	admin, err := sarama.NewClusterAdmin(broker, config)
	if err != nil {
		return fmt.Errorf("ошибка создания admin-клиента Kafka: %w", err)
	}
	defer func() {
		if err := admin.Close(); err != nil {
			log.Printf("failed to close kafka admin: %v", err)
		}
	}()

	//1. получаем список существующих топиков
	topics, err := admin.ListTopics()
	if err != nil {
		return fmt.Errorf("ошибка получения списка топиков: %w", err)
	}
	//2. если топик есть, выходим
	if _, exists := topics[topic]; exists {
		log.Printf("Kafka: топик '%s' уже существует", topic)
		return nil
	}
	//3. если нет, то конфигурируем новый
	topicDetails := &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
		ConfigEntries: map[string]*string{
			"retention.ms": strPtr("604800000"),
		},
	}

	err = admin.CreateTopic(topic, topicDetails, false)
	if err != nil {
		return fmt.Errorf("не удалось создать топик: %w", err)
	}

	log.Printf("Kafka: топик '%s' успешно создан", topic)
	return nil
}

func strPtr(s string) *string {
	return &s
}
