package config

import "os"

type Config struct {
	DB          DBConfig
	KafkaConfig KafkaConfig
}
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type CacheConfig struct {
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	Group   string
}

func LoadConfig() *Config {
	dbconfig := DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "user"),
		Password: getEnv("DB_PASSWORD", "pass"),
		DBName:   getEnv("DB_NAME", "order_db"),
	}

	kafkaConf := KafkaConfig{
		Brokers: []string{getEnv("KAFKA_BROKER", "localhost:9092")},
		Topic:   getEnv("KAFKA_TOPIC", "test-new"),
	}

	return &Config{DB: dbconfig, KafkaConfig: kafkaConf}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
