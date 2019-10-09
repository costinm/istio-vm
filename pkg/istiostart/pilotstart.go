// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package istiostart

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"sync"
	"time"

	"istio.io/pkg/ctrlz/fw"

	"github.com/davecgh/go-spew/spew"
	"github.com/gogo/protobuf/types"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	mcpapi "istio.io/api/mcp/v1alpha1"
	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/pkg/ctrlz"
	"istio.io/pkg/env"
	"istio.io/pkg/filewatcher"
	"istio.io/pkg/log"
	"istio.io/pkg/version"

	"istio.io/istio/pilot/cmd"
	configaggregate "istio.io/istio/pilot/pkg/config/aggregate"
	"istio.io/istio/pilot/pkg/config/coredatamodel"
	"istio.io/istio/pilot/pkg/features"
	"istio.io/istio/pilot/pkg/model"
	istio_networking "istio.io/istio/pilot/pkg/networking/core"
	"istio.io/istio/pilot/pkg/networking/plugin"
	"istio.io/istio/pilot/pkg/networking/util"
	"istio.io/istio/pilot/pkg/proxy/envoy"
	envoyv2 "istio.io/istio/pilot/pkg/proxy/envoy/v2"
	"istio.io/istio/pilot/pkg/serviceregistry/aggregate"
	"istio.io/istio/pilot/pkg/serviceregistry/external"
	"istio.io/istio/pkg/config/constants"
	"istio.io/istio/pkg/config/schemas"
	istiokeepalive "istio.io/istio/pkg/keepalive"
	"istio.io/istio/pkg/mcp/monitoring"
	"istio.io/istio/pkg/mcp/sink"
)

const (
	// ConfigMapKey should match the expected MeshConfig file name
	ConfigMapKey = "mesh"

	requiredMCPCertCheckFreq = 500 * time.Millisecond

	// DefaultMCPMaxMsgSize is the default maximum message size
	DefaultMCPMaxMsgSize = 1024 * 1024 * 4

	// DefaultMCPInitialWindowSize is the default InitialWindowSize value for the gRPC connection.
	DefaultMCPInitialWindowSize = 1024 * 1024

	// DefaultMCPInitialConnWindowSize is the default Initial ConnWindowSize value for the gRPC connection.
	DefaultMCPInitialConnWindowSize = 1024 * 1024

	// URL types supported by the config store
	// example fs:///tmp/configroot
	fsScheme = "fs"
)

var (

	// FilepathWalkInterval dictates how often the file system is walked for config
	FilepathWalkInterval = 100 * time.Millisecond

	// PilotCertDir is the default location for mTLS certificates used by pilot
	// Visible for tests - at runtime can be set by PILOT_CERT_DIR environment variable.
	PilotCertDir = "/etc/certs/"

	// DefaultPlugins is the default list of plugins to enable, when no plugin(s)
	// is specified through the command line
	DefaultPlugins = []string{
		plugin.Authn,
		plugin.Authz,
		plugin.Health,
		plugin.Mixer,
	}
)

func init() {
	// get the grpc server wired up
	// This should only be set before any RPCs are sent or received by this program.
	grpc.EnableTracing = true

	// Export pilot version as metric for fleet analytics.
	pilotVersion := prom.NewGaugeVec(prom.GaugeOpts{
		Name: "pilot_info",
		Help: "Pilot version and build information.",
	}, []string{"version"})
	prom.MustRegister(pilotVersion)
	pilotVersion.With(prom.Labels{"version": version.Info.String()}).Set(1)
}

// MeshArgs provide configuration options for the mesh. If ConfigFile is provided, an attempt will be made to
// load the mesh from the file. Otherwise, a default mesh will be used with optional overrides.
type MeshArgs struct {
	ConfigFile      string
	MixerAddress    string
	RdsRefreshDelay *types.Duration
}

