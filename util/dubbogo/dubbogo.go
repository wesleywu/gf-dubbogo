package dubbogo

import (
	"context"
	"dubbo.apache.org/dubbo-go/v3/common"
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/config"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/natefinch/lumberjack"
	"path"
	"strconv"
)

var consumerConfigBuilder = config.NewConsumerConfigBuilder()

func AddConsumerReference(consumerImplClass string, consumerService common.RPCService, protocol string) {
	consumerConfigBuilder.AddReference(consumerImplClass,
		config.NewReferenceConfigBuilder().
			SetProtocol(protocol).
			Build())
	config.SetConsumerService(consumerService)
}

func buildConsumer(consumerTimeoutSeconds int) *config.ConsumerConfig {
	if consumerTimeoutSeconds > 0 {
		return consumerConfigBuilder.SetRequestTimeout(string(rune(consumerTimeoutSeconds)) + "s").Build()
	}
	return consumerConfigBuilder.Build()
}

func StartConsumers(ctx context.Context, consumerTimeoutSeconds int) error {
	consumerConfig := buildConsumer(consumerTimeoutSeconds)
	if len(consumerConfig.References) == 0 {
		// return when there are no consumer references
		return nil
	}
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

	var (
		loggerPath             string
		loggerLevel            string
		loggerFileName         string
		development            bool
		loggerStdout           bool
		loggerOutputPaths      []string
		loggerErrorOutputPaths []string
	)
	development = g.Cfg().MustGet(ctx, "server.debug", "true").Bool()
	loggerLevel = g.Cfg().MustGet(ctx, "logger.level", "debug").String()
	loggerStdout = g.Cfg().MustGet(ctx, "logger.stdout", "true").Bool()
	loggerPath = g.Cfg().MustGet(ctx, "rpc.consumer.logDir", "./data/log/gf-app").String()
	if g.IsEmpty(loggerPath) {
		loggerPath = g.Cfg().MustGet(ctx, "logger.path", "./data/log/gf-app").String()
	}
	loggerFileName = g.Cfg().MustGet(ctx, "rpc.consumer.logFile", "consumer.log").String()

	if loggerStdout {
		loggerOutputPaths = []string{"stdout", loggerPath}
		loggerErrorOutputPaths = []string{"stderr", loggerPath}
	} else {
		loggerOutputPaths = []string{loggerPath}
		loggerErrorOutputPaths = []string{loggerPath}
	}

	registryConfigBuilder.SetParams(map[string]string{
		constant.NacosLogDirKey:   loggerPath,
		constant.NacosCacheDirKey: loggerPath,
		constant.NacosLogLevelKey: loggerLevel,
	})

	rootConfig := config.NewRootConfigBuilder().
		AddRegistry(registryId, registryConfigBuilder.Build()).
		SetLogger(config.NewLoggerConfigBuilder().
			SetZapConfig(config.ZapConfig{
				Level:            loggerLevel,
				Development:      development,
				OutputPaths:      loggerOutputPaths,
				ErrorOutputPaths: loggerErrorOutputPaths,
			}).
			SetLumberjackConfig(&lumberjack.Logger{
				Filename: path.Join(loggerPath, loggerFileName),
			}).Build()).
		SetConsumer(consumerConfig).
		Build()
	if err := config.Load(config.WithRootConfig(rootConfig)); err != nil {
		return err
	}
	return nil
}

func StartProvider(ctx context.Context, implClassName string, providerService common.RPCService, shutdownCallbacks ...func()) error {
	port := g.Cfg().MustGet(ctx, "rpc.provider.port").String()
	if _, err := strconv.Atoi(port); err != nil {
		return gerror.New("需要指定整形的 port 参数，建议20000以上，不能和其他服务重复")
	}

	config.SetProviderService(providerService)
	if shutdownCallbacks != nil {
		extension.AddCustomShutdownCallback(func() {
			for _, callback := range shutdownCallbacks {
				callback()
			}
		})
	}

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

	var (
		loggerPath             string
		loggerLevel            string
		loggerFileName         string
		development            bool
		loggerStdout           bool
		loggerOutputPaths      []string
		loggerErrorOutputPaths []string
	)
	development = g.Cfg().MustGet(ctx, "server.debug", "true").Bool()
	loggerLevel = g.Cfg().MustGet(ctx, "logger.level", "debug").String()
	loggerStdout = g.Cfg().MustGet(ctx, "logger.stdout", "true").Bool()
	loggerPath = g.Cfg().MustGet(ctx, "rpc.provider.logDir", "./data/log/gf-app").String()
	if g.IsEmpty(loggerPath) {
		loggerPath = g.Cfg().MustGet(ctx, "logger.path", "./data/log/gf-app").String()
	}
	loggerFileName = g.Cfg().MustGet(ctx, "rpc.provider.logFile", "provider.log").String()

	if loggerStdout {
		loggerOutputPaths = []string{"stdout", loggerPath}
		loggerErrorOutputPaths = []string{"stderr", loggerPath}
	} else {
		loggerOutputPaths = []string{loggerPath}
		loggerErrorOutputPaths = []string{loggerPath}
	}

	registryConfigBuilder.SetParams(map[string]string{
		constant.NacosLogDirKey:   loggerPath,
		constant.NacosCacheDirKey: loggerPath,
		constant.NacosLogLevelKey: loggerLevel,
	})

	rootConfig := config.NewRootConfigBuilder().
		SetProvider(config.NewProviderConfigBuilder().
			AddService(implClassName, config.NewServiceConfigBuilder().
				Build()).
			Build()).
		AddRegistry(registryId, registryConfigBuilder.Build()).
		SetMetadataReport(config.NewMetadataReportConfigBuilder().
			SetProtocol(registryProtocol).
			SetAddress(registryAddress).Build()).
		SetLogger(config.NewLoggerConfigBuilder().
			SetZapConfig(config.ZapConfig{
				Level:            loggerLevel,
				Development:      development,
				OutputPaths:      loggerOutputPaths,
				ErrorOutputPaths: loggerErrorOutputPaths,
			}).
			SetLumberjackConfig(&lumberjack.Logger{
				Filename: path.Join(loggerPath, loggerFileName),
			}).Build()).
		AddProtocol("tripleKey", config.NewProtocolConfigBuilder().
			SetName("tri").
			SetPort(port).
			Build()).
		Build()
	if err := config.Load(config.WithRootConfig(rootConfig)); err != nil {
		return err
	}
	select {}
}
