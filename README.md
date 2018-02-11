Krakend Eureka
====

An Eureka client and subscriber for the [KrakenD](http://www.krakend.io) framework.

## Build the example

Go 1.8 is a requirement
```bash
$ make
```

## Using Eureka client

```go
	eurekaClient, err := eureka.New(ctx, serviceConfig.ExtraConfig)
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}

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
```

## Registering your application in Eureka

```go
eurekaClient.Register("krakend-gw", GetLocalIP(), serviceConfig.Port)
```

## Run

Running it as a common executable, logs are send to the stdOut and some options are available at the CLI

```bash
$ ./krakend_eureka_example
Usage of ./krakend_eureka_example:
  -c string
        Path to the configuration filename (default "/etc/krakend/configuration.json")
  -d	Enable the debug
  -l string
        Logging level (default "ERROR")
  -p int
        Port of the service
```