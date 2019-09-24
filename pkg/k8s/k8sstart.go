package k8s

import (
	"fmt"
	"github.com/costinm/istio-vm/pkg/istiostart"
	"github.com/hashicorp/go-multierror"
	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/galley/pkg/config/event"
	"istio.io/istio/galley/pkg/config/schema"
	"istio.io/istio/galley/pkg/config/source/kube"
	"istio.io/istio/galley/pkg/config/source/kube/apiserver"
	"istio.io/istio/pilot/pkg/config/clusterregistry"
	"istio.io/istio/pilot/pkg/config/kube/crd/controller"
	"istio.io/istio/pilot/pkg/config/kube/ingress"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/serviceregistry"
	"istio.io/istio/pilot/pkg/serviceregistry/aggregate"
	"istio.io/istio/pkg/config/mesh"
	"istio.io/istio/pkg/config/schemas"
	"istio.io/pkg/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	controller2 "istio.io/istio/pilot/pkg/serviceregistry/kube/controller"
)

// Helpers to configure the k8s-dependent registries
// To reduce binary size/deps, the standalone hyperistio for VMs will try to not depend on k8s, keeping all
// init deps in this package.

type K8SServer struct {
	IstioServer       *istiostart.Server
	ControllerOptions controller2.Options

	kubeClient   kubernetes.Interface
	kubeCfg      *rest.Config
	kubeRegistry *controller2.Controller
	multicluster *clusterregistry.Multicluster
	args         *istiostart.PilotArgs
}

func InitK8S(is *istiostart.Server, clientset *kubernetes.Clientset, config *rest.Config, args *istiostart.PilotArgs) (*K8SServer, error) {
	s := &K8SServer{
		IstioServer: is,
		kubeCfg:     config,
		kubeClient:  clientset,
		args:        args,
	}

	// Istio's own K8S config controller - shouldn't be needed if MCP is used.
	// TODO: ordering, this needs to go before discovery.
	if err := s.initConfigController(args); err != nil {
		return nil, fmt.Errorf("cluster registries: %v", err)
	}
	return s, nil
}

func (s *K8SServer) OnXDSStart(xds model.XDSUpdater) {
	s.kubeRegistry.XDSUpdater = xds
}

func (s *K8SServer) InitK8SDiscovery(is *istiostart.Server, clientset *kubernetes.Clientset, config *rest.Config, args *istiostart.PilotArgs) (*K8SServer, error) {
	if err := s.createK8sServiceControllers(s.IstioServer.ServiceController, args); err != nil {
		return nil, fmt.Errorf("cluster registries: %v", err)
	}

	if err := s.initClusterRegistries(args); err != nil {
		return nil, fmt.Errorf("cluster registries: %v", err)
	}

	// kubeRegistry may use the environment for push status reporting.
	// TODO: maybe all registries should have this as an optional field ?
	s.kubeRegistry.Env = s.IstioServer.Environment
	s.kubeRegistry.InitNetworkLookup(s.IstioServer.MeshNetworks)
	// EnvoyXDSServer is not initialized yet - since initialization adds all 'service' handlers, which depends
	// on this beeing done. Instead we use the callback.
	//s.kubeRegistry.XDSUpdater = s.IstioServer.EnvoyXdsServer

	return s, nil
}

func (s *K8SServer) WaitForCacheSync(stop <-chan struct{}) bool {
	// TODO: remove dependency on k8s lib
	if !cache.WaitForCacheSync(stop, func() bool {
		//if s.s.kubeRegistry != nil {
		//	if !s.s.kubeRegistry.HasSynced() {
		//		return false
		//	}
		//}
		if !s.IstioServer.ConfigController.HasSynced() {
			return false
		}
		return true
	}) {
		log.Errorf("Failed waiting for cache sync")
		return false
	}

	return true
}

// initClusterRegistries starts the secret controller to watch for remote
// clusters and initialize the multicluster structures.s.
func (s *K8SServer) initClusterRegistries(args *istiostart.PilotArgs) (err error) {

	mc, err := clusterregistry.NewMulticluster(s.kubeClient,
		args.Config.ClusterRegistriesNamespace,
		s.ControllerOptions.WatchedNamespace,
		args.DomainSuffix,
		s.ControllerOptions.ResyncPeriod,
		s.IstioServer.ServiceController,
		s.IstioServer.EnvoyXdsServer,
		s.IstioServer.MeshNetworks)

	if err != nil {
		log.Info("Unable to create new Multicluster object")
		return err
	}

	s.multicluster = mc
	return nil
}

