module sudachen.xyz/pkg/localtest

go 1.16

replace sudachen.xyz/pkg/errstr => ../virtual/errstr

replace sudachen.xyz/pkg/localnet => ../localnet

replace github.com/spacemeshos/go-spacemesh => ../virtual/go-spacemesh

replace github.com/spacemeshos/api/release/go => ../virtual/api/release/go

replace github.com/spacemeshos/ed25519 => ../virtual/ed25519

replace github.com/spacemeshos/poet => ../virtual/poet

replace github.com/spacemeshos/post => ../virtual/post

require (
	github.com/olivere/elastic/v7 v7.0.24
	github.com/spacemeshos/ed25519 v0.0.0-20200604074309-d72da3b5f487
	github.com/spacemeshos/go-spacemesh v0.1.17
	sudachen.xyz/pkg/errstr v0.0.0-00010101000000-000000000000
	sudachen.xyz/pkg/localnet v0.0.0-00010101000000-000000000000
)
