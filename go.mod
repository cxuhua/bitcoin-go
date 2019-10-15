module bitcoin

require golang.org/x/crypto v1.0.0

require (
	github.com/dchest/siphash v1.2.1
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/spaolacci/murmur3 v1.1.0
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/tidwall/pretty v1.0.0 // indirect
	github.com/willf/bitset v1.1.10
	github.com/willf/bloom v2.0.3+incompatible // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.1.2
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
)

replace golang.org/x/crypto v1.0.0 => github.com/cxuhua/crypto v1.0.0

go 1.13
