package grpclib

import (
	"fmt"
	"math/rand"
	"time"

	etcd3 "github.com/coreos/etcd/clientv3"
)

// ServerRegister 服务发现接口
type ServerRegister interface {
	// Regist 注册服务
	Regist() error
	// Revoke 注销服务
	Revoke() error
}

// DefaultServerRegister 默认发现服务说明
type DefaultServerRegister struct {
	id string
	// prefix should start and end with no slash
	prefix string
	target string
	name   string

	ttl      int
	interval time.Duration
	// invoke self-register with ticker
	ticker *time.Ticker

	service string // host & port
	path    string // etcd path

	client *etcd3.Client
	// 如果出错，那么将交由UnRegister 处理
	stopSignal chan bool
}

type RegisterConfig struct {
	Name     string
	Target   string
	Server   string
	Version  string
	Interval time.Duration
	TTL      int
}

const prefix = "etcd_naming"

// NewDefaultServerRegister 生成服务对象
// name 服务名称
// target etcd目标地址，多个以逗号隔开
// serv 服务的地址 host:port
// interal 轮训时间
// ttl 失效时间
func NewDefaultServerRegister(
	name, target, serv string, interval time.Duration, ttl int) ServerRegister {
	rand.Seed(time.Now().Unix())
	p := &DefaultServerRegister{
		id:      fmt.Sprintf("%d-%d", time.Now().Unix(), rand.Intn(10000)),
		prefix:  prefix,
		target:  target,
		name:    name,
		service: serv,

		ttl:      ttl,
		interval: interval,
		ticker:   time.NewTicker(interval),

		stopSignal: make(chan bool, 1),
	}
	p.path = fmt.Sprintf("/%s/%s/%s", p.prefix, p.name, p.service)

	return p
}

func NewDefaultServerRegisterWithConfig(c *RegisterConfig) ServerRegister {
	return NewDefaultServerRegister(
		fmt.Sprintf("%s-%s", c.Name, c.Version), c.Target, c.Server, c.Interval, c.TTL)
}
