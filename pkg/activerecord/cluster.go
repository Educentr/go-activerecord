package activerecord

import (
	"bytes"
	"context"
	"fmt"
	"hash"
	"hash/crc32"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Интерфейс которому должен соответствовать билдер опций подключения к конретному инстансу
type OptionInterface interface {
	GetConnectionID() string
	InstanceMode() ServerModeType
}

// Тип и константы для выбора инстанса в шарде
type ShardInstanceType uint8

const (
	MasterInstanceType          ShardInstanceType = iota // Любой из описанных мастеров. По умолчанию используется для rw запросов
	ReplicaInstanceType                                  // Любой из описанных реплик
	ReplicaOrMasterInstanceType                          // Любая реплика если есть, если нет то любой мастер. По умолчанию используется при селекте
)

// Тип и константы для определения режима работы конкретного инстанса.
type ServerModeType uint8

const (
	ModeMaster ServerModeType = iota
	ModeReplica
)

// Структура используется для описания конфигурации конктретного инстанса
type ShardInstanceConfig struct {
	Timeout  time.Duration
	Mode     ServerModeType
	PoolSize int32
	Addr     string
	User     string
	Password string
	Port     uint16
	DB       string
}

// Структура описывающая инстанс в кластере
type ShardInstance struct {
	ParamsID string
	Config   ShardInstanceConfig
	Options  interface{}
	Offline  bool
}

// Структура описывающая конкретный шард. Каждый шард может состоять из набора мастеров и реплик
type Shard struct {
	Masters    []ShardInstance
	Replicas   []ShardInstance
	curMaster  int32
	curReplica int32
}

// Функция выбирающая следующий доступный инстанс мастера в конкретном шарде
func (s *Shard) NextMaster() ShardInstance {
	length := len(s.Masters)
	switch length {
	case 0:
		panic("no master configured")
	case 1:
		master := s.Masters[0]
		if master.Offline {
			panic("no available master")
		}

		return master
	}

	// Из-за гонок при большом кол-ве недоступных инстансов может потребоватся много попыток найти доступный узел
	attempt := 10 * length

	for i := 0; i < attempt; i++ {
		newVal := atomic.AddInt32(&s.curMaster, 1)
		lenM := len(s.Masters)
		if lenM > math.MaxInt32 {
			panic("too many masters")
		}

		newValMod := newVal % int32(lenM)

		if newValMod != newVal {
			atomic.CompareAndSwapInt32(&s.curMaster, newVal, newValMod)
		}

		master := s.Masters[newValMod]
		if master.Offline {
			continue
		}

		return master
	}

	//nolint:gosec
	// Есть небольшая вероятность при большой нагрузке и большом проценте недоступных инстансов можно залипнуть на доступном узле
	// Чтобы не паниковать выбираем рандомный узел
	return s.Masters[rand.Int()%length]
}

// Инстанс выбирающий следующий доступный инстанс реплики в конкретном шарде
func (s *Shard) NextReplica() ShardInstance {
	length := len(s.Replicas)
	if length == 1 && !s.Replicas[0].Offline {
		return s.Replicas[0]
	}

	// Из-за гонок при большом кол-ве недоступных инстансов может потребоватся много попыток найти доступный узел
	attempt := 10 * length

	for i := 0; i < attempt; i++ {
		newVal := atomic.AddInt32(&s.curReplica, 1)
		lenR := len(s.Replicas)
		if lenR > math.MaxInt32 {
			panic("too many replicas")
		}

		newValMod := newVal % int32(lenR)

		if newValMod != newVal {
			atomic.CompareAndSwapInt32(&s.curReplica, newVal, newValMod)
		}

		replica := s.Replicas[newValMod]
		if replica.Offline {
			continue
		}

		return replica
	}

	//nolint:gosec
	// Есть небольшая вероятность при большой нагрузке и большом проценте недоступных инстансов поиск может залипнуть на недоступном узле
	// Чтобы не паниковать выбираем рандомный узел
	return s.Replicas[rand.Int()%length]
}

// Instances копия списка конфигураций всех инстансов шарды. В начале списка следуют мастера, потом реплики
func (c *Shard) Instances() []ShardInstance {
	instances := make([]ShardInstance, 0, len(c.Masters)+len(c.Replicas))
	instances = append(instances, c.Masters...)
	instances = append(instances, c.Replicas...)

	return instances
}

// Тип описывающий кластер. Сейчас кластер - это набор шардов.
type Cluster struct {
	m      sync.RWMutex
	shards []Shard
	hash   hash.Hash
}

func NewCluster(shardCnt int) *Cluster {
	return &Cluster{
		m:      sync.RWMutex{},
		shards: make([]Shard, 0, shardCnt),
		hash:   crc32.NewIEEE(),
	}
}

// NextMaster выбирает следующий доступный инстанс мастера в шарде shardNum
func (c *Cluster) NextMaster(shardNum int) ShardInstance {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.shards[shardNum].NextMaster()
}

// NextMaster выбирает следующий доступный инстанс реплики в шарде shardNum
func (c *Cluster) NextReplica(shardNum int) (ShardInstance, bool) {
	c.m.RLock()
	defer c.m.RUnlock()

	for _, replica := range c.shards[shardNum].Replicas {
		if replica.Offline {
			continue
		}

		return c.shards[shardNum].NextReplica(), true
	}

	return ShardInstance{}, false

}

// Append добавляет новый шард в кластер
func (c *Cluster) Append(shard Shard) {
	c.m.Lock()
	defer c.m.Unlock()

	c.shards = append(c.shards, shard)

	c.hash.Reset()
	for i := 0; i < len(c.shards); i++ {
		for _, instance := range c.shards[i].Instances() {
			c.hash.Write([]byte(instance.ParamsID))
		}
	}
}

// ShardInstances копия всех инстансов из шарды shardNum
func (c *Cluster) ShardInstances(shardNum int) []ShardInstance {
	c.m.Lock()
	defer c.m.Unlock()

	return c.shards[shardNum].Instances()
}

// Shards кол-во доступных шард кластера
func (c *Cluster) Shards() int {
	return len(c.shards)
}

// SetShardInstances заменяет инстансы кластера в шарде shardNum на инстансы из instances
func (c *Cluster) SetShardInstances(shardNum int, instances []ShardInstance) {
	c.m.Lock()
	defer c.m.Unlock()

	shard := c.shards[shardNum]
	shard.Masters = shard.Masters[:0]
	shard.Replicas = shard.Replicas[:0]
	for _, shardInstance := range instances {
		switch shardInstance.Config.Mode {
		case ModeMaster:
			shard.Masters = append(shard.Masters, shardInstance)
		case ModeReplica:
			shard.Replicas = append(shard.Replicas, shardInstance)
		}
	}

	c.shards[shardNum] = shard
}

// Equal сравнивает загруженные конфигурации кластеров на основе контрольной суммы всех инстансов кластера
func (c *Cluster) Equal(c2 *Cluster) bool {
	if c == nil {
		return false
	}

	if c2 == nil {
		return false
	}

	return bytes.Equal(c.hash.Sum(nil), c2.hash.Sum(nil))
}

// Тип используемый для передачи набора значений по умолчанию для параметров
type MapGlobParam struct {
	Timeout  time.Duration
	PoolSize int
}

// Конструктор который позволяет проинициализировать новый кластер. В опциях передаются все шарды,
// сколько шардов, столько и опций. Используется в случаях, когда информация по кластеру прописана
// непосредственно в декларации модели, а не в конфиге.
// Так же используется при тестировании.
func NewClusterInfo(opts ...clusterOption) *Cluster {
	cl := NewCluster(0)

	for _, opt := range opts {
		opt.apply(cl)
	}

	return cl
}

// Констркуктор позволяющий проинициализировать кластер их конфигурации.
// На вход передаётся путь в конфиге, значения по умолчанию, и ссылка на функцию, которая
// создаёт структуру опций и считает контрольную сумму, для того, что бы следить за их изменением в онлайне.
func GetClusterInfoFromCfg(ctx context.Context, path string, globs MapGlobParam, optionCreator func(ShardInstanceConfig) (OptionInterface, error)) (*Cluster, error) {
	cfg := Config(ctx)

	shardCnt, exMaxShardEx, err := cfg.GetIntIfExists(path + "/max-shard")
	if err != nil {
		return nil, fmt.Errorf("can't get max-shard: %w", err)
	}

	if !exMaxShardEx {
		shardCnt = 1
	}

	cluster := NewCluster(int(shardCnt))

	globalTimeout, exGlobalTimeout, err := cfg.GetDurationIfExists(path + "/Timeout")
	if err != nil {
		return nil, fmt.Errorf("can't get global timeout: %w", err)
	}

	if exGlobalTimeout {
		globs.Timeout = globalTimeout
	}

	globalPoolSize, exGlobalPoolSize, err := cfg.GetIntIfExists(path + "/PoolSize")
	if err != nil {
		return nil, fmt.Errorf("can't get global pool size: %w", err)
	}

	if !exGlobalPoolSize {
		globalPoolSize = 1
	}

	globs.PoolSize = int(globalPoolSize)

	if exMaxShardEx {
		// Если используется много шардов
		for f := 0; f < int(shardCnt); f++ {
			shard, err := getShardInfoFromCfg(ctx, path+"/"+strconv.Itoa(f), globs, optionCreator)
			if err != nil {
				return nil, fmt.Errorf("can't get shard %d info: %w", f, err)
			}

			cluster.Append(shard)
		}
	} else {
		// Когда только один шард
		shard, err := getShardInfoFromCfg(ctx, path, globs, optionCreator)
		if err != nil {
			return nil, fmt.Errorf("can't get shard info: %w", err)
		}

		cluster.Append(shard)
	}

	return cluster, nil
}

func fillShardConnectionParams(masterDef string) ([]ShardInstanceConfig, error) {
	shards := strings.Split(masterDef, ",") // ToDo check length

	ret := make([]ShardInstanceConfig, 0, len(shards))

	for _, inst := range shards {
		if inst == "" {
			return nil, fmt.Errorf("invalid master instance options: addr is empty")
		}

		hostport := strings.SplitN(inst, ":", 2) // ToDo check
		if len(hostport) != 2 {
			return nil, fmt.Errorf("invalid master instance options: port is empty")
		}

		port, errPort := strconv.Atoi(hostport[1])
		if errPort != nil || port > math.MaxUint16 {
			return nil, fmt.Errorf("invalid port(%s): %w", hostport[1], errPort)
		}

		ret = append(ret, ShardInstanceConfig{
			Addr: hostport[0],
			Port: uint16(port), // ToDo check type conversion
		})
	}

	return ret, nil
}

// Чтение информации по конкретному шарду из конфига
func getShardInfoFromCfg(ctx context.Context, path string, globParam MapGlobParam, optionCreator func(ShardInstanceConfig) (OptionInterface, error)) (Shard, error) {
	cfg := Config(ctx)
	ret := Shard{
		Masters:  []ShardInstance{},
		Replicas: []ShardInstance{},
	}

	shardTimeout, err := cfg.GetDuration(path+"/Timeout", globParam.Timeout)
	if err != nil {
		return Shard{}, fmt.Errorf("can't get timeout: %w", err)
	}

	shardPoolSize, err := cfg.GetInt(path+"/PoolSize", int64(globParam.PoolSize))
	if err != nil {
		return Shard{}, fmt.Errorf("can't get pool size: %w", err)
	}

	// ToDo сделать возможность указать параметры на уровне кластера, если используются идентичные данные в каждом шарде
	// ToDo сделать возможность для каждого шарда указывать свои креды
	shardUserName, err := cfg.GetString(path+"/User", "")
	if err != nil {
		return Shard{}, fmt.Errorf("can't get user: %w", err)
	}

	shardPassword, err := cfg.GetString(path+"/Password", "")
	if err != nil {
		return Shard{}, fmt.Errorf("can't get password: %w", err)
	}

	shardDBName, err := cfg.GetString(path+"/DB", "")
	if err != nil {
		return Shard{}, fmt.Errorf("can't get db: %w", err)
	}

	// ToDo different DB different rules
	// if shardDBName == "" {
	// 	return Shard{}, fmt.Errorf("shard db name should be specified in %s", path+"/DB")
	// }

	// информация по местерам
	master, exMaster, err := cfg.GetStringIfExists(path + "/master")
	if err != nil {
		return Shard{}, fmt.Errorf("can't get master: %w", err)
	}

	if !exMaster {
		master, exMaster, err = cfg.GetStringIfExists(path)
		if !exMaster || err != nil {
			return Shard{}, fmt.Errorf("master should be specified in '%s' or in '%s/master' and replica in '%s/replica'", path, path, path)
		}
	}

	if master != "" {
		shards, errFill := fillShardConnectionParams(master)
		if errFill != nil {
			return Shard{}, fmt.Errorf("can't fill shard connection params: %w", errFill)
		}

		for _, shardCfg := range shards {
			shardCfg.Mode = ModeMaster
			shardCfg.PoolSize = int32(shardPoolSize) // ToDo check type conversion
			shardCfg.Timeout = shardTimeout
			shardCfg.User = shardUserName
			shardCfg.Password = shardPassword
			shardCfg.DB = shardDBName

			opt, errOpt := optionCreator(shardCfg)
			if errOpt != nil {
				return Shard{}, fmt.Errorf("can't create instanceOption: %w", errOpt)
			}

			ret.Masters = append(ret.Masters, ShardInstance{
				ParamsID: opt.GetConnectionID(),
				Config:   shardCfg,
				Options:  opt,
			})
		}
	}

	// Информация по репликам
	replica, exReplica, err := cfg.GetStringIfExists(path + "/replica")
	if err != nil {
		return Shard{}, fmt.Errorf("can't get replica: %w", err)
	}

	if exReplica {
		shards, errFill := fillShardConnectionParams(replica)
		if errFill != nil {
			return Shard{}, fmt.Errorf("can't fill shard connection params: %w", errFill)
		}

		for _, shardCfg := range shards {
			shardCfg.Mode = ModeReplica

			if shardPoolSize > math.MaxInt32 {
				return Shard{}, fmt.Errorf("can't get pool size: %w", err)
			}

			shardCfg.PoolSize = int32(shardPoolSize)
			shardCfg.Timeout = shardTimeout
			shardCfg.User = shardUserName
			shardCfg.Password = shardPassword
			shardCfg.DB = shardDBName

			opt, errOpt := optionCreator(shardCfg)
			if errOpt != nil {
				return Shard{}, fmt.Errorf("can't create instanceOption: %w", errOpt)
			}

			ret.Masters = append(ret.Masters, ShardInstance{
				ParamsID: opt.GetConnectionID(),
				Config:   shardCfg,
				Options:  opt,
			})
		}
	}

	return ret, nil
}

// Структура для кеширования полученных конфигураций. Инвалидация происходит посредством
// сравнения updateTime сданной труктуры и самого конфига.
// Используется для шаринга конфигов между можелями если они используют одну и ту же
// конфигурацию для подключений
type DefaultConfigCacher struct {
	lock             sync.RWMutex
	container        map[string]*Cluster
	updateTime       time.Time
	configUpdateTime time.Time
}

// Конструктор для создания нового кешера конфигов
func NewConfigCacher() *DefaultConfigCacher {
	return &DefaultConfigCacher{
		lock:             sync.RWMutex{},
		container:        make(map[string]*Cluster),
		updateTime:       time.Now(),
		configUpdateTime: time.Now(),
	}
}

// Получение конфигурации. Если есть в кеше и он еще валидный, то конфигурация берётся из кеша
// если в кеше нет, то достаём из конфига и кешируем.
func (cc *DefaultConfigCacher) Get(ctx context.Context, path string, globs MapGlobParam, optionCreator func(ShardInstanceConfig) (OptionInterface, error)) (*Cluster, error) {
	cc.lock.RLock()
	conf, ex := cc.container[path]
	confCacherUpdateTime := cc.updateTime
	confUpdateTime := cc.configUpdateTime
	cc.lock.RUnlock()

	// Если конфигурация не найдена в кеше или конфигурация была обновлена, то перегружаем конфигурацию
	if !ex || confCacherUpdateTime != confUpdateTime {
		cc.lock.Lock()
		newConf, err := GetClusterInfoFromCfg(ctx, path, globs, optionCreator)
		if err != nil {
			cc.lock.Unlock()

			return nil, fmt.Errorf("can't get config: %w", err)
		}

		// если конфигурация поменялась, то обновляем её в кеше
		if !newConf.Equal(conf) {
			conf = newConf
			cc.container[path] = conf
			cc.updateTime = cc.configUpdateTime
		}

		cc.lock.Unlock()
	}

	return conf, nil
}
