package activerecord

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/Educentr/go-activerecord/v3/pkg/logger"
	"github.com/Educentr/go-activerecord/v3/pkg/logger/arsimplelog"
)

var ErrNoData = errors.New("no data")

type SelectorLimiter interface {
	Limit() uint32
	Offset() uint32
	FulfillWarn() bool
	fmt.Stringer
}

type Limiter struct {
	limit, offset uint32
	fulfillWarn   bool
}

func EmptyLimiter() Limiter {
	return Limiter{}
}

func NewLimiter(limit uint32) Limiter {
	return Limiter{limit: limit}
}

func NewLimitOffset(limit uint32, offset uint32) Limiter {
	return Limiter{limit: limit, offset: offset}
}

func NewThreshold(limit uint32) Limiter {
	return Limiter{limit: limit, fulfillWarn: true}
}

func (l Limiter) Offset() uint32 {
	return l.offset
}

func (l Limiter) Limit() uint32 {
	return l.limit
}

func (l Limiter) FulfillWarn() bool {
	return l.fulfillWarn
}

func (l Limiter) String() string {
	return fmt.Sprintf("Limit: %d, Offset: %d, Is Threshold: %t", l.limit, l.offset, l.fulfillWarn)
}

//go:generate mockery --name ConfigInterface --filename mock_config.go --structname MockConfig --with-expecter=true  --inpackage
type ConfigInterface interface {
	GetBool(confPath string, dfl ...bool) (bool, error)
	GetBoolIfExists(confPath string) (value bool, ok bool, err error)
	GetInt(confPath string, dfl ...int64) (int64, error)
	GetIntIfExists(confPath string) (int64, bool, error)
	GetDuration(confPath string, dfl ...time.Duration) (time.Duration, error)
	GetDurationIfExists(confPath string) (time.Duration, bool, error)
	GetString(confPath string, dfl ...string) (string, error)
	GetStringIfExists(confPath string) (string, bool, error)
	GetStrings(confPath string, dfl []string) ([]string, error)
	GetStruct(confPath string, valuePtr interface{}) (bool, error)
}

type ConnectionCacherInterface interface {
	Add(shard ShardInstance, connector func(interface{}) (ConnectionInterface, error)) (ConnectionInterface, error)
	GetOrAdd(shard ShardInstance, connector func(interface{}) (ConnectionInterface, error)) (ConnectionInterface, error)
	Get(shard ShardInstance) ConnectionInterface
	CloseConnection(context.Context)
}

type ClusterCheckerInterface interface {
	AddClusterChecker(ctx context.Context, path string, params ClusterConfigParameters) (*Cluster, error)
}

type ConfigCacherInterface interface {
	Get(ctx context.Context, path string, glob MapGlobParam, optionCreator func(ShardInstanceConfig) (OptionInterface, error)) (*Cluster, error)
}

type SerializerInterface interface {
	Unmarshal(val interface{}) (interface{}, error)
	Marshal(data interface{}) (interface{}, error)
}

type MetricTimerInterface interface {
	Timing(ctx context.Context, name string)
	Finish(ctx context.Context, name string)
}

type MetricStatCountInterface interface {
	Inc(ctx context.Context, name string, val float64)
}

type MetricErrorCountInterface interface {
	Inc(ctx context.Context, name string, val float64)
}

type MetricInterface interface {
	StatCount(storage, entity string) MetricStatCountInterface
	ErrorCount(storage, entity string) MetricStatCountInterface
	Timer(storage, entity string) MetricTimerInterface
}

type ActiveRecord struct {
	instanceCreator  string
	config           func(ctx context.Context) ConfigInterface
	logger           logger.LoggerInterface
	metric           MetricInterface
	connectionCacher ConnectionCacherInterface
	configCacher     ConfigCacherInterface
	pinger           ClusterCheckerInterface
}

var instance *ActiveRecord
var createMutex sync.Mutex

func ReinitActiveRecord(opts ...Option) {
	instance = nil

	InitActiveRecord(opts...)
}

func InitActiveRecord(opts ...Option) {
	createMutex.Lock()
	defer createMutex.Unlock()

	if instance != nil {
		panic(fmt.Sprintf("can't initialize twice, first from `%s`", instance.instanceCreator))
	}

	caller := "unknown_caller"

	_, file, no, ok := runtime.Caller(1)
	if ok {
		caller = fmt.Sprintf("%s:%d ", file, no)
	}

	instance = &ActiveRecord{
		instanceCreator:  caller,
		logger:           arsimplelog.NewLogger(),
		config:           NewDefaultConfig(),
		metric:           NewDefaultNoopMetric(),
		connectionCacher: newConnectionPool(),
		configCacher:     NewConfigCacher(),
	}

	for _, opt := range opts {
		opt.apply(instance)
	}
}

func GetInstance() *ActiveRecord {
	if instance == nil {
		panic("get instance before initialization")
	}

	return instance
}

func Logger() logger.LoggerInterface {
	return GetInstance().logger
}

func Metric() MetricInterface {
	return GetInstance().metric
}

func Config(ctx context.Context) ConfigInterface {
	return GetInstance().config(ctx)
}

func ConnectionCacher() ConnectionCacherInterface {
	return GetInstance().connectionCacher
}

func ConfigCacher() ConfigCacherInterface {
	return GetInstance().configCacher
}

// AddClusterChecker регистрирует конфигурацию кластера в локальном пингере
func AddClusterChecker(ctx context.Context, configPath string, params ClusterConfigParameters) (*Cluster, error) {
	if GetInstance().pinger == nil {
		return nil, fmt.Errorf("connection pinger is not configured. Configure it with function InitActiveRecord and WithConnectionPinger option ")
	}

	return GetInstance().pinger.AddClusterChecker(ctx, configPath, params)
}