// ConfigArgs provide configuration options for the configuration controller. If FileDir is set, that directory will
// be monitored for CRD yaml files and will update the controller as those files change (This is used for testing
// purposes). Otherwise, a CRD client is created based on the configuration.
type ConfigArgs struct {
	ClusterRegistriesNamespace string
	KubeConfig                 string
	FileDir                    string
	DisableInstallCRDs         bool

	// Controller if specified, this controller overrides the other config settings.
	Controller model.ConfigStoreCache
}

// ConsulArgs provides configuration for the Consul service registry.
type ConsulArgs struct {
	Config    string
	ServerURL string
	Interval  time.Duration
}

// ServiceArgs provides the composite configuration for all service registries in the system.
type ServiceArgs struct {
	Registries []string
	Consul     ConsulArgs
}

// PilotArgs provides all of the configuration parameters for the Pilot discovery service.
type PilotArgs struct {
	DiscoveryOptions         envoy.DiscoveryServiceOptions
	Namespace                string
	Mesh                     MeshArgs
	Config                   ConfigArgs
	Service                  ServiceArgs
	DomainSuffix             string
	MeshConfig               *meshconfig.MeshConfig
	NetworksConfigFile       string
	CtrlZOptions             *ctrlz.Options
	Plugins                  []string
	MCPMaxMessageSize        int
	MCPInitialWindowSize     int
	MCPInitialConnWindowSize int
	KeepaliveOptions         *istiokeepalive.Options
	// ForceStop is set as true when used for testing to make the server stop quickly
	ForceStop bool
}

// Server contains the runtime configuration for the Pilot discovery service.
type Server struct {
	HTTPListeningAddr       net.Addr
	GRPCListeningAddr       net.Addr
	SecureGRPCListeningAddr net.Addr
	MonitorListeningAddr    net.Addr

	EnvoyXdsServer    *envoyv2.DiscoveryServer
	ServiceController *aggregate.Controller

	Mesh         *meshconfig.MeshConfig
	MeshNetworks *meshconfig.MeshNetworks

	ConfigStores []model.ConfigStoreCache

	// Underlying config stores. To simplify, this is a configaggregate instance, created just before
	// start from the configStores
	ConfigController model.ConfigStoreCache

	// Interface abstracting all config operations, including the high-level objects
	// and the low-level untyped model.ConfigStore
	IstioConfigStore model.IstioConfigStore

	startFuncs       []startFunc
	httpServer       *http.Server
	grpcServer       *grpc.Server
	secureHTTPServer *http.Server
	secureGRPCServer *grpc.Server

	mux         *http.ServeMux
	fileWatcher filewatcher.FileWatcher
	Args        *PilotArgs

	CertKey      []byte
	CertChain    []byte
	RootCA       []byte
	Galley       *GalleyServer
	grpcListener net.Listener
	httpListener net.Listener
	Environment  *model.Environment

	// basePort defaults to 15000, used to allow multiple control plane instances on same machine
	// for testing.
	basePort     int32
}

var podNamespaceVar = env.RegisterStringVar("POD_NAMESPACE", "", "")

// NewServer creates a new Server instance, using defaults for combined Istio and loading optional mesh config
// file.
//
//
func NewServer(args *PilotArgs) (*Server, error) {

	// If the namespace isn't set, try looking it up from the environment.
	if args.Namespace == "" {
		args.Namespace = podNamespaceVar.Get()
	}
	if args.KeepaliveOptions == nil {
		args.KeepaliveOptions = istiokeepalive.DefaultOption()
	}
	if args.Config.ClusterRegistriesNamespace == "" {
		if args.Namespace != "" {
			args.Config.ClusterRegistriesNamespace = args.Namespace
		} else {
			args.Config.ClusterRegistriesNamespace = constants.IstioSystemNamespace
		}
	}

	s := &Server{
		Args: args,
	}
	s.fileWatcher = filewatcher.NewWatcher()
	return s, nil
}

