package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Service 服务信息
type Service struct {
	ID       string
	Name     string
	Address  string
	Port     int
	Tags     []string
	Metadata map[string]string
}

// Discovery 服务发现接口
type Discovery interface {
	Register(ctx context.Context, service *Service) error
	Deregister(ctx context.Context, serviceID string) error
	GetService(ctx context.Context, name string) ([]*Service, error)
	Watch(ctx context.Context, name string) (<-chan []*Service, error)
}

// ConsulDiscovery Consul服务发现实现
type ConsulDiscovery struct {
	client *api.Client
}

// NewConsulDiscovery 创建Consul服务发现
func NewConsulDiscovery(addr string) (*ConsulDiscovery, error) {
	config := api.DefaultConfig()
	config.Address = addr

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %v", err)
	}

	return &ConsulDiscovery{
		client: client,
	}, nil
}

// Register 注册服务
func (d *ConsulDiscovery) Register(ctx context.Context, service *Service) error {
	registration := &api.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Address: service.Address,
		Port:    service.Port,
		Tags:    service.Tags,
		Meta:    service.Metadata,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", service.Address, service.Port),
			Interval: "10s",
			Timeout:  "5s",
		},
	}

	return d.client.Agent().ServiceRegister(registration)
}

// Deregister 注销服务
func (d *ConsulDiscovery) Deregister(ctx context.Context, serviceID string) error {
	return d.client.Agent().ServiceDeregister(serviceID)
}

// GetService 获取服务
func (d *ConsulDiscovery) GetService(ctx context.Context, name string) ([]*Service, error) {
	services, _, err := d.client.Health().Service(name, "", true, nil)
	if err != nil {
		return nil, err
	}

	result := make([]*Service, 0, len(services))
	for _, s := range services {
		result = append(result, &Service{
			ID:      s.Service.ID,
			Name:    s.Service.Service,
			Address: s.Service.Address,
			Port:    s.Service.Port,
			Tags:    s.Service.Tags,
			Metadata: map[string]string{
				"status": s.Checks.AggregatedStatus(),
			},
		})
	}

	return result, nil
}

// Watch 监听服务变化
func (d *ConsulDiscovery) Watch(ctx context.Context, name string) (<-chan []*Service, error) {
	ch := make(chan []*Service)
	go func() {
		defer close(ch)

		lastIndex := uint64(0)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				services, meta, err := d.client.Health().Service(name, "", true, &api.QueryOptions{
					WaitIndex: lastIndex,
				})
				if err != nil {
					continue
				}

				if meta.LastIndex > lastIndex {
					lastIndex = meta.LastIndex
					result := make([]*Service, 0, len(services))
					for _, s := range services {
						result = append(result, &Service{
							ID:      s.Service.ID,
							Name:    s.Service.Service,
							Address: s.Service.Address,
							Port:    s.Service.Port,
							Tags:    s.Service.Tags,
							Metadata: map[string]string{
								"status": s.Checks.AggregatedStatus(),
							},
						})
					}
					ch <- result
				}
			}
		}
	}()

	return ch, nil
}

// EtcdDiscovery etcd服务发现实现
type EtcdDiscovery struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

// NewEtcdDiscovery 创建etcd服务发现
func NewEtcdDiscovery(endpoints []string) (*EtcdDiscovery, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %v", err)
	}

	return &EtcdDiscovery{
		client: client,
		kv:     clientv3.NewKV(client),
		lease:  clientv3.NewLease(client),
	}, nil
}

// Register 注册服务
func (d *EtcdDiscovery) Register(ctx context.Context, service *Service) error {
	key := fmt.Sprintf("/services/%s/%s", service.Name, service.ID)
	value := fmt.Sprintf("%s:%d", service.Address, service.Port)

	// 创建租约
	lease, err := d.lease.Grant(ctx, 10)
	if err != nil {
		return err
	}

	// 注册服务并绑定租约
	_, err = d.kv.Put(ctx, key, value, clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}

	// 自动续租
	keepAliveCh, err := d.lease.KeepAlive(ctx, lease.ID)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-keepAliveCh:
				if !ok {
					return
				}
			}
		}
	}()

	return nil
}

// Deregister 注销服务
func (d *EtcdDiscovery) Deregister(ctx context.Context, serviceID string) error {
	key := fmt.Sprintf("/services/%s", serviceID)
	_, err := d.kv.Delete(ctx, key)
	return err
}

// GetService 获取服务
func (d *EtcdDiscovery) GetService(ctx context.Context, name string) ([]*Service, error) {
	key := fmt.Sprintf("/services/%s", name)
	resp, err := d.kv.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	services := make([]*Service, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		services = append(services, &Service{
			ID:      string(kv.Key),
			Name:    name,
			Address: string(kv.Value),
		})
	}

	return services, nil
}

// Watch 监听服务变化
func (d *EtcdDiscovery) Watch(ctx context.Context, name string) (<-chan []*Service, error) {
	ch := make(chan []*Service)
	key := fmt.Sprintf("/services/%s", name)

	go func() {
		defer close(ch)
		watchCh := d.client.Watch(ctx, key, clientv3.WithPrefix())
		for {
			select {
			case <-ctx.Done():
				return
			case resp := <-watchCh:
				services := make([]*Service, 0)
				for _, ev := range resp.Events {
					if ev.Type == clientv3.EventTypePut {
						services = append(services, &Service{
							ID:      string(ev.Kv.Key),
							Name:    name,
							Address: string(ev.Kv.Value),
						})
					}
				}
				if len(services) > 0 {
					ch <- services
				}
			}
		}
	}()

	return ch, nil
}
