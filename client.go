package eureka

import (
	"context"
	"strconv"
	"time"

	"github.com/ArthurHlt/go-eureka-client/eureka"
)

const (
	defaultTTL = 5 * time.Second
	heartBeat  = 10 * time.Second
)

type Client interface {
	GetEntries(appId string) ([]string, error)
	Register(appId string, ip string, port int) error
}

type client struct {
	eurekaClient *eureka.Client
	ctx          context.Context
}

type ClientOptions struct {
	CertFile    string
	KeyFile     string
	CaCertFile  []string
	DialTimeout time.Duration
	Consistency string
}

func NewClient(ctx context.Context, machines []string, options ClientOptions) (Client, error) {

	if options.DialTimeout == 0 {
		options.DialTimeout = defaultTTL
	}

	return &client{ctx: ctx, eurekaClient: eureka.NewClient(machines)}, nil
}

func (c *client) GetEntries(appId string) ([]string, error) {
	appIdNoHttp := appId[7:]
	resp, err := c.eurekaClient.GetApplication(appIdNoHttp)
	entries := make([]string, len(resp.Instances))

	for i, instance := range resp.Instances {
		if instance.Status == "UP" {
			entries[i] = instance.HomePageUrl
		}
	}

	if err != nil {
		return nil, err
	}

	return entries, nil
}

func (c *client) Register(appId string, ip string, port int) error {

	instanceId := ip + ":" + appId + ":" + strconv.Itoa(port)
	instance := eureka.NewInstanceInfo(instanceId, appId, ip, port, 30, false)
	err := c.eurekaClient.RegisterInstance(appId, instance)
	if err != nil {
		return err
	}

	go c.loop(instance)

	return nil
}

func (c *client) loop(instanceInfo *eureka.InstanceInfo) {
	t := time.NewTicker(heartBeat)
	for {
		select {
		case <-t.C:
			c.eurekaClient.SendHeartbeat(instanceInfo.App, instanceInfo.HostName)
		case <-c.ctx.Done():
			c.eurekaClient.UnregisterInstance(instanceInfo.App, instanceInfo.HostName)
			return
		}

	}
}
