package metric

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// 1. Группа Kafka: только транспортные ошибки
	KafkaMessagesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "order",
		Subsystem: "kafka",
		Name:      "messages_received_total",
		Help:      "Сколько сообщений пришло из топика",
	}, []string{"status"}) // success (распарсили) / error (битый JSON)

	// 2.1 Группа Database: только ошибки записи/чтения
	DbOperationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "order",
		Subsystem: "db",
		Name:      "operations_total",
		Help:      "Статистика операций с БД",
	}, []string{"operation", "status"})

	//2.2 Гистограмма для БД
	DbDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "order",
		Subsystem: "db",
		Name:      "operation_duration_seconds",
		Help:      "Время выполнения операций с БД",
		Buckets:   prometheus.DefBuckets,
	}, []string{"operation"}) // "save" или "get"

	//4.1 Размер кеша
	CacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "order",
		Subsystem: "cache",
		Name:      "items_count",
		Help:      "Текущее количество заказов в оперативной памяти",
	})

	//4.2 коф попадаения в кеш
	CacheHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "order",
		Subsystem: "cache",
		Name:      "cof_items_count",
		Help:      "Текущее количество заказов в оперативной памяти",
	}, []string{"result"}) //hit-нашли, miss-нет

	//5 запросы
	RequestMetrics = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  "order",
		Subsystem:  "http",
		Name:       "request",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"status"})
)

func ObserveRequest(t time.Duration, status int) {
	RequestMetrics.WithLabelValues(strconv.Itoa(status)).Observe(t.Seconds())
}
