package postgres

import (
	"fmt"
	"hash/crc32"
	"time"

	"github.com/Educentr/go-activerecord/pkg/activerecord"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Константы определяющие дефолтное поведение конектора к postgres-у
const (
	DefaultTimeout           = 50 * time.Millisecond
	DefaultConnectionTimeout = 250 * time.Millisecond
	DefaultRedialInterval    = 100 * time.Millisecond
	DefaultPingInterval      = 1 * time.Second
	DefaultPoolSize          = 5
)

// Используется для подсчета connectionID
var crc32table = crc32.MakeTable(0x4C11DB7)

// ConnectionOptions - опции используемые для подключения
type ConnectionOptions struct {
	activerecord.BaseConnectionOptions
	poolCfg pgxpool.Config
}

// NewConnectionOptions - создание структуры с опциями и дефолтными значениями. Для модификации значений по умолчанию,
// надо передавать опции в конструктор
func NewConnectionOptions(server string, port uint16, mode activerecord.ServerModeType, opts ...ConnectionOption) (*ConnectionOptions, error) {
	if server == "" {
		return nil, fmt.Errorf("invalid param: server is empty")
	}

	pgxConf, err := pgxpool.ParseConfig("")
	if err != nil {
		return nil, fmt.Errorf("can't parse config: %w", err)
	}

	pgxConf.ConnConfig.Host = server
	pgxConf.ConnConfig.Port = port
	pgxConf.ConnConfig.ConnectTimeout = DefaultConnectionTimeout
	pgxConf.MaxConns = DefaultPoolSize

	postgresOpts := &ConnectionOptions{
		poolCfg: *pgxConf,
	}

	postgresOpts.BaseConnectionOptions.ConnectionHash = crc32.New(crc32table)
	postgresOpts.BaseConnectionOptions.Mode = mode

	for _, opt := range opts {
		if err := opt.apply(postgresOpts); err != nil {
			return nil, fmt.Errorf("error apply options: %w", err)
		}
	}

	err = postgresOpts.UpdateHash("S", server, port)
	if err != nil {
		return nil, fmt.Errorf("can't get pool: %w", err)
	}

	return postgresOpts, nil
}

// ConnectionOption - интерфейс которому должны соответствовать опции передаваемые в конструктор
type ConnectionOption interface {
	apply(*ConnectionOptions) error
}

type optionConnectionFunc func(*ConnectionOptions) error

func (o optionConnectionFunc) apply(c *ConnectionOptions) error {
	return o(c)
}

// WithDatabase - имя базы данных
func WithDatabase(dbName string) ConnectionOption {
	return optionConnectionFunc(func(postgresCfg *ConnectionOptions) error {
		postgresCfg.poolCfg.ConnConfig.Database = dbName

		return postgresCfg.UpdateHash("D", dbName)
	})
}

// WithCredentials - если используется авторизация на уровне СУБД
func WithCredentials(username, password string) ConnectionOption {
	return optionConnectionFunc(func(postgresCfg *ConnectionOptions) error {
		postgresCfg.poolCfg.ConnConfig.Password = password
		postgresCfg.poolCfg.ConnConfig.User = username

		return postgresCfg.UpdateHash("C", username, password)
	})
}

// WithTimeout - опция для изменений таймаутов
func WithTimeout(connection time.Duration) ConnectionOption {
	return optionConnectionFunc(func(postgresCfg *ConnectionOptions) error {
		postgresCfg.poolCfg.ConnConfig.ConnectTimeout = connection

		return postgresCfg.UpdateHash("T", connection)
	})
}

// WithPoolSize - опция для изменения размера пулла подключений
func WithPoolSize(size int32) ConnectionOption {
	return optionConnectionFunc(func(octopusCfg *ConnectionOptions) error {
		octopusCfg.poolCfg.MaxConns = size

		return octopusCfg.UpdateHash("s", size)
	})
}

// ToDo mock server for postgres
// //go:generate mockery --name MockServerLogger --with-expecter=true --inpackage
// type MockServerLogger interface {
// 	Debug(fmt string, args ...any)
// 	DebugSelectRequest(ns uint32, indexnum uint32, offset uint32, limit uint32, keys [][][]byte, fixtures ...SelectMockFixture)
// 	DebugUpdateRequest(ns uint32, primaryKey [][]byte, updateOps []Ops, fixtures ...UpdateMockFixture)
// 	DebugInsertRequest(ns uint32, needRetVal bool, insertMode InsertMode, tuple TupleData, fixtures ...InsertMockFixture)
// 	DebugDeleteRequest(ns uint32, primaryKey [][]byte, fixtures ...DeleteMockFixture)
// 	DebugCallRequest(procName string, args [][]byte, fixtures ...CallMockFixture)
// }

// type MockServerOption interface {
// 	apply(*MockServer) error
// }

// type optionFunc func(*MockServer) error

// func (o optionFunc) apply(c *MockServer) error {
// 	return o(c)
// }

// // WithHost - опция для изменения сервера в конфиге
// func WithHost(host, port string) MockServerOption {
// 	return optionFunc(func(oms *MockServer) error {
// 		oms.host = host
// 		oms.port = port
// 		return nil
// 	})
// }

// func WithLogger(logger MockServerLogger) MockServerOption {
// 	return optionFunc(func(oms *MockServer) error {
// 		oms.logger = logger
// 		return nil
// 	})
// }

// func WithIprotoLogger(logger iproto.Logger) MockServerOption {
// 	return optionFunc(func(oms *MockServer) error {
// 		oms.iprotoLogger = logger
// 		return nil
// 	})
// }
