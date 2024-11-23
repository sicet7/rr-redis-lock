package redislock

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
	"redislock/client"
	"time"
)

const PluginName string = "redislock"

type Config struct {
	RetryInterval time.Duration  `mapstructure:"retry_interval"`
	ClientConfig  *client.Config `mapstructure:"client"`
}

type Plugin struct {
	enabled     bool
	log         *zap.Logger
	redisClient redis.UniversalClient
	cfg         *Config
}

type Configurer interface {
	// UnmarshalKey takes a single key and unmarshal it into a Struct.
	UnmarshalKey(name string, out any) error
	// Has checks if a config section exists.
	Has(name string) bool
}

type Logger interface {
	NamedLogger(name string) *zap.Logger
}

func (p *Plugin) Init(cfg Configurer, log Logger) error {
	const op = errors.Op("redislock_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &p.cfg)
	if err != nil {
		return errors.E(op, err)
	}

	if p.cfg.ClientConfig == nil {
		p.cfg.ClientConfig = &client.Config{}
		p.cfg.ClientConfig.InitDefaults()
	}

	if p.cfg.RetryInterval == 0 {
		p.cfg.RetryInterval = 50 * time.Millisecond
	}

	p.log = log.NamedLogger(PluginName)
	p.enabled = true
	return nil
}

func (p *Plugin) Serve() chan error {
	p.enabled = true
	return make(chan error, 1)
}

func (p *Plugin) Stop(ctx context.Context) error {
	p.enabled = false
	if p.redisClient != nil {
		_ = p.redisClient.Close()
		p.redisClient = nil
	}
	return nil
}

func (p *Plugin) Enabled() bool {
	return p.enabled
}

func (p *Plugin) RedisClient() (redis.UniversalClient, error) {
	if p.redisClient == nil {
		c, err := client.NewRedisDriver(p.log, p.cfg.ClientConfig)
		if err != nil {
			return nil, err
		}
		p.redisClient = c
	}
	return p.redisClient, nil
}

func (p *Plugin) Name() string {
	return PluginName
}
