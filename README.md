# Waku Go Bindings

Go bindings for the Waku library.

## Install

```
go get -u github.com/waku-org/waku-go-bindings
```

## Dependencies

This repository doesn't download or build `nwaku`. You must provide `libwaku` and its headers.

To do so, you can:
- Build `libwaku` from https://github.com/waku-org/nwaku.
- Point `cgo` to the headers and compiled library when building your project.

Example environment setup (adjust paths to your nwaku checkout):
```
export NWAKU_DIR=/path/to/nwaku
export CGO_CFLAGS="-I${NWAKU_DIR}/library"
export CGO_LDFLAGS="-L${NWAKU_DIR}/build -lwaku -Wl,-rpath,${NWAKU_DIR}/build"
```

Such setup would look like this in a `Makefile`:
```Makefile
NWAKU_DIR ?= /path/to/nwaku
CGO_CFLAGS = -I$(NWAKU_DIR)/library
CGO_LDFLAGS = -L$(NWAKU_DIR)/build -lwaku -Wl,-rpath,$(NWAKU_DIR)/build

build: ## Your project build command
	go build ./...
```

For a reference integration, see how `status-go` wires `CGO_CFLAGS` and `CGO_LDFLAGS` in its build setup.

NOTE: If your project is itself used as a Go dependency, all its clients will have to follow the same nwaku setup. 

## Development

When working on this repository itself, `nwaku` is included as a git submodule for convenience.

- Initialize and update the submodule, then build `libwaku`
    ```sh
    git submodule update --init --recursive
    make -C waku build-libwaku
    ```
- Build the project. Submodule paths are used by default to find `libwaku`.
    ```shell
    make -C waku build
    ```