// initConfigController creates the config controller in the pilotConfig.
func (s *K8SServer) initConfigController(args *istiostart.PilotArgs) error {
	cfgController, err := s.makeKubeConfigController(args)
	if err != nil {
		return err
	}

	s.IstioServer.ConfigStores = append(s.IstioServer.ConfigStores, cfgController)

	// Defer starting the controller until after the service is created.
	s.IstioServer.AddStartFunc(func(stop <-chan struct{}) error {
		go cfgController.Run(stop)
		return nil
	})

	// If running in ingress mode (requires k8s), wrap the config controller.
	if s.IstioServer.Mesh.IngressControllerMode != meshconfig.MeshConfig_OFF {
		s.IstioServer.ConfigStores = append(s.IstioServer.ConfigStores, ingress.NewController(s.kubeClient, s.IstioServer.Mesh, s.ControllerOptions))

		if ingressSyncer, errSyncer := ingress.NewStatusSyncer(s.IstioServer.Mesh, s.kubeClient,
			args.Namespace, s.ControllerOptions); errSyncer != nil {
			log.Warnf("Disabled ingress status syncer due to %v", errSyncer)
		} else {
			s.IstioServer.AddStartFunc(func(stop <-chan struct{}) error {
				go ingressSyncer.Run(stop)
				return nil
			})
		}
	}

	return nil
}

// createK8sServiceControllers creates all the k8s service controllers under this pilot
func (s *K8SServer) createK8sServiceControllers(serviceControllers *aggregate.Controller, args *istiostart.PilotArgs) (err error) {
	clusterID := string(serviceregistry.KubernetesRegistry)
	log.Infof("Primary Cluster name: %s", clusterID)
	s.ControllerOptions.ClusterID = clusterID
	kubectl := controller2.NewController(s.kubeClient, s.ControllerOptions)
	s.kubeRegistry = kubectl
	serviceControllers.AddRegistry(
		aggregate.Registry{
			Name:             serviceregistry.KubernetesRegistry,
			ClusterID:        clusterID,
			ServiceDiscovery: kubectl,
			Controller:       kubectl,
		})

	return
}

func (s *K8SServer) makeKubeConfigController(args *istiostart.PilotArgs) (model.ConfigStoreCache, error) {
	kubeCfgFile := args.Config.KubeConfig
	configClient, err := controller.NewClient(kubeCfgFile, "", schemas.Istio, s.ControllerOptions.DomainSuffix)
	if err != nil {
		return nil, multierror.Prefix(err, "failed to open a config client.")
	}

	if !args.Config.DisableInstallCRDs {
		if err = configClient.RegisterResources(); err != nil {
			return nil, multierror.Prefix(err, "failed to register custom resources.")
		}
	}

	return controller.NewController(configClient, s.ControllerOptions), nil
}

const (
	// ConfigMapKey should match the expected MeshConfig file name
	ConfigMapKey = "mesh"
)

// GetMeshConfig fetches the ProxyMesh configuration from Kubernetes ConfigMap.
func GetMeshConfig(kube kubernetes.Interface, namespace, name string) (*v1.ConfigMap, *meshconfig.MeshConfig, error) {

	if kube == nil {
		defaultMesh := mesh.DefaultMeshConfig()
		return nil, &defaultMesh, nil
	}

	cfg, err := kube.CoreV1().ConfigMaps(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			defaultMesh := mesh.DefaultMeshConfig()
			return nil, &defaultMesh, nil
		}
		return nil, nil, err
	}

	// values in the data are strings, while proto might use a different data type.
	// therefore, we have to get a value by a key
	cfgYaml, exists := cfg.Data[ConfigMapKey]
	if !exists {
		return nil, nil, fmt.Errorf("missing configuration map key %q", ConfigMapKey)
	}

	meshConfig, err := mesh.ApplyMeshConfigDefaults(cfgYaml)
	if err != nil {
		return nil, nil, err
	}
	return cfg, meshConfig, nil
}

type testHandler struct {
}

func (t testHandler) Handle(e event.Event) {
	log.Debugf("Event %e", e)
}

func (s *K8SServer) NewGalleyK8SSource(resources schema.KubeResources) (src event.Source, err error) {

	//if !p.args.DisableResourceReadyCheck {
	//	if err = checkResourceTypesPresence(k, resources); err != nil {
	//		return
	//	}
	//}

	o := apiserver.Options{
		Client:       kube.NewInterfaces(s.kubeCfg),
		ResyncPeriod: s.ControllerOptions.ResyncPeriod,
		Resources:    resources,
	}
	src = apiserver.New(o)

	src.Dispatch(testHandler{})
	//src.Start()
	//src.Stop()

	return
}
