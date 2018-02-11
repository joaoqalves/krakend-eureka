package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/devopsfaith/krakend-viper"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
	krakendgin "github.com/devopsfaith/krakend/router/gin"
	"github.com/joaoqalves/krakend-eureka"
	"os/signal"
	"syscall"
	"net"
)

func main() {
	port := flag.Int("p", 0, "Port of the service")
	logLevel := flag.String("l", "ERROR", "Logging level")
	debug := flag.Bool("d", false, "Enable the debug")
	configFile := flag.String("c", "/etc/krakend/configuration.json",
		"Path to the configuration filename")
	flag.Parse()

	parser := viper.New()
	serviceConfig, err := parser.Parse(*configFile)
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}
	serviceConfig.Debug = serviceConfig.Debug || *debug
	if *port != 0 {
		serviceConfig.Port = *port
	}

	logger, err := logging.NewLogger(*logLevel, os.Stdout, "[KRAKEND]")
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())

	eurekaClient, err := eureka.New(ctx, serviceConfig.ExtraConfig)
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}

	eurekaClient.Register("krakend-gw", GetLocalIP(), serviceConfig.Port)

	routerFactory := krakendgin.NewFactory(krakendgin.Config{
		Engine:         gin.Default(),
		Middlewares:    []gin.HandlerFunc{},
		Logger:         logger,
		HandlerFactory: krakendgin.EndpointHandler,
		ProxyFactory: customProxyFactory{
			logger,
			proxy.DefaultFactoryWithSubscriber(logger, eureka.SubscriberFactory(ctx, eurekaClient)),
		},
	})

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		cancel()
	}()

	routerFactory.New().Run(serviceConfig)

	cancel()
}

// customProxyFactory adds a logging middleware wrapping the internal factory
type customProxyFactory struct {
	logger  logging.Logger
	factory proxy.Factory
}

// New implements the Factory interface
func (cf customProxyFactory) New(cfg *config.EndpointConfig) (p proxy.Proxy, err error) {
	p, err = cf.factory.New(cfg)
	if err == nil {
		p = proxy.NewLoggingMiddleware(cf.logger, cfg.Endpoint)(p)
	}
	return
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
