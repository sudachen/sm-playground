module sudachen.xyz/pkg/localnet

go 1.16

replace sudachen.xyz/pkg/errstr => ../virtual/errstr

replace github.com/spacemeshos/go-spacemesh => ../virtual/go-spacemesh

replace github.com/spacemeshos/api/release/go => ../virtual/api/release/go

replace github.com/spacemeshos/ed25519 => ../virtual/ed25519

replace github.com/spacemeshos/poet => ../virtual/poet

replace github.com/spacemeshos/post => ../virtual/post

require (
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/containerd/containerd v1.4.4 // indirect
	github.com/davecgh/go-xdr v0.0.0-20161123171359-e6a2ba005892
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v20.10.5+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0 // indirect
	github.com/fullstorydev/grpcurl v1.8.1
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spacemeshos/api/release/go v0.0.0-20210605150702-81ae866b0342
	github.com/spacemeshos/ed25519 v0.0.0-20190530014421-e235766d15a1
	github.com/spacemeshos/go-spacemesh v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.1.3
	golang.org/x/crypto v0.0.0-20201208171446-5f87f3452ae9
	golang.org/x/net v0.0.0-20210428140749-89ef3d95e781
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	google.golang.org/genproto v0.0.0-20201007142714-5c0e72c5e71e
	google.golang.org/grpc v1.37.0
	gotest.tools/v3 v3.0.3 // indirect
	sudachen.xyz/pkg/errstr v0.0.0-00010101000000-000000000000
)
