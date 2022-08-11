package dubbogo

import "dubbo.apache.org/dubbo-go/v3/common"

type Registry struct {
	Id        string // 注册中心的ID，随便去，
	Type      string // 注册中心类型，例如 nacos
	Address   string // 注册中心地址，例如 ip:port
	Namespace string // 仅当注册中心协议为 nacos 时有效
}

type LoggerOption struct {
	Development bool   // 是否为开发阶段，可能输出更详细的日志
	Stdout      bool   // 是否将日志同时输出到 stdout/stderr
	LogDir      string // 日志文件的路径
	LogFileName string // 日志文件的文件名（不含路径）
	Level       string // 日志级别，可选项：[error|warn|info|debug]
}

type ConsumerOption struct {
	CheckProviderExists bool // 是否要求在consumer启动时，对应的provider必须存在
	TimeoutSeconds      int  // 调用 provider 方法的超时时间（秒数）
}

type ProviderInfo struct {
	ServerImplStructName string            // provider 实现 IXxx 接口的 struct 名称
	Service              common.RPCService // provider 实现 IXxx 接口的 struct 实例
	Protocol             string            // 协议，当前只支持 "tri"
	Port                 int               // 侦听的端口号
	IP                   string            // 不建议指定，如果要绑定运行的IP，可指定
	ShutdownCallbacks    []func()          // 关闭时要执行的回调
}

type ConsumerReference struct {
	ClientImplStructName string            // consumer ClientImpl 的 struct 名称（通常在 protobuf 生成的文件中被命名为 XxxClientImpl）
	Service              common.RPCService // consumer ClientImpl 的实例
	Protocol             string            // 协议，当前只支持 "tri"
}
