# Метрики

Справочник метрик, собираемых go-activerecord.

## Интерфейс

```go
type MetricInterface interface {
    StatCount(namespace, name string, cnt int)
    StatTime(namespace, operation string, dt time.Duration)
    StatErr(namespace, name string, cnt int)
}
```

## Временные метрики

Время выполнения операций (StatTime):

| Метрика | Описание |
|---------|----------|
| `select_db` | Время SELECT запроса |
| `update_db` | Время UPDATE запроса |
| `delete_db` | Время DELETE запроса |
| `insertreplace_db` | Время INSERT/REPLACE запроса |
| `call_proc` | Время вызова процедуры |
| `select_pack` | Время упаковки SELECT |
| `select_process` | Время обработки результата |
| `select_newobj` | Время создания объектов |
| `insertreplace_pack` | Время упаковки INSERT |
| `insertreplace_packtuple` | Время упаковки tuple |

### Deprecated

| Метрика | Замена |
|---------|--------|
| `select_box` | `select_db` |
| `update_box` | `update_db` |
| `delete_box` | `delete_db` |
| `insertreplace_box` | `insertreplace_db` |

## Счётчики операций

Количество операций (StatCount):

| Метрика | Описание |
|---------|----------|
| `insert_request` | Запросы на Insert |
| `insert_success` | Успешные Insert |
| `insertorreplace_request` | Запросы на InsertOrReplace |
| `insertorreplace_success` | Успешные InsertOrReplace |
| `replace_request` | Запросы на Replace |
| `replace_success` | Успешные Replace |
| `update_request` | Запросы на Update |
| `update_success` | Успешные Update |
| `update_empty` | Update без изменений |
| `update_repaired` | Update восстановленных объектов |
| `delete_request` | Запросы на Delete |
| `delete_success` | Успешные Delete |
| `select_keys` | Количество ключей в SELECT |
| `select_tuples_res` | Количество полученных записей |

### Bulk операции

| Метрика | Описание |
|---------|----------|
| `bulk_update_request` | Вызовы BulkUpdate |
| `bulk_update_objects` | Объектов в BulkUpdate |
| `bulk_update_success` | Обновлённых строк |
| `bulk_update_empty` | Объектов без изменений |
| `bulk_update_notexists` | Несуществующих объектов |
| `bulk_update_failed` | Неудачных BulkUpdate |
| `bulk_insertorreplace_request` | Вызовы BulkInsertReplace |
| `bulk_insertorreplace_success` | Вставленных строк |

## Метрики ошибок

Количество ошибок (StatErr):

| Метрика | Описание |
|---------|----------|
| `select_db` | Ошибки SELECT |
| `update_db` | Ошибки UPDATE |
| `delete_db` | Ошибки DELETE |
| `insertreplace_db` | Ошибки INSERT/REPLACE |
| `insert_exists` | Конфликт при Insert |
| `replace_notexists` | Replace несуществующего |
| `update_notexists` | Update несуществующего |
| `call_proc` | Ошибки вызова процедуры |

### Ошибки подготовки

| Метрика | Описание |
|---------|----------|
| `select_preparebox` | Ошибка подготовки SELECT |
| `select_resp` | Ошибка ответа SELECT |
| `update_preparedb` | Ошибка подготовки UPDATE |
| `update_resp` | Ошибка ответа UPDATE |
| `update_packpk` | Ошибка упаковки PK |
| `delete_preparedb` | Ошибка подготовки DELETE |
| `delete_resp` | Ошибка ответа DELETE |
| `delete_pack` | Ошибка упаковки DELETE |
| `insertreplace_preparedb` | Ошибка подготовки INSERT |
| `insertreplace_obj` | Ошибка объекта INSERT |
| `insertreplace_packfield` | Ошибка упаковки поля |
| `call_proc_preparedb` | Ошибка подготовки процедуры |
| `compare_packfield` | Ошибка сравнения полей |

### Bulk ошибки

| Метрика | Описание |
|---------|----------|
| `bulk_insert_gen` | Ошибка генерации SQL |
| `bulk_insert_preparedb` | Ошибка подготовки соединения |
| `bulk_insert_db` | Ошибка выполнения |
| `bulk_insert_scan` | Ошибка чтения результатов |
| `bulk_insert_match` | Ошибка сопоставления |

## Пример реализации

### Prometheus

```go
package metrics

import (
    "time"
    "github.com/prometheus/client_golang/prometheus"
)

type PrometheusMetrics struct {
    counts   *prometheus.CounterVec
    times    *prometheus.HistogramVec
    errors   *prometheus.CounterVec
}

func NewPrometheusMetrics() *PrometheusMetrics {
    m := &PrometheusMetrics{
        counts: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "activerecord_operations_total",
                Help: "Total number of operations",
            },
            []string{"namespace", "operation"},
        ),
        times: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "activerecord_operation_duration_seconds",
                Help:    "Operation duration in seconds",
                Buckets: prometheus.DefBuckets,
            },
            []string{"namespace", "operation"},
        ),
        errors: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "activerecord_errors_total",
                Help: "Total number of errors",
            },
            []string{"namespace", "error_type"},
        ),
    }

    prometheus.MustRegister(m.counts, m.times, m.errors)
    return m
}

func (m *PrometheusMetrics) StatCount(namespace, name string, cnt int) {
    m.counts.WithLabelValues(namespace, name).Add(float64(cnt))
}

func (m *PrometheusMetrics) StatTime(namespace, operation string, dt time.Duration) {
    m.times.WithLabelValues(namespace, operation).Observe(dt.Seconds())
}

func (m *PrometheusMetrics) StatErr(namespace, name string, cnt int) {
    m.errors.WithLabelValues(namespace, name).Add(float64(cnt))
}
```

### Регистрация

```go
func main() {
    metrics := NewPrometheusMetrics()
    activerecord.RegisterMetrics(metrics)
}
```

## Namespace

Все метрики включают namespace модели:

```
activerecord_operations_total{namespace="users", operation="select_db"} 1234
activerecord_operations_total{namespace="orders", operation="insert_success"} 567
```

## Рекомендации

### 1. Мониторьте ошибки

```promql
rate(activerecord_errors_total[5m]) > 0
```

### 2. Отслеживайте latency

```promql
histogram_quantile(0.99, rate(activerecord_operation_duration_seconds_bucket[5m]))
```

### 3. Настройте алерты

```yaml
- alert: HighSelectLatency
  expr: histogram_quantile(0.99, rate(activerecord_operation_duration_seconds_bucket{operation="select_db"}[5m])) > 0.1
  for: 5m
  labels:
    severity: warning
```

## Следующие шаги

- [Конфигурация](../usage/configuration.md) — регистрация метрик