// InitConfig will initialize the ConfigStores.
func (s *Server) InitConfig() error {
	prometheus.EnableHandlingTimeHistogram()
	args := s.Args

	if err := s.initMesh(args); err != nil {
		return fmt.Errorf("mesh: %v", err)
	}
	if err := s.initMeshNetworks(args); err != nil {
		return fmt.Errorf("mesh networks: %v", err)
	}

	// MCP controllers - currently using localhost by default or configured addresses.
	// This is used for config.
	if err := s.initConfigController(args); err != nil {
		return fmt.Errorf("config controller: %v", err)
	}
	return nil
}

// InitDiscovery is called after InitConfig, will initialize the discovery services and
// discovery server.
func (s *Server) InitDiscovery() error {
	args := s.Args

	// Wrap the config controller with a cache.
	configController, err := configaggregate.MakeCache(s.ConfigStores)
	if err != nil {
		return err
	}

	// Update the config controller
	s.ConfigController = configController
	// Create the config store.
	s.IstioConfigStore = model.MakeIstioStore(s.ConfigController)

	s.ServiceController = aggregate.NewController()
	// Defer running of the service controllers until Start is called, init may add more.
	s.AddStartFunc(func(stop <-chan struct{}) error {
		go s.ServiceController.Run(stop)
		return nil
	})

	s.Environment = &model.Environment{
		Mesh:             s.Mesh,
		MeshNetworks:     s.MeshNetworks,
		IstioConfigStore: s.IstioConfigStore,
		ServiceDiscovery: s.ServiceController,
		PushContext:      model.NewPushContext(),
	}

	// ServiceEntry from config and aggregate discovery in s.ServiceController
	// This will use the istioConfigStore and ConfigController.
	if err := s.addConfig2ServiceEntry(args); err != nil {
		return fmt.Errorf("service controllers: %v", err)
	}

	return nil
}

// initMonitor initializes the configuration for the pilot monitoring server.
func (s *Server) initMonitor(args *PilotArgs) error { //nolint: unparam
	s.AddStartFunc(func(stop <-chan struct{}) error {
		monitor, addr, err := startMonitor(args.DiscoveryOptions.MonitoringAddr, s.mux)
		if err != nil {
			return err
		}
		s.MonitorListeningAddr = addr

		go func() {
			<-stop
			err := monitor.Close()
			log.Debugf("Monitoring server terminated: %v", err)
		}()
		return nil
	})
	return nil
}

// Start starts all components of the Pilot discovery service on the port specified in DiscoveryServiceOptions.
// If Port == 0, a port number is automatically chosen. Content serving is started by this method,
// but is executed asynchronously. Serving can be canceled at any time by closing the provided stop channel.
func (s *Server) Start(stop <-chan struct{}, onXDSStart func(model.XDSUpdater)) error {

	// grpc, http listeners and XDS service added to grpc
	if err := s.initDiscoveryService(s.Args, onXDSStart); err != nil {
		return fmt.Errorf("discovery service: %v", err)
	}

	// TODO: at the end, moved to common (same as ctrlz)
	if err := s.initMonitor(s.Args); err != nil {
		return err
	}

	err := s.Galley.Start()
	if err != nil {
		return err
	}

	// Now start all of the components.
	for _, fn := range s.startFuncs {
		if err := fn(stop); err != nil {
			return err
		}
	}

	if !s.waitForCacheSync(stop) {
		return nil
	}

	// Start the XDS server (non blocking)
	s.EnvoyXdsServer.Start(stop)

	log.Infof("starting discovery service at http=%s grpc=%s", s.httpListener.Addr(), s.grpcListener.Addr())

	go func() {
		<-stop
//		authn_model.JwtKeyResolver.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			log.Warna(err)
		}
		if s.Args.ForceStop {
			s.grpcServer.Stop()
		} else {
			s.grpcServer.GracefulStop()
		}
	}()

	go func() {
		if err := s.httpServer.Serve(s.httpListener); err != nil {
			log.Warna(err)
		}
	}()
	go func() {
		if err := s.grpcServer.Serve(s.grpcListener); err != nil {
			log.Warna(err)
		}
	}()

	if s.Args.CtrlZOptions != nil {
		ctrlz.Run(s.Args.CtrlZOptions, []fw.Topic{s.Galley.ConfigZTopic()})
	}
	return nil
}

