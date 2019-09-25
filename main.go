package main

import (
	"bitcoin/net"
	"context"
)

func main() {
	net.StartLookUp(context.Background())
}
