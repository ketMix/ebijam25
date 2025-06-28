package main

import (
	"github.com/ketMix/ebijam25/internal/server"
)

func main() {
	var garçon server.Garçon
	garçon.Serve(9099, false)
}