// startFunc defines a function that will be used to start one or more components of the Pilot discovery service.
type startFunc func(stop <-chan struct{}) error

// initMesh creates the mesh in the pilotConfig from the input arguments.
func (s *Server) initMesh(args *PilotArgs) error {
	var meshConfig *meshconfig.MeshConfig
	var err error

	// Mesh config is required - this is the primary source of config.
	meshConfig, err = cmd.ReadMeshConfig(args.Mesh.ConfigFile)
	if err != nil {
		log.Fatalf("failed to read mesh configuration, exit: %v", err)
	}

	// Watch the config file for changes and reload if it got modified
	s.addFileWatcher(args.Mesh.ConfigFile, func() {
		// Reload the config file
		meshConfig, err = cmd.ReadMeshConfig(args.Mesh.ConfigFile)
		if err != nil {
			log.Warnf("failed to read mesh configuration, using default: %v", err)
			return
		}
		if !reflect.DeepEqual(meshConfig, s.Mesh) {
			log.Infof("mesh configuration updated to: %s", spew.Sdump(meshConfig))
			if !reflect.DeepEqual(meshConfig.ConfigSources, s.Mesh.ConfigSources) {
				log.Infof("mesh configuration sources have changed")
				//TODO Need to re-create or reload initConfigController()
			}
			s.Mesh = meshConfig
			if s.EnvoyXdsServer != nil {
				s.EnvoyXdsServer.Env.Mesh = meshConfig
				s.EnvoyXdsServer.ConfigUpdate(&model.PushRequest{Full: true})
			}
		}
	})

	log.Infof("mesh configuration %s", spew.Sdump(meshConfig))
	log.Infof("version %s", version.Info.String())
	log.Infof("flags %s", spew.Sdump(args))

	s.Mesh = meshConfig
	return nil
}

// initMeshNetworks loads the mesh networks configuration from the file provided
// in the args and add a watcher for changes in this file.
func (s *Server) initMeshNetworks(args *PilotArgs) error { //nolint: unparam
	if args.NetworksConfigFile == "" {
		log.Info("mesh networks configuration not provided")
		return nil
	}

	var meshNetworks *meshconfig.MeshNetworks
	var err error

	meshNetworks, err = cmd.ReadMeshNetworksConfig(args.NetworksConfigFile)
	if err != nil {
		log.Warnf("failed to read mesh networks configuration from %q. using default.", args.NetworksConfigFile)
		return nil
	}
	log.Infof("mesh networks configuration %s", spew.Sdump(meshNetworks))
	util.ResolveHostsInNetworksConfig(meshNetworks)
	log.Infof("mesh networks configuration post-resolution %s", spew.Sdump(meshNetworks))
	s.MeshNetworks = meshNetworks

	// Watch the networks config file for changes and reload if it got modified
	s.addFileWatcher(args.NetworksConfigFile, func() {
		// Reload the config file
		meshNetworks, err := cmd.ReadMeshNetworksConfig(args.NetworksConfigFile)
		if err != nil {
			log.Warnf("failed to read mesh networks configuration from %q", args.NetworksConfigFile)
			return
		}
		if !reflect.DeepEqual(meshNetworks, s.MeshNetworks) {
			log.Infof("mesh networks configuration file updated to: %s", spew.Sdump(meshNetworks))
			util.ResolveHostsInNetworksConfig(meshNetworks)
			log.Infof("mesh networks configuration post-resolution %s", spew.Sdump(meshNetworks))
			s.MeshNetworks = meshNetworks

			// TODO
			//if s.kubeRegistry != nil {
			//	s.kubeRegistry.InitNetworkLookup(meshNetworks)
			//}
			// TODO
			//if s.Multicluster != nil {
			//	s.multicluster.ReloadNetworkLookup(meshNetworks)
			//}
			if s.EnvoyXdsServer != nil {
				s.EnvoyXdsServer.Env.MeshNetworks = meshNetworks
				s.EnvoyXdsServer.ConfigUpdate(&model.PushRequest{Full: true})
			}
		}
	})

	return nil
}

