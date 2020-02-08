module github.com/costinm/istiod

go 1.13

replace istio.io/istio => /ws/istio-master/go/src/istio.io/istio

//replace istio.io/istio => github.com/costinm/istio v0.0.0-20191022214508-382391b5279c

// Reminder cut&pasted from istio/istio
replace github.com/golang/glog => github.com/istio/glog v0.0.0-20190424172949-d7cfb6fa2ccd

replace k8s.io/klog => github.com/istio/klog v0.0.0-20190424230111-fb7481ea8bcf

replace github.com/spf13/viper => github.com/istio/viper v1.3.3-0.20190515210538-2789fed3109c

replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20191011211953-adfac697dc5b

require (
	cloud.google.com/go v0.50.0
	cloud.google.com/go/logging v1.0.0
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	contrib.go.opencensus.io/exporter/stackdriver v0.12.9
	contrib.go.opencensus.io/exporter/zipkin v0.1.1
	fortio.org/fortio v1.3.1
	github.com/DataDog/datadog-go v2.2.0+incompatible
	github.com/Masterminds/semver v1.4.2
	github.com/alicebob/miniredis v0.0.0-20180201100744-9d52b1fc8da9
	github.com/aws/aws-sdk-go v1.23.20
	github.com/cactus/go-statsd-client v3.1.1+incompatible
	github.com/cenkalti/backoff v2.0.0+incompatible
	github.com/census-instrumentation/opencensus-proto v0.2.1
	github.com/circonus-labs/circonus-gometrics v2.3.1+incompatible
	github.com/cncf/udpa/go v0.0.0-20191209042840-269d4d468f6f
	github.com/coreos/etcd v3.3.15+incompatible
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/envoyproxy/go-control-plane v0.9.2
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/fluent/fluent-logger-golang v1.3.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/go-redis/redis v6.10.2+incompatible
	github.com/gogo/protobuf v1.3.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f
	github.com/google/cel-go v0.2.0
	github.com/google/go-cmp v0.3.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/googleapis/gax-go v2.0.2+incompatible
	github.com/googleapis/gax-go/v2 v2.0.5
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.1-0.20190118093823-f849b5445de4
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20171214222146-0e7658f8ee99
	github.com/hashicorp/consul v1.3.1
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/vault/api v1.0.3
	github.com/howeyc/fsnotify v0.9.0
	github.com/istio.io/proxy/src/envoy/tcp/metadata_exchange/config v0.0.0
	github.com/kr/pretty v0.1.0
	github.com/kylelemons/godebug v1.1.0
	github.com/lestrrat-go/jwx v0.9.0
	github.com/mattn/go-isatty v0.0.12
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/mitchellh/copystructure v1.0.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/onsi/gomega v1.7.1
	github.com/open-policy-agent/opa v0.8.2
	github.com/openshift/api v3.9.1-0.20191008181517-e4fd21196097+incompatible
	github.com/opentracing/opentracing-go v1.0.2
	github.com/openzipkin/zipkin-go v0.1.7
	github.com/pkg/errors v0.8.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4
	github.com/prometheus/common v0.6.0
	github.com/prometheus/prom2json v1.2.2
	github.com/prometheus/tsdb v0.7.1 // indirect
	github.com/ryanuber/go-glob v1.0.0
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/uber/jaeger-client-go v0.0.0-20190228190846-ecf2d03a9e80
	github.com/yl2chen/cidranger v0.0.0-20180214081945-928b519e5268
	go.opencensus.io v0.22.2
	go.uber.org/atomic v1.4.0
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20191014212845-da9a3fd4c582
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	golang.org/x/tools v0.0.0-20191216173652-a0e659d51361
	google.golang.org/api v0.15.0
	google.golang.org/genproto v0.0.0-20191223191004-3caeed10a8bf
	google.golang.org/grpc v1.26.0
	gopkg.in/d4l3k/messagediff.v1 v1.2.1
	gopkg.in/square/go-jose.v2 v2.3.1
	gopkg.in/yaml.v2 v2.2.7
	istio.io/api v0.0.0-20200201001607-7b905a0a6e49
	istio.io/gogo-genproto v0.0.0-20200122005450-9b171d92064b
	istio.io/istio v0.0.0-00010101000000-000000000000 // indirect
	istio.io/pkg v0.0.0-20200131182711-9ba13e0e34bb
	k8s.io/api v0.17.2
	k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/cli-runtime v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/helm v2.14.3+incompatible
	k8s.io/kubectl v0.17.2
	k8s.io/utils v0.0.0-20191114184206-e782cd3c129f
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/yaml v1.1.0
)

replace github.com/Azure/go-autorest/autorest => github.com/Azure/go-autorest/autorest v0.9.0

replace github.com/Azure/go-autorest/autorest/adal => github.com/Azure/go-autorest/autorest/adal v0.5.0

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.2.0+incompatible

replace github.com/istio.io/proxy/src/envoy/tcp/metadata_exchange/config => istio.io/istio/pilot/pkg/metadata_exchange v0.0.0-20200123191201-47e363a83438
