package container

import (
	"context"
	"github.com/sarulabs/di"
	"github.com/teneta-io/dcp/internal/config"
	"github.com/teneta-io/dcp/internal/http"
	"github.com/teneta-io/dcp/internal/service"
	"github.com/teneta-io/dcp/pkg/rabbitmq"
	"github.com/teneta-io/dcp/pkg/redis"
	"go.uber.org/zap"
)

var container di.Container

func Build(ctx context.Context) di.Container {
	builder, _ := di.NewBuilder()

	if err := builder.Add([]di.Def{
		{
			Name: "Logger",
			Build: func(ctn di.Container) (i interface{}, e error) {
				l, err := zap.NewDevelopment()

				if err != nil {
					return nil, err
				}

				zap.ReplaceGlobals(l)

				return l, nil
			},
			Close: func(obj interface{}) error {
				return obj.(*zap.Logger).Sync()
			},
		},
		{
			Name: "Config",
			Build: func(ctn di.Container) (interface{}, error) {
				return config.New()
			},
		},
		{
			Name: "Server",
			Build: func(ctn di.Container) (interface{}, error) {
				cfg := ctn.Get("Config").(*config.Config)
				logger := ctn.Get("Logger").(*zap.Logger)
				taskService := ctn.Get("TaskService").(*service.TaskService)

				return http.New(ctx, &cfg.ServerConfig, logger, taskService), nil
			},
		},
		{
			Name: "Redis",
			Build: func(ctn di.Container) (interface{}, error) {
				cfg := ctn.Get("Config").(*config.Config)

				return redis.New(ctx, &cfg.RedisConfig)
			},
		},
		{
			Name: "RabbitMQ",
			Build: func(ctn di.Container) (interface{}, error) {
				cfg := ctn.Get("Config").(*config.Config)
				logger := ctn.Get("Logger").(*zap.Logger)

				return rabbitmq.NewClient(ctx, &cfg.RabbitMQConfig, logger)
			},
		},
		{
			Name: "TaskSubscriber",
			Build: func(ctn di.Container) (interface{}, error) {
				client := ctn.Get("RabbitMQ").(*rabbitmq.RabbitMQ)

				return rabbitmq.NewTaskSubscriber(client), nil
			},
		},
		{
			Name: "TaskService",
			Build: func(ctn di.Container) (interface{}, error) {
				logger := ctn.Get("Logger").(*zap.Logger)
				taskPublisher := ctn.Get("TaskSubscriber").(*rabbitmq.TaskSubscriber)

				return service.NewTaskService(logger, taskPublisher), nil
			},
		},
	}...); err != nil {
		panic(err)
	}
	container = builder.Build()

	return container
}