func (s *Server) initMCPConfigController(args *PilotArgs) error {
	clientNodeID := ""
	collections := make([]sink.CollectionOptions, len(schemas.Istio))
	for i, t := range schemas.Istio {
		collections[i] = sink.CollectionOptions{Name: t.Collection, Incremental: false}
	}

	options := coredatamodel.Options{
		DomainSuffix: args.DomainSuffix,
		ClearDiscoveryServerCache: func(configType string) {
			s.EnvoyXdsServer.ConfigUpdate(&model.PushRequest{Full: true})
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	var clients []*sink.Client
	var conns []*grpc.ClientConn

	reporter := monitoring.NewStatsContext("pilot")

	for _, configSource := range s.Mesh.ConfigSources {
		// Pilot connection to MCP is handled by the envoy sidecar if out of process, so we can use SDS
		// and envoy features.
		securityOption := grpc.WithInsecure()

		keepaliveOption := grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    args.KeepaliveOptions.Time,
			Timeout: args.KeepaliveOptions.Timeout,
		})

		initialWindowSizeOption := grpc.WithInitialWindowSize(int32(args.MCPInitialWindowSize))
		initialConnWindowSizeOption := grpc.WithInitialConnWindowSize(int32(args.MCPInitialConnWindowSize))
		msgSizeOption := grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(args.MCPMaxMessageSize))

		conn, err := grpc.DialContext(
			ctx, configSource.Address,
			securityOption, msgSizeOption, keepaliveOption, initialWindowSizeOption, initialConnWindowSizeOption)
		if err != nil {
			log.Errorf("Unable to dial MCP Server %q: %v", configSource.Address, err)
			cancel()
			return err
		}

		mcpController := coredatamodel.NewController(options)
		sinkOptions := &sink.Options{
			CollectionOptions: collections,
			Updater:           mcpController,
			ID:                clientNodeID,
			Reporter:          reporter,
		}

		cl := mcpapi.NewResourceSourceClient(conn)
		mcpClient := sink.NewClient(cl, sinkOptions)
		// Configz already registered by galley
		// configz.Register(mcpClient)
		clients = append(clients, mcpClient)

		conns = append(conns, conn)
		s.ConfigStores = append(s.ConfigStores, mcpController)
	}

	s.AddStartFunc(func(stop <-chan struct{}) error {
		var wg sync.WaitGroup

		for i := range clients {
			client := clients[i]
			wg.Add(1)
			go func() {
				client.Run(ctx)
				wg.Done()
			}()
		}

		go func() {
			<-stop

			// Stop the MCP clients and any pending connection.
			cancel()

			// Close all of the open grpc connections once the mcp
			// client(s) have fully stopped.
			wg.Wait()
			for _, conn := range conns {
				_ = conn.Close() // nolint: errcheck
			}

			_ = reporter.Close()
		}()

		return nil
	})

	return nil
}

// initConfigController creates the config controller in the pilotConfig.
func (s *Server) initConfigController(args *PilotArgs) error {

	if len(s.Mesh.ConfigSources) > 0 {
		if err := s.initMCPConfigController(args); err != nil {
			return err
		}
	}

	// Defer starting the controller until after the service is created.
	s.AddStartFunc(func(stop <-chan struct{}) error {
		go s.ConfigController.Run(stop)
		return nil
	})

	// Create the config store.
	s.IstioConfigStore = model.MakeIstioStore(s.ConfigController)

	return nil
}

// addConfig2ServiceEntry creates and initializes the ServiceController used for translating
// ServiceEntries from config store to discovery.
func (s *Server) addConfig2ServiceEntry(args *PilotArgs) error {
	serviceEntryStore := external.NewServiceDiscovery(s.ConfigController, s.IstioConfigStore)

	// add service entry registry to aggregator by default
	serviceEntryRegistry := aggregate.Registry{
		Name:             "ServiceEntries",
		Controller:       serviceEntryStore,
		ServiceDiscovery: serviceEntryStore,
	}
	s.ServiceController.AddRegistry(serviceEntryRegistry)

	return nil
}

