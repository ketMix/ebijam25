package main

import (
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/kettek/gobl"
)

func main() {
	var exe string
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}

	runArgs := append([]interface{}{}, "./game"+exe)
	var wasmSrc string

	Task("build").
		Exec("go", "build", "./cmd/game")
	Task("run").
		Exec(runArgs...)
	Task("watch").
		Watch("cmd/**", "internal/**", "pkg/**", "stuff/**").
		Signaler(SigQuit).
		Run("build").
		Run("run")
	Task("build-web").
		Env("GOOS=js", "GOARCH=wasm").
		Exec("go", "build", "-o", "web/ebijam25.wasm", "./cmd/game").
		Exec("go", "env", "GOROOT").
		Result(func(i interface{}) {
			goRoot := strings.TrimSpace(i.(string))
			wasmSrc = filepath.Join(goRoot, "lib/wasm/wasm_exec.js")
		}).
		Exec("cp", &wasmSrc, "./web/")

	// Mob
	Task("mob-build").
		Exec("go", "build", "./prototype/mob")
	Task("mob-run").
		Exec(append([]interface{}{}, "./mob"+exe)...)
	Task("mob-watch").
		Watch("internal/**", "pkg/**", "prototype/mob/**").
		Signaler(SigQuit).
		Run("mob-build").
		Run("mob-run")

	Go()
}
