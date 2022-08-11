package dubbogo

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/common"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/config"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/natefinch/lumberjack"
	"path"
)

func StartProviderByCfg(ctx context.Context, implClassName string, providerService common.RPCService, shutdownCallbacks ...func()) error {
	ip := g.Cfg().MustGet(ctx, "rpc.provider.ip", "").String()
	port := g.Cfg().MustGet(ctx, "rpc.provider.port").Int()
	if port <= 100 {
		return gerror.New("需要指定大于100的 port 参数，建议20000以上，不能和其他服务重复")
	}
	registryId := g.Cfg().MustGet(ctx, "rpc.registry.id", "nacosRegistry").String()
	registryProtocol := g.Cfg().MustGet(ctx, "rpc.registry.protocol", "nacos").String()
	registryAddress := g.Cfg().MustGet(ctx, "rpc.registry.address", "127.0.0.1:8848").String()
	registryNamespace := g.Cfg().MustGet(ctx, "rpc.registry.namespace", "public").String()
	development := g.Cfg().MustGet(ctx, "server.debug", "true").Bool()
	loggerStdout := g.Cfg().MustGet(ctx, "logger.stdout", "true").Bool()
	loggerPath := g.Cfg().MustGet(ctx, "rpc.provider.logDir", "./data/log/gf-app").String()
	if g.IsEmpty(loggerPath) {
		loggerPath = g.Cfg().MustGet(ctx, "logger.path", "./data/log/gf-app").String()
	}
	loggerFileName := g.Cfg().MustGet(ctx, "rpc.provider.logFile", "provider.log").String()
	loggerLevel := g.Cfg().MustGet(ctx, "rpc.provider.logLevel", "warn").String()
	return StartProvider(ctx, &Registry{
		Id:        registryId,
		Type:      registryProtocol,
		Address:   registryAddress,
		Namespace: registryNamespace,
	}, &ProviderInfo{
		ServerImplStructName: implClassName,
		Service:              providerService,
		Protocol:             "tri",
		Port:                 port,
		IP:                   ip,
		ShutdownCallbacks:    shutdownCallbacks,
	}, &LoggerOption{
		Development: development,
		Stdout:      loggerStdout,
		LogDir:      loggerPath,
		LogFileName: loggerFileName,
		Level:       loggerLevel,
	})
}

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
