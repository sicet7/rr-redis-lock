package redislock

import "go.uber.org/zap"

type rpc struct {
	plugin *Plugin
	log    *zap.Logger
}

func (p *Plugin) RPC() any {
	return &rpc{
		plugin: p,
		log:    p.log,
	}
}

func (s *rpc) Hello(input string, output *string) error {
	*output = input
	// s.plugin.Foo() <-- you may also use methods from the Plugin itself
	s.log.Info("foo")
	return nil
}
