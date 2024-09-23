package octopus

import (
	"fmt"
	"hash/crc32"
	"time"

	"github.com/mailru/activerecord/pkg/activerecord"
	"github.com/mailru/activerecord/pkg/iproto/iproto"
)

// Константы определяющие дефолтное поведение конектора к octopus-у
const (
	DefaultTimeout           = 20 * time.Millisecond
	DefaultConnectionTimeout = 20 * time.Millisecond
	DefaultRedialInterval    = 50 * time.Millisecond
	DefaultPingInterval      = 1 * time.Second
	DefaultPoolSize          = 1
)

// Используется для подсчета connectionID
var crc32table = crc32.MakeTable(0x4C11DB7)

// ConnectionOptions - опции используемые для подключения
type ConnectionOptions struct {
	activerecord.BaseConnectionOptions
	server  string
	poolCfg *iproto.PoolConfig
}

// NewConnectionOptions - cоздание структуры с опциями и дефолтными значениями. Для мидификации значений по умолчанию,
// надо передавать опции в конструктор
func NewConnectionOptions(server string, port uint16, mode activerecord.ServerModeType, opts ...ConnectionOption) (*ConnectionOptions, error) {
	if server == "" {
		return nil, fmt.Errorf("invalid param: server is empty")
	}

	octopusOpts := &ConnectionOptions{
		server: fmt.Sprintf("%s:%d", server, port),
		poolCfg: &iproto.PoolConfig{
			Size:              DefaultPoolSize,
			ConnectTimeout:    DefaultConnectionTimeout,
			DialTimeout:       DefaultConnectionTimeout,
			RedialInterval:    DefaultRedialInterval,
			MaxRedialInterval: DefaultRedialInterval,
			ChannelConfig: &iproto.ChannelConfig{
				WriteTimeout:   DefaultTimeout,
				RequestTimeout: DefaultTimeout,
				PingInterval:   DefaultPingInterval,
			},
		},
	}

	octopusOpts.BaseConnectionOptions.Mode = mode
	octopusOpts.BaseConnectionOptions.ConnectionHash = crc32.New(crc32table)

	for _, opt := range opts {
		if err := opt.apply(octopusOpts); err != nil {
			return nil, fmt.Errorf("error apply options: %w", err)
		}
	}

	err := octopusOpts.UpdateHash("S", server)
	if err != nil {
		return nil, fmt.Errorf("can't get pool: %w", err)
	}

	return octopusOpts, nil
}

// ConnectionOption - интерфейс которому должны соответствовать опции передаваемые в конструктор
type ConnectionOption interface {
	apply(*ConnectionOptions) error
}

type optionConnectionFunc func(*ConnectionOptions) error

func (o optionConnectionFunc) apply(c *ConnectionOptions) error {
	return o(c)
}

// WithTimeout - опция для изменений таймаутов
func WithTimeout(request, connection time.Duration) ConnectionOption {
	return optionConnectionFunc(func(octopusCfg *ConnectionOptions) error {
		octopusCfg.poolCfg.ConnectTimeout = connection
		octopusCfg.poolCfg.DialTimeout = connection
		octopusCfg.poolCfg.ChannelConfig.WriteTimeout = request
		octopusCfg.poolCfg.ChannelConfig.RequestTimeout = request

		return octopusCfg.UpdateHash("T", request, connection)
	})
}

// WithIntervals - опция для изменения интервалов
func WithIntervals(redial, maxRedial, ping time.Duration) ConnectionOption {
	return optionConnectionFunc(func(octopusCfg *ConnectionOptions) error {
		octopusCfg.poolCfg.RedialInterval = redial
		octopusCfg.poolCfg.MaxRedialInterval = maxRedial
		octopusCfg.poolCfg.ChannelConfig.PingInterval = ping

		return octopusCfg.UpdateHash("I", redial, maxRedial, ping)
	})
}

// WithPoolSize - опция для изменения размера пулла подключений
func WithPoolSize(size int32) ConnectionOption {
	return optionConnectionFunc(func(octopusCfg *ConnectionOptions) error {
		octopusCfg.poolCfg.Size = int(size) // ToDo check type conversion

		return octopusCfg.UpdateHash("s", size)
	})
}

// WithPoolLogger - опция для логера конекшен пула
func WithPoolLogger(logger iproto.Logger) ConnectionOption {
	return optionConnectionFunc(func(octopusCfg *ConnectionOptions) error {
		octopusCfg.poolCfg.Logger = logger
		octopusCfg.poolCfg.ChannelConfig.Logger = logger

		return octopusCfg.UpdateHash("L", logger)
	})
}

//go:generate mockery --name MockServerLogger --with-expecter=true  --inpackage
type MockServerLogger interface {
	Debug(fmt string, args ...any)
	DebugSelectRequest(ns uint32, indexnum uint32, offset uint32, limit uint32, keys [][][]byte, fixtures ...SelectMockFixture)
	DebugUpdateRequest(ns uint32, primaryKey [][]byte, updateOps []Ops, fixtures ...UpdateMockFixture)
	DebugInsertRequest(ns uint32, needRetVal bool, insertMode InsertMode, tuple TupleData, fixtures ...InsertMockFixture)
	DebugDeleteRequest(ns uint32, primaryKey [][]byte, fixtures ...DeleteMockFixture)
	DebugCallRequest(procName string, args [][]byte, fixtures ...CallMockFixture)
}

type MockServerOption interface {
	apply(*MockServer) error
}

type optionFunc func(*MockServer) error

func (o optionFunc) apply(c *MockServer) error {
	return o(c)
}

// WithHost - опция для изменения сервера в конфиге
func WithHost(host, port string) MockServerOption {
	return optionFunc(func(oms *MockServer) error {
		oms.host = host
		oms.port = port
		return nil
	})
}

func WithLogger(logger MockServerLogger) MockServerOption {
	return optionFunc(func(oms *MockServer) error {
		oms.logger = logger
		return nil
	})
}

func WithIprotoLogger(logger iproto.Logger) MockServerOption {
	return optionFunc(func(oms *MockServer) error {
		oms.iprotoLogger = logger
		return nil
	})
}
