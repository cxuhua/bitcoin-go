module bitcoin

go 1.13

require golang.org/x/crypto v1.0.0

require (
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.1.1
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
)

replace golang.org/x/crypto v1.0.0 => github.com/cxuhua/crypto v1.0.0
