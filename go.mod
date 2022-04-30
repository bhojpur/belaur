module github.com/bhojpur/belaur

go 1.16

require (
	github.com/GeertJohan/go.rice v1.0.2
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/bhojpur/policy v1.0.0
	github.com/docker/docker v20.10.8+incompatible
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/google/go-github v17.0.0+incompatible
	github.com/hashicorp/go-hclog v1.0.0
	github.com/hashicorp/go-memdb v1.3.2
	github.com/hashicorp/go-plugin v1.4.3
	github.com/labstack/echo/v4 v4.5.0
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron v1.2.0
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/echo-swagger v1.1.3
	github.com/swaggo/swag v1.7.1
	go.etcd.io/bbolt v1.3.6
	golang.org/x/crypto v0.0.0-20220128200615-198e4374d7ed
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	google.golang.org/grpc v1.44.0
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible // indirect
	github.com/containerd/containerd v1.5.5 // indirect
	github.com/daaku/go.zipexe v1.0.1 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/spec v0.20.3 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/yamux v0.0.0-20210826001029-26ff87cf9493 // indirect
	github.com/kevinburke/ssh_config v1.1.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/swaggo/files v0.0.0-20210815190702-a29dd2bc99b2 // indirect
	github.com/xanzy/ssh-agent v0.3.1 // indirect
	google.golang.org/protobuf v1.27.1
)

replace github.com/swaggo/swag => github.com/swaggo/swag v1.6.10-0.20201104153820-3f47d68f8872

replace github.com/ugorji/go/codec => github.com/ugorji/go/codec v1.2.0

replace k8s.io/api => k8s.io/api v0.23.6

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.6

replace k8s.io/apimachinery => k8s.io/apimachinery v0.23.7-rc.0

replace k8s.io/apiserver => k8s.io/apiserver v0.23.6

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.23.6

replace k8s.io/client-go => k8s.io/client-go v0.23.6

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.23.6

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.23.6

replace k8s.io/code-generator => k8s.io/code-generator v0.23.7-rc.0

replace k8s.io/component-base => k8s.io/component-base v0.23.6

replace k8s.io/component-helpers => k8s.io/component-helpers v0.23.6

replace k8s.io/controller-manager => k8s.io/controller-manager v0.23.6

replace k8s.io/cri-api => k8s.io/cri-api v0.23.7-rc.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.23.6

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.23.6

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.23.6

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.23.6

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.23.6

replace k8s.io/kubectl => k8s.io/kubectl v0.23.6

replace k8s.io/kubelet => k8s.io/kubelet v0.23.6

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.23.6

replace k8s.io/metrics => k8s.io/metrics v0.23.6

replace k8s.io/mount-utils => k8s.io/mount-utils v0.23.7-rc.0

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.23.6

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.23.6

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.23.6

replace k8s.io/sample-controller => k8s.io/sample-controller v0.23.6
