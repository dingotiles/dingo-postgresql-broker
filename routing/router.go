package routing

import (
	"fmt"
	"regexp"
	"strconv"

	etcd "github.com/coreos/etcd/client"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/pivotal-golang/lager"
	"golang.org/x/net/context"
)

const (
	maxNumberOfRetries = 5
	initialPort        = 30000
	nextPortKey        = "routing/nextport"
)

type Router struct {
	etcd   etcd.KeysAPI
	prefix string
	logger lager.Logger
}

func NewRouter(etcdConfig config.Etcd, logger lager.Logger) (*Router, error) {
	return NewRouterWithPrefix(etcdConfig, "", logger)
}

func NewRouterWithPrefix(etcdConfig config.Etcd, prefix string, logger lager.Logger) (*Router, error) {
	router := &Router{
		prefix: prefix,
		logger: logger,
	}

	var err error
	router.etcd, err = router.setupEtcd(etcdConfig)
	if err != nil {
		return nil, err
	}

	err = router.initializePort()
	if err != nil {
		return nil, err
	}

	return router, nil
}

func (r *Router) AllocatePort() (int, error) {
	r.logger.Info("allocate-port")

	ctx := context.TODO()
	key := fmt.Sprintf("%s/%s", r.prefix, nextPortKey)

	var err error
	for i := 0; i < maxNumberOfRetries; i++ {
		nextPort, err := r.getNextPort(ctx, key)
		if err != nil {
			r.logger.Error("allocate-port.get", err)
			continue
		}

		err = r.increaseNextPort(ctx, key, nextPort)
		if err != nil {
			r.logger.Error("allocate-port.increase", err)
			continue
		}

		return nextPort, nil
	}

	return 0, err
}

func (r *Router) AssignPortToCluster(clusterID string, port int) error {
	r.logger.Info("assign-port-to-cluster", lager.Data{
		"clusterID": clusterID,
		"port":      port,
	})

	ctx := context.TODO()
	key := fmt.Sprintf("%s/routing/allocation/%s", r.prefix, clusterID)
	_, err := r.etcd.Set(ctx, key, fmt.Sprintf("%d", port), &etcd.SetOptions{})
	if err != nil {
		r.logger.Error("assign-port-to-cluster.set", err)
		return err
	}

	return nil
}

func (r *Router) RemoveClusterAssignment(clusterID string) error {
	r.logger.Info("remove-cluster-assignment", lager.Data{
		"clusterID": clusterID,
	})

	ctx := context.TODO()
	key := fmt.Sprintf("%s/routing/allocation/%s", r.prefix, clusterID)

	_, err := r.etcd.Delete(ctx, key, &etcd.DeleteOptions{})
	if err != nil {
		r.logger.Error("remove-cluster-assignment.delete", err)
		return err
	}

	return nil
}

func (r *Router) setupEtcd(cfg config.Etcd) (etcd.KeysAPI, error) {
	client, err := etcd.New(etcd.Config{Endpoints: cfg.Machines})
	if err != nil {
		return nil, err
	}

	api := etcd.NewKeysAPI(client)

	return api, nil
}

func (r *Router) initializePort() error {
	ctx := context.TODO()
	key := fmt.Sprintf("%s/%s", r.prefix, nextPortKey)

	r.logger.Info("initialize-port", lager.Data{"key": nextPortKey})

	_, err := r.getNextPort(ctx, key)

	if err != nil {
		// if the key wasn't found etcd is available
		// but routing hasn't been initialized
		notFoundRegExp, _ := regexp.Compile("Key not found")
		if notFoundRegExp.FindString(err.Error()) == "Key not found" {

			_, err := r.etcd.Set(ctx, key, fmt.Sprintf("%d", initialPort), &etcd.SetOptions{
				PrevExist: etcd.PrevNoExist,
			})
			if err != nil {
				r.logger.Error("initialize-port.set-value", err)
				return err
			}
		} else {
			r.logger.Error("initialize-port.get", err)
			return err
		}
	}

	return nil
}

func (r *Router) getNextPort(ctx context.Context, key string) (int, error) {
	resp, err := r.etcd.Get(ctx, key, &etcd.GetOptions{Quorum: true})
	if err != nil {
		return 0, err
	}

	port, err := strconv.Atoi(resp.Node.Value)
	if err != nil {
		return 0, err
	}

	return port, nil
}

func (r *Router) increaseNextPort(ctx context.Context, key string, current int) error {
	_, err := r.etcd.Set(ctx, key, fmt.Sprintf("%d", current+1), &etcd.SetOptions{
		PrevValue: fmt.Sprintf("%d", current),
		PrevExist: etcd.PrevExist,
	})

	return err
}
