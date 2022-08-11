package dubbogo

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/config"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/natefinch/lumberjack"
	"path"
)

var consumerConfigBuilder = config.NewConsumerConfigBuilder()

// AddConsumerReference 增加一个 consumer 的依赖配置
func AddConsumerReference(consumer *ConsumerReference) {
	consumerConfigBuilder.AddReference(consumer.ClientImplStructName,
		config.NewReferenceConfigBuilder().
			SetProtocol(consumer.Protocol).
			Build())
	config.SetConsumerService(consumer.Service)
}

func buildConsumer(consumerOption *ConsumerOption) *config.ConsumerConfig {
	if consumerOption.TimeoutSeconds > 0 {
		consumerConfigBuilder.SetRequestTimeout(string(rune(consumerOption.TimeoutSeconds)) + "s")
	}
	return consumerConfigBuilder.SetCheck(consumerOption.CheckProviderExists).Build()
}

// StartConsumersByCfg 启动通过 AddConsumerReference 添加的所有的 Consumer, 从 config 配置文件中读取若干配置
func StartConsumersByCfg(ctx context.Context, checkProviderExists bool, timeoutSeconds int) error {
	registryId := g.Cfg().MustGet(ctx, "rpc.registry.id", "nacosRegistry").String()
	registryProtocol := g.Cfg().MustGet(ctx, "rpc.registry.protocol", "nacos").String()
	registryAddress := g.Cfg().MustGet(ctx, "rpc.registry.address", "127.0.0.1:8848").String()
	registryConfigBuilder := config.NewRegistryConfigBuilder().
		SetProtocol(registryProtocol).
		SetAddress(registryAddress)
	registryNamespace := g.Cfg().MustGet(ctx, "rpc.registry.namespace", "public").String()
	if registryProtocol == "nacos" {
		registryConfigBuilder = registryConfigBuilder.SetNamespace(registryNamespace)
	}
	development := g.Cfg().MustGet(ctx, "server.debug", "true").Bool()
	loggerStdout := g.Cfg().MustGet(ctx, "logger.stdout", "true").Bool()
	loggerPath := g.Cfg().MustGet(ctx, "rpc.consumer.logDir", "./data/log/gf-app").String()
	if g.IsEmpty(loggerPath) {
		loggerPath = g.Cfg().MustGet(ctx, "logger.path", "./data/log/gf-app").String()
	}
	loggerFileName := g.Cfg().MustGet(ctx, "rpc.consumer.logFile", "consumer.log").String()
	loggerLevel := g.Cfg().MustGet(ctx, "rpc.provider.logLevel", "warn").String()

	return StartConsumers(ctx, &Registry{
		Id:        registryId,
		Type:      registryProtocol,
		Address:   registryAddress,
		Namespace: registryNamespace,
	}, &ConsumerOption{
		CheckProviderExists: checkProviderExists,
		TimeoutSeconds:      timeoutSeconds,
	}, &LoggerOption{
		Development: development,
		Stdout:      loggerStdout,
		LogDir:      loggerPath,
		LogFileName: loggerFileName,
		Level:       loggerLevel,
	})
}

// StartConsumers 启动通过 AddConsumerReference 添加的所有的 Consumer
func StartConsumers(_ context.Context, registry *Registry, consumerOption *ConsumerOption, loggerOption *LoggerOption) error {
	consumerConfig := buildConsumer(consumerOption)
	if len(consumerConfig.References) == 0 {
		// return when there are no consumer references
		return nil
	}
	registryConfigBuilder := config.NewRegistryConfigBuilder().
		SetProtocol(registry.Type).
		SetAddress(registry.Address)
	if registry.Type == "nacos" && !g.IsEmpty(registry.Namespace) {
		registryConfigBuilder = registryConfigBuilder.SetNamespace(registry.Namespace)
	}

	var (
		loggerOutputPaths      []string
		loggerErrorOutputPaths []string
	)
	if loggerOption.Stdout {
		loggerOutputPaths = []string{"stdout", loggerOption.LogDir}
		loggerErrorOutputPaths = []string{"stderr", loggerOption.LogDir}
	} else {
		loggerOutputPaths = []string{loggerOption.LogDir}
		loggerErrorOutputPaths = []string{loggerOption.LogDir}
	}

	registryConfigBuilder.SetParams(map[string]string{
		constant.NacosLogDirKey:   loggerOption.LogDir,
		constant.NacosCacheDirKey: loggerOption.LogDir,
		constant.NacosLogLevelKey: loggerOption.Level,
	})

	rootConfig := config.NewRootConfigBuilder().
		AddRegistry(registry.Id, registryConfigBuilder.Build()).
		SetLogger(config.NewLoggerConfigBuilder().
			SetZapConfig(config.ZapConfig{
				Level:            loggerOption.Level,
				Development:      loggerOption.Development,
				OutputPaths:      loggerOutputPaths,
				ErrorOutputPaths: loggerErrorOutputPaths,
			}).
			SetLumberjackConfig(&lumberjack.Logger{
				Filename: path.Join(loggerOption.LogDir, loggerOption.LogFileName),
			}).Build()).
		SetConsumer(consumerConfig).
		Build()
	if err := config.Load(config.WithRootConfig(rootConfig)); err != nil {
		return err
	}
	return nil
}
