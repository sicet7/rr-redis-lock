package redislock

import (
	"context"
	_ "embed"
	"errors"
	"github.com/redis/go-redis/v9"
	lockApi "github.com/roadrunner-server/api/v4/build/lock/v1beta1"
	"go.uber.org/zap"
	"time"
)

//go:embed scripts/exists.lua
var existsScript string

//go:embed scripts/lock.lua
var lockScript string

//go:embed scripts/lockRead.lua
var readLockScript string

//go:embed scripts/release.lua
var releaseScript string

//go:embed scripts/forceRelease.lua
var forceReleaseScript string

//go:embed scripts/updateTtl.lua
var updateTtlScript string

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

func (r *rpc) Lock(req *lockApi.Request, resp *lockApi.Response) error {
	r.log.Debug(
		"lock request received",
		zap.Int("ttl", int(req.GetTtl())),
		zap.Int("wait_ttl", int(req.GetWait())),
		zap.String("resource", req.GetResource()),
		zap.String("id", req.GetId()),
	)

	if !r.plugin.Enabled() {
		return errors.New("service has stopped")
	}

	if req.GetId() == "" {
		return errors.New("empty ID is not allowed")
	}

	c, err := r.plugin.RedisClient()
	if err != nil {
		return err
	}

	timeout := time.Microsecond * time.Duration(req.GetWait())
	if req.GetWait() == int64(0) {
		timeout = time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	retryInterval := r.plugin.cfg.RetryInterval

	script := redis.NewScript(lockScript)

	ttl := req.GetTtl()
	if ttl > 0 {
		ttl = ttl / 1000
	}

	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()

	for {
		cmd := script.Run(ctx, c, []string{req.GetResource()}, req.GetId(), ttl)
		ok, convertErr := cmd.Bool()
		if convertErr != nil {
			if errors.Is(convertErr, context.DeadlineExceeded) {
				resp.Ok = false
				return nil
			}
			return convertErr
		}

		if ok {
			resp.Ok = true
			return nil
		}

		select {
		case <-ctx.Done():
			//timeout or cancel
			resp.Ok = false
			return nil
		case <-ticker.C:
			//retry
		}
	}
}

func (r *rpc) LockRead(req *lockApi.Request, resp *lockApi.Response) error {
	r.log.Debug(
		"read lock request received",
		zap.Int("ttl", int(req.GetTtl())),
		zap.Int("wait_ttl", int(req.GetWait())),
		zap.String("resource", req.GetResource()),
		zap.String("id", req.GetId()),
	)

	if !r.plugin.Enabled() {
		return errors.New("service has stopped")
	}

	if req.GetId() == "" {
		return errors.New("empty ID is not allowed")
	}

	c, err := r.plugin.RedisClient()
	if err != nil {
		return err
	}

	timeout := time.Microsecond * time.Duration(req.GetWait())
	if req.GetWait() == int64(0) {
		timeout = time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	retryInterval := r.plugin.cfg.RetryInterval

	script := redis.NewScript(readLockScript)

	ttl := req.GetTtl()
	if ttl > 0 {
		ttl = ttl / 1000
	}

	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()

	for {
		cmd := script.Run(ctx, c, []string{req.GetResource()}, req.GetId(), ttl)
		ok, convertErr := cmd.Bool()
		if convertErr != nil {
			if errors.Is(convertErr, context.DeadlineExceeded) {
				resp.Ok = false
				return nil
			}
			return convertErr
		}

		if ok {
			resp.Ok = true
			return nil
		}

		select {
		case <-ctx.Done():
			//timeout or cancel
			resp.Ok = false
			return nil
		case <-ticker.C:
			//retry
		}
	}
}

func (r *rpc) Release(req *lockApi.Request, resp *lockApi.Response) error {
	r.log.Debug(
		"release request received",
		zap.Int("ttl", int(req.GetTtl())),
		zap.Int("wait_ttl", int(req.GetWait())),
		zap.String("resource", req.GetResource()),
		zap.String("id", req.GetId()),
	)

	if !r.plugin.Enabled() {
		return errors.New("service has stopped")
	}

	if req.GetId() == "" {
		return errors.New("empty ID is not allowed")
	}

	c, err := r.plugin.RedisClient()
	if err != nil {
		return err
	}

	timeout := time.Microsecond * time.Duration(req.GetWait())
	if req.GetWait() == int64(0) {
		timeout = time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	script := redis.NewScript(releaseScript)

	cmd := script.Run(ctx, c, []string{req.GetResource()}, req.GetId())
	ok, convertErr := cmd.Bool()
	if convertErr != nil {
		if errors.Is(convertErr, context.DeadlineExceeded) {
			resp.Ok = false
			return nil
		}
		return convertErr
	}
	resp.Ok = ok
	return nil
}

func (r *rpc) ForceRelease(req *lockApi.Request, resp *lockApi.Response) error {
	r.log.Debug(
		"force release request received",
		zap.Int("ttl", int(req.GetTtl())),
		zap.Int("wait_ttl", int(req.GetWait())),
		zap.String("resource", req.GetResource()),
		zap.String("id", req.GetId()),
	)

	if !r.plugin.Enabled() {
		return errors.New("service has stopped")
	}

	c, err := r.plugin.RedisClient()
	if err != nil {
		return err
	}

	timeout := time.Microsecond * time.Duration(req.GetWait())
	if req.GetWait() == int64(0) {
		timeout = time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	script := redis.NewScript(forceReleaseScript)

	cmd := script.Run(ctx, c, []string{req.GetResource()})

	ok, convertErr := cmd.Bool()
	if convertErr != nil {
		if errors.Is(convertErr, context.DeadlineExceeded) {
			resp.Ok = false
			return nil
		}
		return convertErr
	}
	resp.Ok = ok
	return nil
}

func (r *rpc) Exists(req *lockApi.Request, resp *lockApi.Response) error {
	r.log.Debug(
		"exists request received",
		zap.Int("ttl", int(req.GetTtl())),
		zap.Int("wait_ttl", int(req.GetWait())),
		zap.String("resource", req.GetResource()),
		zap.String("id", req.GetId()),
	)

	if !r.plugin.Enabled() {
		return errors.New("service has stopped")
	}

	if req.GetId() == "" {
		return errors.New("empty ID is not allowed")
	}

	c, err := r.plugin.RedisClient()
	if err != nil {
		return err
	}

	timeout := time.Microsecond * time.Duration(req.GetWait())
	if req.GetWait() == int64(0) {
		timeout = time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	script := redis.NewScript(existsScript)

	cmd := script.Run(ctx, c, []string{req.GetResource()}, req.GetId())

	ok, convertErr := cmd.Bool()
	if convertErr != nil {
		if errors.Is(convertErr, context.DeadlineExceeded) {
			resp.Ok = false
			return nil
		}
		return convertErr
	}
	resp.Ok = ok
	return nil
}

func (r *rpc) UpdateTTL(req *lockApi.Request, resp *lockApi.Response) error {
	r.log.Debug(
		"exists request received",
		zap.Int("ttl", int(req.GetTtl())),
		zap.Int("wait_ttl", int(req.GetWait())),
		zap.String("resource", req.GetResource()),
		zap.String("id", req.GetId()),
	)

	if !r.plugin.Enabled() {
		return errors.New("service has stopped")
	}

	if req.GetId() == "" {
		return errors.New("empty ID is not allowed")
	}

	c, err := r.plugin.RedisClient()
	if err != nil {
		return err
	}

	timeout := time.Microsecond * time.Duration(req.GetWait())
	if req.GetWait() == int64(0) {
		timeout = time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	script := redis.NewScript(updateTtlScript)

	ttl := req.GetTtl()
	if ttl > 0 {
		ttl = ttl / 1000
	}

	cmd := script.Run(ctx, c, []string{req.GetResource()}, req.GetId(), ttl)

	ok, convertErr := cmd.Bool()
	if convertErr != nil {
		if errors.Is(convertErr, context.DeadlineExceeded) {
			resp.Ok = false
			return nil
		}
		return convertErr
	}
	resp.Ok = ok
	return nil
}
