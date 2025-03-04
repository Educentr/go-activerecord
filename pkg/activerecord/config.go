package activerecord

import (
	"context"
	"fmt"
	"time"
)

type DefaultConfig struct {
	cfg     map[string]interface{}
	created time.Time
}

func NewDefaultConfig() func(ctx context.Context) ConfigInterface {
	return func(ctx context.Context) ConfigInterface {
		return &DefaultConfig{
			cfg: make(map[string]interface{}),
		}
	}
}

func NewDefaultConfigFromMap(cfg map[string]interface{}) *DefaultConfig {
	return &DefaultConfig{
		cfg:     cfg,
		created: time.Now(),
	}
}

func (dc *DefaultConfig) GetLastUpdateTime() time.Time {
	return dc.created
}

func (dc *DefaultConfig) GetBool(confPath string, dfl ...bool) (bool, error) {
	ret, ex, err := dc.GetBoolIfExists(confPath)
	if err != nil {
		return false, err
	}

	if ex {
		return ret, nil
	}

	if len(dfl) != 0 {
		return dfl[0], nil
	}

	return false, nil
}

func (dc *DefaultConfig) GetBoolIfExists(confPath string) (value bool, ok bool, err error) {
	if param, ex := dc.cfg[confPath]; ex {
		if ret, ok := param.(bool); ok {
			return ret, true, nil
		}

		return false, false, fmt.Errorf("param %s has type %T, want bool", confPath, param)
	}

	return false, false, nil
}

func (dc *DefaultConfig) GetInt(confPath string, dfl ...int64) (int64, error) {
	ret, ok, err := dc.GetIntIfExists(confPath)
	if err != nil {
		return 0, err
	}

	if ok {
		return ret, nil
	}

	if len(dfl) != 0 {
		return dfl[0], nil
	}

	return 0, nil
}

func (dc *DefaultConfig) GetIntIfExists(confPath string) (int64, bool, error) {
	if param, ex := dc.cfg[confPath]; ex {
		if ret, ok := param.(int64); ok {
			return ret, true, nil
		}

		return 0, false, fmt.Errorf("param %s has type %T, want int", confPath, param)
	}

	return 0, false, nil
}

func (dc *DefaultConfig) GetDuration(confPath string, dfl ...time.Duration) (time.Duration, error) {
	ret, ok, err := dc.GetDurationIfExists(confPath)
	if err != nil {
		return 0, nil
	}

	if ok {
		return ret, nil
	}

	if len(dfl) != 0 {
		return dfl[0], nil
	}

	return 0, nil
}

func (dc *DefaultConfig) GetDurationIfExists(confPath string) (time.Duration, bool, error) {
	if param, ex := dc.cfg[confPath]; ex {
		if ret, ok := param.(time.Duration); ok {
			return ret, true, nil
		}

		return 0, false, fmt.Errorf("param %s has type %T, want time.Duration", confPath, param)
	}

	return 0, false, nil
}

func (dc *DefaultConfig) GetString(confPath string, dfl ...string) (string, error) {
	ret, ok, err := dc.GetStringIfExists(confPath)
	if err != nil {
		return "", err
	}

	if ok {
		return ret, nil
	}

	if len(dfl) != 0 {
		return dfl[0], nil
	}

	return "", nil
}

func (dc *DefaultConfig) GetStringIfExists(confPath string) (string, bool, error) {
	if param, ex := dc.cfg[confPath]; ex {
		if ret, ok := param.(string); ok {
			return ret, true, nil
		}

		return "", false, fmt.Errorf("param %s has type %T, want string", confPath, param)
	}

	return "", false, nil
}

func (dc *DefaultConfig) GetStrings(confPath string, dfl []string) ([]string, error) {
	return []string{}, fmt.Errorf("GetStrings not implemented")
}

func (dc *DefaultConfig) GetStruct(confPath string, valuePtr interface{}) (bool, error) {
	return false, fmt.Errorf("GetStruct not implemented")
}
