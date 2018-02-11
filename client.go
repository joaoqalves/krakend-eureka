package eureka

import (
	"github.com/ArthurHlt/go-eureka-client/eureka"
	"context"
	"time"
)

const defaultTTL = 5 * time.Second

type Client interface {
	GetEntries(appId string) ([]string, error)
}

type client struct {
  	eurekaClient *eureka.Client
	ctx     context.Context
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

	return 	&client {ctx: ctx, eurekaClient: eureka.NewClient(machines)}, nil
}

func (c *client) GetEntries(appId string) ([]string, error) {
	appIdNoHttp := appId[7:]
	resp, err := c.eurekaClient.GetApplication(appIdNoHttp)
	entries := make([]string, len(resp.Instances))

	for i, instance := range resp.Instances {
		entries[i] = instance.HomePageUrl
	}

	if err != nil {
		return nil, err
	}

	return entries, nil
}