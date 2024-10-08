package activerecord

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"
)

type ConnectionInterface interface {
	Close()
	Done() <-chan struct{}
}

type connectionPool struct {
	lock      sync.Mutex
	container map[string]ConnectionInterface
}

func newConnectionPool() *connectionPool {
	return &connectionPool{
		lock:      sync.Mutex{},
		container: make(map[string]ConnectionInterface),
	}
}

// GetConnectionID - получение ConnecitionID. После первого получения, больше нельзя его модифицировать. Можно только новый Options создать
func (o *BaseConnectionOptions) GetConnectionID() string {
	o.Calculated = true
	hashInBytes := o.ConnectionHash.Sum(nil)[:]

	return hex.EncodeToString(hashInBytes)
}

// InstanceMode - метод для получения режима работы инстанса RO или RW
func (o *BaseConnectionOptions) InstanceMode() ServerModeType {
	return ServerModeType(o.Mode)
}

// UpdateHash - функция расчета ConnectionID, необходима для шаринга конектов между моделями.
func (o *BaseConnectionOptions) UpdateHash(data ...interface{}) error {
	if o.Calculated {
		return fmt.Errorf("can't update hash after calculate")
	}

	for _, d := range data {
		var err error

		switch v := d.(type) {
		case string:
			err = binary.Write(o.ConnectionHash, binary.LittleEndian, []byte(v))
		case int:
			err = binary.Write(o.ConnectionHash, binary.LittleEndian, int64(v))
		case nil:
			err = fmt.Errorf("nil data to uprateHash[%+v]", data)
		default:
			err = binary.Write(o.ConnectionHash, binary.LittleEndian, v)
		}

		if err != nil {
			return fmt.Errorf("can't calculate connectionID [%+v]: %w", data, err)
		}
	}

	return nil
}

// TODO при долгом неиспользовании какого то пула надо закрывать его. Это для случаев когда в конфиге поменялась конфигурация
// надо зачищать старые пулы, что бы освободить конекты.
// если будут колбеки о том, что сменилась конфигурация то можно подчищать по этим колбекам.
func (cp *connectionPool) add(shard ShardInstance, connector func(interface{}) (ConnectionInterface, error)) (ConnectionInterface, error) {
	if _, ex := cp.container[shard.ParamsID]; ex {
		return nil, fmt.Errorf("attempt to add duplicate connID: %s", shard.ParamsID)
	}

	pool, err := connector(shard.Options)
	if err != nil {
		return nil, fmt.Errorf("error add connection to shard: %w", err)
	}

	cp.container[shard.ParamsID] = pool

	return pool, nil
}

func (cp *connectionPool) Add(shard ShardInstance, connector func(interface{}) (ConnectionInterface, error)) (ConnectionInterface, error) {
	cp.lock.Lock()
	defer cp.lock.Unlock()

	return cp.add(shard, connector)
}

func (cp *connectionPool) GetOrAdd(shard ShardInstance, connector func(interface{}) (ConnectionInterface, error)) (ConnectionInterface, error) {
	cp.lock.Lock()
	defer cp.lock.Unlock()

	var err error

	conn := cp.Get(shard)
	if conn == nil {
		conn, err = cp.add(shard, connector)
	}

	return conn, err
}

func (cp *connectionPool) Get(shard ShardInstance) ConnectionInterface {
	if conn, ex := cp.container[shard.ParamsID]; ex {
		return conn
	}

	return nil
}

func (cp *connectionPool) CloseConnection(ctx context.Context) {
	cp.lock.Lock()

	for name, pool := range cp.container {
		pool.Close()
		Logger().Debug(ctx, "connection close: %s", name)
	}

	for _, pool := range cp.container {
		<-pool.Done()
		Logger().Debug(ctx, "pool closed done")
	}

	cp.lock.Unlock()
}

// TODO
// - сделать статистику по используемым инстансам
// - прикрутить локальный пингер и исключать недоступные инстансы
func GetConnection(
	ctx context.Context,
	configPath string,
	globParam MapGlobParam,
	optionCreator func(ShardInstanceConfig) (OptionInterface, error),
	instType ShardInstanceType,
	shard int,
	getConnection func(options interface{}) (ConnectionInterface, error),
) (ConnectionInterface, error) {
	clusterInfo, err := ConfigCacher().Get(
		ctx,
		configPath,
		globParam,
		optionCreator,
	)
	if err != nil {
		return nil, fmt.Errorf("can't get cluster %s info: %w", configPath, err)
	}

	if clusterInfo.Shards() < shard {
		return nil, fmt.Errorf("invalid shard num %d, max = %d", shard, clusterInfo.Shards())
	}

	var (
		configBox ShardInstance
		ok        bool
	)

	switch instType {
	case ReplicaInstanceType:
		configBox, ok = clusterInfo.NextReplica(shard)
		if !ok {
			return nil, fmt.Errorf("replicas not set")
		}
	case ReplicaOrMasterInstanceType:
		configBox, ok = clusterInfo.NextReplica(shard)
		if ok {
			break
		}

		fallthrough
	case MasterInstanceType:
		configBox = clusterInfo.NextMaster(shard)
	}

	return ConnectionCacher().GetOrAdd(configBox, getConnection)
}
