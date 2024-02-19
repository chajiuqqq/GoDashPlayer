module dashquic

go 1.21

//replace (
//	github.com/lucas-clemente/quic-go/h2quic => github.com/quic-go/quic-go v0.41.0
//)

require (
	github.com/quic-go/quic-go v0.41.0
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/net v0.21.0
)

require (
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38 // indirect
	github.com/onsi/ginkgo/v2 v2.9.5 // indirect
	github.com/quic-go/qpack v0.4.0 // indirect
	go.uber.org/mock v0.3.0 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	golang.org/x/exp v0.0.0-20221205204356-47842c84f3db // indirect
	golang.org/x/mod v0.11.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.9.1 // indirect
)
