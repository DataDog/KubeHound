module github.com/DataDog/KubeHound

go 1.19

require (
	github.com/DataDog/datadog-go v4.8.3+incompatible
	github.com/aws/aws-sdk-go-v2 v1.17.8
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.2
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.63
	github.com/aws/aws-sdk-go-v2/service/batch v1.23.3
	github.com/aws/aws-sdk-go-v2/service/s3 v1.32.0
	github.com/aws/smithy-go v1.13.5
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/inconshreveable/go-vhost v0.0.0-20160627193104-06d84117953b
	github.com/jedib0t/go-pretty/v6 v6.4.4
	github.com/spf13/cobra v1.6.1
	github.com/spf13/viper v1.15.0
	github.com/stretchr/testify v1.8.1
	github.com/stripe/goproxy v0.0.0-20220308202309-3f1dfba6d1a4
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/DataDog/dd-trace-go.v1 v1.47.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	mvdan.cc/xurls/v2 v2.4.0
)

require (
	github.com/DataDog/datadog-agent/pkg/obfuscate v0.0.0-20211129110424-6491aa3bf583 // indirect
	github.com/DataDog/datadog-agent/pkg/remoteconfig/state v0.42.0-rc.1 // indirect
	github.com/DataDog/datadog-go/v5 v5.1.1 // indirect
	github.com/DataDog/go-tuf v0.3.0--fix-localmeta-fork // indirect
	github.com/DataDog/gostackparse v0.5.0 // indirect
	github.com/DataDog/sketches-go v1.3.0 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.20 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.32 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.26 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.33 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.24 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.27 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.26 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.14.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.9 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/elazarl/goproxy v0.0.0-20190911111923-ecfe977594f1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-acme/lego/v3 v3.1.0 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/pprof v0.0.0-20210720184732-4bb14d4b1be1 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/miekg/dns v1.1.46 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mwitkow/go-http-dialer v0.0.0-20161116154839-378f744fb2b8 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/richardartoul/molecule v1.0.1-0.20221107223329-32cfee06a052 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.4.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/tidwall/btree v1.1.0 // indirect
	github.com/tidwall/buntdb v1.2.9 // indirect
	github.com/tidwall/gjson v1.12.1 // indirect
	github.com/tidwall/grect v0.1.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/rtred v0.1.2 // indirect
	github.com/tidwall/tinyqueue v0.1.1 // indirect
	github.com/tinylib/msgp v1.1.6 // indirect
	go4.org/intern v0.0.0-20211027215823-ae77deb06f29 // indirect
	go4.org/unsafe/assume-no-moving-gc v0.0.0-20220617031537-928513b29760 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/mod v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/time v0.1.0 // indirect
	golang.org/x/tools v0.4.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	inet.af/netaddr v0.0.0-20220617031823-097006376321 // indirect
)

require (
	github.com/aws/aws-sdk-go-v2/config v1.18.21
	github.com/aws/aws-sdk-go-v2/service/kms v1.20.10
	github.com/aws/aws-sdk-go-v2/service/sesv2 v1.17.4
	github.com/google/go-cmp v0.5.9
	github.com/kgretzky/evilginx2 v0.0.0-20210208221752-511860ca993b
	github.com/qmuntal/stateless v1.6.3
	github.com/sirupsen/logrus v1.9.0
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/text v0.7.0 // indirect
)
