module fedb

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/cznic/golex v0.0.0-20181122101858-9c343928389c // indirect
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548
	github.com/cznic/parser v0.0.0-20181122101858-d773202d5b1f
	github.com/cznic/sortutil v0.0.0-20181122101858-f5f958428db8
	github.com/cznic/strutil v0.0.0-20181122101858-275e90344537
	github.com/cznic/y v0.0.0-20181122101901-b05e8c2e8d7b
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/juju/errors v0.0.0-20181118221551-089d3ea4e4d5
	github.com/opentracing/opentracing-go v1.0.2
	github.com/pingcap/errors v0.11.0 // indirect
	github.com/pingcap/goleveldb v0.0.0-20171020122428-b9ff6c35079e // indirect
	github.com/pingcap/parser v0.0.0-20190214121452-6d10a0b75f3e // indirect
	github.com/pingcap/tidb v2.0.11+incompatible
	github.com/pingcap/tipb v0.0.0-20190107072121-abbec73437b7 // indirect
	github.com/pingyu/parser v0.0.0-20190214121452-6d10a0b75f3e
	github.com/pkg/errors v0.8.1 // indirect
	github.com/prometheus/client_golang v0.9.2 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20170806203942-52369c62f446 // indirect
	github.com/sirupsen/logrus v1.3.0
	github.com/spaolacci/murmur3 v0.0.0-20180118202830-f09979ecbc72 // indirect
	github.com/uber/jaeger-client-go v2.15.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.0.0+incompatible // indirect
	golang.org/x/net v0.0.0-20190125091013-d26f9f9a57f3
	golang.org/x/text v0.3.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)

replace (
	golang.org/x/net v0.0.0-20181201002055-351d144fa1fc => github.com/golang/net v0.0.0-20181201002055-351d144fa1fc
	golang.org/x/net v0.0.0-20190125091013-d26f9f9a57f3 => github.com/golang/net v0.0.0-20190125091013-d26f9f9a57f3
	golang.org/x/text v0.3.0 => github.com/golang/text v0.3.0
)
