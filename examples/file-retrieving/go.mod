module github.com/star_sample


go 1.23.3

toolchain go1.23.6

require (
	github.com/spinframework/spin-go-sdk/v2 v2.0.0-20250422162322-8ffe6d3efa29
	github.com/ydnar/wasi-http-go v0.0.0-20250324053847-ca78b3198aeb
)

require go.bytecodealliance.org/cm v0.2.2 // indirect
require github.com/cedweber/spin-s3-api v0.0.0

// replace github.com/ydnar/wasi-http-go => ../../ydnar/wasi-http-go
replace github.com/ydnar/wasi-http-go => github.com/rajatjindal/wasi-http-go v0.0.0-20250430163340-bf83542051da
replace github.com/cedweber/spin-s3-api => ../../