func (s *Server) initDiscoveryService(args *PilotArgs, onXDSStart func(model.XDSUpdater)) error {

	// Set up discovery service
	discovery, err := envoy.NewDiscoveryService(
		s.Environment,
		args.DiscoveryOptions,
	)
	if err != nil {
		return fmt.Errorf("failed to create discovery service: %v", err)
	}
	s.mux = discovery.RestContainer.ServeMux

	s.EnvoyXdsServer = envoyv2.NewDiscoveryServer(s.Environment,
		istio_networking.NewConfigGenerator(args.Plugins),
		s.ServiceController, nil, s.ConfigController)

	if onXDSStart != nil {
		onXDSStart(s.EnvoyXdsServer)
	}

	// This is  the XDSUpdater

	s.EnvoyXdsServer.InitDebug(s.mux, s.ServiceController)

	// create grpc/http server
	s.initGrpcServer(args.KeepaliveOptions)
	s.httpServer = &http.Server{
		Addr:    args.DiscoveryOptions.HTTPAddr,
		Handler: s.mux,
	}

	// create http listener
	listener, err := net.Listen("tcp", args.DiscoveryOptions.HTTPAddr)
	if err != nil {
		return err
	}
	s.HTTPListeningAddr = listener.Addr()
	s.httpListener = listener

	// create grpc listener
	grpcListener, err := net.Listen("tcp", args.DiscoveryOptions.GrpcAddr)
	if err != nil {
		return err
	}
	s.grpcListener = grpcListener
	s.GRPCListeningAddr = grpcListener.Addr()

	return nil
}

func (s *Server) initGrpcServer(options *istiokeepalive.Options) {
	grpcOptions := s.grpcServerOptions(options)
	s.grpcServer = grpc.NewServer(grpcOptions...)
	s.EnvoyXdsServer.Register(s.grpcServer)
}

func (s *Server) grpcServerOptions(options *istiokeepalive.Options) []grpc.ServerOption {
	interceptors := []grpc.UnaryServerInterceptor{
		// setup server prometheus monitoring (as final interceptor in chain)
		prometheus.UnaryServerInterceptor,
	}

	// Temp setting, default should be enough for most supported environments. Can be used for testing
	// envoy with lower values.
	maxStreams := features.MaxConcurrentStreams

	grpcOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(middleware.ChainUnaryServer(interceptors...)),
		grpc.MaxConcurrentStreams(uint32(maxStreams)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:                  options.Time,
			Timeout:               options.Timeout,
			MaxConnectionAge:      options.MaxServerConnectionAge,
			MaxConnectionAgeGrace: options.MaxServerConnectionAgeGrace,
		}),
	}

	return grpcOptions
}

func (s *Server) AddStartFunc(fn startFunc) {
	s.startFuncs = append(s.startFuncs, fn)
}

// Add to the FileWatcher the provided file and execute the provided function
// on any change event for this file.
// Using a debouncing mechanism to avoid calling the callback multiple times
// per event.
func (s *Server) addFileWatcher(file string, callback func()) {
	_ = s.fileWatcher.Add(file)
	go func() {
		var timerC <-chan time.Time
		for {
			select {
			case <-timerC:
				timerC = nil
				callback()
			case <-s.fileWatcher.Events(file):
				// Use a timer to debounce configuration updates
				if timerC == nil {
					timerC = time.After(100 * time.Millisecond)
				}
			}
		}
	}()
}

func (s *Server) waitForCacheSync(stop <-chan struct{}) bool {
	// TODO: remove dependency on k8s lib
	for {
		if !s.ConfigController.HasSynced() {
			time.Sleep(200 * time.Millisecond)
		}
		return true
	}
	log.Debug("Sync done")
	return true
}
