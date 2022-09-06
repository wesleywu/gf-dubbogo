package dubbogo

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/config"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/natefinch/lumberjack"
	"path"
)

func StartProvider(_ context.Context, registry *Registry, provider *ProviderInfo, logger *LoggerOption) error {
	if provider.Port <= 100 {
		return gerror.New("需要指定大于100的 port 参数，建议20000以上，不能和其他服务重复")
	}

	config.SetProviderService(provider.Service)
	if provider.ShutdownCallbacks != nil {
		extension.AddCustomShutdownCallback(func() {
			for _, callback := range provider.ShutdownCallbacks {
				callback()
			}
		})
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

	if logger.Stdout {
		loggerOutputPaths = []string{"stdout", logger.LogDir}
		loggerErrorOutputPaths = []string{"stderr", logger.LogDir}
	} else {
		loggerOutputPaths = []string{logger.LogDir}
		loggerErrorOutputPaths = []string{logger.LogDir}
	}

	registryConfigBuilder.SetParams(map[string]string{
		constant.NacosLogDirKey:   logger.LogDir,
		constant.NacosCacheDirKey: logger.LogDir,
		constant.NacosLogLevelKey: logger.Level,
	})

	protocolConfigBuilder := config.NewProtocolConfigBuilder().
		SetName(provider.Protocol).
		SetPort(gconv.String(provider.Port))
	if !g.IsEmpty(provider.IP) {
		protocolConfigBuilder = protocolConfigBuilder.SetIp(provider.IP)
	}

	rootConfig := config.NewRootConfigBuilder().
		SetProvider(config.NewProviderConfigBuilder().
			AddService(provider.ServerImplStructName, config.NewServiceConfigBuilder().
				Build()).
			Build()).
		AddRegistry(registry.Id, registryConfigBuilder.Build()).
		SetMetadataReport(config.NewMetadataReportConfigBuilder().
			SetProtocol(registry.Type).
			SetAddress(registry.Address).Build()).
		SetLogger(config.NewLoggerConfigBuilder().
			SetZapConfig(config.ZapConfig{
				Level:            logger.Level,
				Development:      logger.Development,
				OutputPaths:      loggerOutputPaths,
				ErrorOutputPaths: loggerErrorOutputPaths,
			}).
			SetLumberjackConfig(&lumberjack.Logger{
				Filename: path.Join(logger.LogDir, logger.LogFileName),
			}).Build()).
		AddProtocol("tripleKey", protocolConfigBuilder.Build()).
		Build()
	if err := config.Load(config.WithRootConfig(rootConfig)); err != nil {
		return err
	}
	select {}
}

// StartProviderAndConsumers 同时启动一个provider，以及一到多个 consumer
func StartProviderAndConsumers(_ context.Context, registry *Registry, provider *ProviderInfo, consumers []*ConsumerReference, consumerOption *ConsumerOption, logger *LoggerOption) error {
	for _, consumer := range consumers {
		AddConsumerReference(consumer)
	}
	consumerConfig := buildConsumerConfig(consumerConfigBuilder, consumerOption)

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

	if logger.Stdout {
		loggerOutputPaths = []string{"stdout", logger.LogDir}
		loggerErrorOutputPaths = []string{"stderr", logger.LogDir}
	} else {
		loggerOutputPaths = []string{logger.LogDir}
		loggerErrorOutputPaths = []string{logger.LogDir}
	}

	registryConfigBuilder.SetParams(map[string]string{
		constant.NacosLogDirKey:   logger.LogDir,
		constant.NacosCacheDirKey: logger.LogDir,
		constant.NacosLogLevelKey: logger.Level,
	})

	protocolConfigBuilder := config.NewProtocolConfigBuilder().
		SetName(provider.Protocol).
		SetPort(gconv.String(provider.Port))
	if !g.IsEmpty(provider.IP) {
		protocolConfigBuilder = protocolConfigBuilder.SetIp(provider.IP)
	}

	rootConfig := config.NewRootConfigBuilder().
		AddRegistry(registry.Id, registryConfigBuilder.Build()).
		SetMetadataReport(config.NewMetadataReportConfigBuilder().
			SetProtocol(registry.Type).
			SetAddress(registry.Address).Build()).
		SetProvider(config.NewProviderConfigBuilder().
			AddService(provider.ServerImplStructName, config.NewServiceConfigBuilder().
				Build()).
			Build()).
		SetConsumer(consumerConfig).
		SetLogger(config.NewLoggerConfigBuilder().
			SetZapConfig(config.ZapConfig{
				Level:            logger.Level,
				Development:      logger.Development,
				OutputPaths:      loggerOutputPaths,
				ErrorOutputPaths: loggerErrorOutputPaths,
			}).
			SetLumberjackConfig(&lumberjack.Logger{
				Filename: path.Join(logger.LogDir, logger.LogFileName),
			}).Build()).
		AddProtocol("tripleKey", protocolConfigBuilder.Build()).
		Build()
	if err := config.Load(config.WithRootConfig(rootConfig)); err != nil {
		return err
	}
	select {}
}
