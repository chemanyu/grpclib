package grpclib

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
)

// Regist 注册服务
func (p *DefaultServerRegister) Regist() (err error) {

	// get endpoints for register dial address

	if p.client, err = clientv3.New(clientv3.Config{
		Endpoints: strings.Split(p.target, ","),
	}); err != nil {
		return fmt.Errorf("grpclib: create client failed: %v", err)
	}

	go func() {
		for {
			if err = p.regist(); err != nil {
				fmt.Println(err)
			}
			select {
			case <-p.stopSignal:
				return
			case <-p.ticker.C:
			}
		}
	}()

	return
}

func (p *DefaultServerRegister) regist() (err error) {
	// minimum lease TTL is ttl-second
	ctxGrant, cancelGrant := context.WithTimeout(context.TODO(), p.interval)
	defer cancelGrant()
	resp, ie := p.client.Grant(ctxGrant, int64(p.ttl))
	if ie != nil {
		return fmt.Errorf("grpclib: set service %q with ttl to client failed: %s", p.name, ie.Error())
	}

	ctxGet, cancelGet := context.WithTimeout(context.Background(), p.interval)
	defer cancelGet()
	_, err = p.client.Get(ctxGet, p.path)
	// should get first, if not exist, set it
	if err != nil {
		if err != rpctypes.ErrKeyNotFound {
			return fmt.Errorf("grpclib: set service %q with ttl to client failed: %s", p.name, err.Error())
		}
		ctxPut, cancelPut := context.WithTimeout(context.TODO(), p.interval)
		defer cancelPut()
		if _, err = p.client.Put(ctxPut, p.path, p.service, clientv3.WithLease(resp.ID)); err != nil {
			return fmt.Errorf("grpclib: set service %q with ttl to client failed: %s", p.name, err.Error())
		}
		return
	}

	ctxPut, cancelPut := context.WithTimeout(context.Background(), p.interval)
	defer cancelPut()
	// refresh set to true for not notifying the watcher
	if _, err = p.client.Put(ctxPut, p.path, p.service, clientv3.WithLease(resp.ID)); err != nil {
		return fmt.Errorf("grpclib: refresh service %q with ttl to client failed: %s", p.name, err.Error())
	}
	return
}

// Revoke delete registered service from etcd
func (p *DefaultServerRegister) Revoke() error {
	if p.client == nil {
		return nil
	}
	p.stopSignal <- true
	ctx, cancel := context.WithTimeout(context.Background(), p.interval)
	defer func() {
		cancel()
	}()
	defer func() {
		p.client.Close()
	}()
	_, err := p.client.Delete(ctx, p.path)
	return err
}
