# 【K8s源码品读】005：Phase 1 - kube-apiserver 权限相关的三个核心概念

## 聚焦目标

理解启动kube-apiserver的权限相关的三个核心概念 `Authentication`/`Authorization`/`Admission`



## 目录

1. [kube-apiserver的启动](#run)
2. [kube-apiserver的三个Server](#three-servers)
3. [KubeAPIServer的创建过程](#KubeAPIServer)
   1. [通用配置概况](#GenericConfig)
   2. [通用配置中的认证](#Authentication)
   3. [通用配置中的授权](#Authorization)
   4. [通用配置中的准入机制](#Admission)



## Run

```go
// 类似kubectl的源代码，kube-apiserver的命令行工具也使用了cobra，我们很快就能找到启动的入口
RunE: func(cmd *cobra.Command, args []string) error {
			// 这里包含2个参数，前者是参数completedOptions，后者是一个stopCh <-chan struct{}
			return Run(completedOptions, genericapiserver.SetupSignalHandler())
		}

/*
	在这里，我们可以和kubectl结合起来思考：
	kubectl是一个命令行工具，执行完命令就退出；kube-apiserver是一个常驻的服务器进程，监听端口
	这里引入了一个stopCh <-chan struct{}，可以在启动后，用一个 <-stopCh 作为阻塞，使程序不退出
	用channel阻塞进程退出，对比传统的方法 - 用一个永不退出的for循环，是一个很优雅的实现
*/
```



## Three Servers

```go
// 在CreateServerChain这个函数下，创建了3个server
func CreateServerChain(){
  // API扩展服务，主要针对CRD
	createAPIExtensionsServer(){} 
  // API核心服务，包括常见的Pod/Deployment/Service，我们今天的重点聚焦在这里
  // 我会跳过很多非核心的配置参数，一开始就去研究细节，很影响整体代码的阅读效率
	CreateKubeAPIServer(){} 
  // API聚合服务，主要针对metrics
	createAggregatorServer(){} 
}
```



## KubeAPIServer

```go
// 创建配置的流程
func CreateKubeAPIServerConfig(){
  // 创建通用配置genericConfig
  genericConfig, versionedInformers, insecureServingInfo, serviceResolver, pluginInitializers, admissionPostStartHook, storageFactory, err := buildGenericConfig(s.ServerRunOptions, proxyTransport)
}
```



### GenericConfig

```go
// 通用配置的创建
func buildGenericConfig(s *options.ServerRunOptions,proxyTransport *http.Transport){
  // Insecure对应的非安全的通信，也就是HTTP
  if lastErr = s.InsecureServing...
  // Secure对应的就是HTTPS
  if lastErr = s.SecureServing...
  // OpenAPIConfig是对外提供的API文档
  genericConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig()
  // 这一块是storageFactory的实例化，可以看到采用的是etcd作为存储方案
  storageFactoryConfig := kubeapiserver.NewStorageFactoryConfig()
	storageFactoryConfig.APIResourceConfig = genericConfig.MergedResourceConfig
	completedStorageFactoryConfig, err := storageFactoryConfig.Complete(s.Etcd)
	storageFactory, lastErr = completedStorageFactoryConfig.New()
  // Authentication 认证相关
  if lastErr = s.Authentication.ApplyTo()...
  // Authorization 授权相关
  genericConfig.Authorization.Authorizer, genericConfig.RuleResolver, err = BuildAuthorizer()
  // Admission 准入机制
  err = s.Admission.ApplyTo()
}
```



### Authentication

```go
func (o *BuiltInAuthenticationOptions) ApplyTo(){
  // 前面都是对认证config进行参数设置，这里才是真正的实例化
  authInfo.Authenticator, openAPIConfig.SecurityDefinitions, err = authenticatorConfig.New()
}

// New这块的代码，我们要抓住核心变量authenticators和tokenAuthenticators，也就是各种认证方法
func (config Config) New() (authenticator.Request, *spec.SecurityDefinitions, error) {
  // 核心变量authenticators和tokenAuthenticators
	var authenticators []authenticator.Request
  var tokenAuthenticators []authenticator.Token

	if config.RequestHeaderConfig != nil {
		// 1. 添加requestHeader
		authenticators = append(authenticators, authenticator.WrapAudienceAgnosticRequest(config.APIAudiences, requestHeaderAuthenticator))
	}

	if config.ClientCAContentProvider != nil {
		// 2. 添加ClientCA
    authenticators = append(authenticators, certAuth)
	}

	if len(config.TokenAuthFile) > 0 {
		// 3. token 添加tokenfile
		tokenAuthenticators = append(tokenAuthenticators, authenticator.WrapAudienceAgnosticToken(config.APIAudiences, tokenAuth))
	}
  
  // 4. token 添加 service account，分两种来源
	if len(config.ServiceAccountKeyFiles) > 0 {
		tokenAuthenticators = append(tokenAuthenticators, serviceAccountAuth)
	}
	if utilfeature.DefaultFeatureGate.Enabled(features.TokenRequest) && config.ServiceAccountIssuer != "" {
		tokenAuthenticators = append(tokenAuthenticators, serviceAccountAuth)
	}
	if config.BootstrapToken {
		if config.BootstrapTokenAuthenticator != nil {
      // 5. token 添加 bootstrap
			tokenAuthenticators = append(tokenAuthenticators, authenticator.WrapAudienceAgnosticToken(config.APIAudiences, config.BootstrapTokenAuthenticator))
		}
	}

	if len(config.OIDCIssuerURL) > 0 && len(config.OIDCClientID) > 0 {
		// 6. token 添加 oidc
    Authenticators = append(tokenAuthenticators, authenticator.WrapAudienceAgnosticToken(config.APIAudiences, oidcAuth))
	}
	if len(config.WebhookTokenAuthnConfigFile) > 0 {
		// 7. token 添加 webhook
		tokenAuthenticators = append(tokenAuthenticators, webhookTokenAuth)
	}

  // 8. 组合tokenAuthenticators到tokenAuthenticators中
	if len(tokenAuthenticators) > 0 {
		tokenAuth := tokenunion.New(tokenAuthenticators...)
		if config.TokenSuccessCacheTTL > 0 || config.TokenFailureCacheTTL > 0 {
			tokenAuth = tokencache.New(tokenAuth, true, config.TokenSuccessCacheTTL, config.TokenFailureCacheTTL)
		}
		authenticators = append(authenticators, bearertoken.New(tokenAuth), websocket.NewProtocolAuthenticator(tokenAuth))
	}

  // 9. 没有任何认证方式且启用了Anonymous
	if len(authenticators) == 0 {
		if config.Anonymous {
			return anonymous.NewAuthenticator(), &securityDefinitions, nil
		}
		return nil, &securityDefinitions, nil
	}

  // 10. 组合authenticators
	authenticator := union.New(authenticators...)

	return authenticator, &securityDefinitions, nil
}
```

复杂的Authentication模块的初始化顺序我们看完了，有初步的了解即可，没必要去强制记忆其中的加载顺序。



### Authorization

```go
func BuildAuthorizer(){
  // 与上面一致，实例化是在这个New中
  return authorizationConfig.New()
}

// 不得不说，Authorizer这块的阅读体验更好
func (config Config) New() (authorizer.Authorizer, authorizer.RuleResolver, error) {
  // 必须传入一个Authorizer机制
	if len(config.AuthorizationModes) == 0 {
		return nil, nil, fmt.Errorf("at least one authorization mode must be passed")
	}

	var (
		authorizers   []authorizer.Authorizer
		ruleResolvers []authorizer.RuleResolver
	)

	for _, authorizationMode := range config.AuthorizationModes {
		// 具体的mode定义，可以跳转到对应的链接去看，今天不细讲
		switch authorizationMode {
		case modes.ModeNode:
			authorizers = append(authorizers, nodeAuthorizer)
			ruleResolvers = append(ruleResolvers, nodeAuthorizer)

		case modes.ModeAlwaysAllow:
			authorizers = append(authorizers, alwaysAllowAuthorizer)
			ruleResolvers = append(ruleResolvers, alwaysAllowAuthorizer)
      
		case modes.ModeAlwaysDeny:
			authorizers = append(authorizers, alwaysDenyAuthorizer)
			ruleResolvers = append(ruleResolvers, alwaysDenyAuthorizer)
      
		case modes.ModeABAC:
			authorizers = append(authorizers, abacAuthorizer)
			ruleResolvers = append(ruleResolvers, abacAuthorizer)
      
		case modes.ModeWebhook:
			authorizers = append(authorizers, webhookAuthorizer)
			ruleResolvers = append(ruleResolvers, webhookAuthorizer)
      
		case modes.ModeRBAC:
			authorizers = append(authorizers, rbacAuthorizer)
			ruleResolvers = append(ruleResolvers, rbacAuthorizer)
		default:
			return nil, nil, fmt.Errorf("unknown authorization mode %s specified", authorizationMode)
		}
	}

	return union.New(authorizers...), union.NewRuleResolvers(ruleResolvers...), nil
}

const (
	// ModeAlwaysAllow is the mode to set all requests as authorized
	ModeAlwaysAllow string = "AlwaysAllow"
	// ModeAlwaysDeny is the mode to set no requests as authorized
	ModeAlwaysDeny string = "AlwaysDeny"
	// ModeABAC is the mode to use Attribute Based Access Control to authorize
	ModeABAC string = "ABAC"
	// ModeWebhook is the mode to make an external webhook call to authorize
	ModeWebhook string = "Webhook"
	// ModeRBAC is the mode to use Role Based Access Control to authorize
	ModeRBAC string = "RBAC"
	// ModeNode is an authorization mode that authorizes API requests made by kubelets.
	ModeNode string = "Node"
)
```



### Admission

```go
// 查看定义
err = s.Admission.ApplyTo()
func (a *AdmissionOptions) ApplyTo(){
  return a.GenericAdmission.ApplyTo()
}

func (ps *Plugins) NewFromPlugins(){
  for _, pluginName := range pluginNames {
		// InitPlugin 为初始化的工作
		plugin, err := ps.InitPlugin(pluginName, pluginConfig, pluginInitializer)
		if err != nil {
			return nil, err
		}
	}
}

func (ps *Plugins) InitPlugin(name string, config io.Reader, pluginInitializer PluginInitializer) (Interface, error){
  // 获取plugin
  plugin, found, err := ps.getPlugin(name, config)
}

// 查看一下Interface的定义，就是对准入机制的控制
// Interface is an abstract, pluggable interface for Admission Control decisions.
type Interface interface {
	Handles(operation Operation) bool
}

// 再去看看获取plugin的地方
func (ps *Plugins) getPlugin(name string, config io.Reader) (Interface, bool, error) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
  // 我们再去研究ps.registry这个参数是在哪里被初始化的
	f, found := ps.registry[name]
}

// 接下来，我们从kube-apiserver启动过程，逐步找到Admission被初始化的地方
// 启动命令
command := app.NewAPIServerCommand()
// server配置
s := options.NewServerRunOptions()
// admission选项
Admission:               kubeoptions.NewAdmissionOptions()
// 注册准入机制
RegisterAllAdmissionPlugins(options.Plugins)
// 准入机制的所有内容
func RegisterAllAdmissionPlugins(plugins *admission.Plugins){
  // 这里有很多plugin的注册
}

// 往上翻，我们能找到所有plugin，也就是准入机制的定义
var AllOrderedPlugins = []string{
	admit.PluginName,                        // AlwaysAdmit
	autoprovision.PluginName,                // NamespaceAutoProvision
	lifecycle.PluginName,                    // NamespaceLifecycle
	exists.PluginName,                       // NamespaceExists
	scdeny.PluginName,                       // SecurityContextDeny
	antiaffinity.PluginName,                 // LimitPodHardAntiAffinityTopology
	podpreset.PluginName,                    // PodPreset
	limitranger.PluginName,                  // LimitRanger
	serviceaccount.PluginName,               // ServiceAccount
	noderestriction.PluginName,              // NodeRestriction
	nodetaint.PluginName,                    // TaintNodesByCondition
	alwayspullimages.PluginName,             // AlwaysPullImages
	imagepolicy.PluginName,                  // ImagePolicyWebhook
	podsecuritypolicy.PluginName,            // PodSecurityPolicy
	podnodeselector.PluginName,              // PodNodeSelector
	podpriority.PluginName,                  // Priority
	defaulttolerationseconds.PluginName,     // DefaultTolerationSeconds
	podtolerationrestriction.PluginName,     // PodTolerationRestriction
	exec.DenyEscalatingExec,                 // DenyEscalatingExec
	exec.DenyExecOnPrivileged,               // DenyExecOnPrivileged
	eventratelimit.PluginName,               // EventRateLimit
	extendedresourcetoleration.PluginName,   // ExtendedResourceToleration
	label.PluginName,                        // PersistentVolumeLabel
	setdefault.PluginName,                   // DefaultStorageClass
	storageobjectinuseprotection.PluginName, // StorageObjectInUseProtection
	gc.PluginName,                           // OwnerReferencesPermissionEnforcement
	resize.PluginName,                       // PersistentVolumeClaimResize
	runtimeclass.PluginName,                 // RuntimeClass
	certapproval.PluginName,                 // CertificateApproval
	certsigning.PluginName,                  // CertificateSigning
	certsubjectrestriction.PluginName,       // CertificateSubjectRestriction
	defaultingressclass.PluginName,          // DefaultIngressClass

	// new admission plugins should generally be inserted above here
	// webhook, resourcequota, and deny plugins must go at the end

	mutatingwebhook.PluginName,   // MutatingAdmissionWebhook
	validatingwebhook.PluginName, // ValidatingAdmissionWebhook
	resourcequota.PluginName,     // ResourceQuota
	deny.PluginName,              // AlwaysDeny
}
```

