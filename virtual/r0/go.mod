module sudachen.xyz/pkg/r0

go 1.16

replace github.com/spacemeshos/go-spacemesh => ../go-spacemesh-overwrite

replace github.com/spacemeshos/ed25519 => ../ed25519

replace github.com/spacemeshos/fixed => ../fixed

replace github.com/spacemeshos/api/release/go => ../api/release/go

require (
	github.com/spacemeshos/fixed v0.0.0-00010101000000-000000000000 // indirect
	github.com/spacemeshos/go-spacemesh v0.0.0-00010101000000-000000000000
)
