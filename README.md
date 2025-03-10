# Waku Go Bindings

This repository provides Go bindings for the Waku library, enabling seamless integration with Go projects.

## Installation

To build the required dependencies for this module, the `make` command needs to be executed. If you are integrating this module into another project via `go get`, ensure that you navigate to the `waku-go-bindings` module directory and run `make`.

### Steps to Install

Follow these steps to install and set up the module:

1. Make sure your system has the [prerequisites](https://docs.waku.org/guides/nwaku/build-source#prerequisites) to run a local nwaku node

2. Retrieve the module using `go get`:
   ```
   go get -u github.com/waku-org/waku-go-bindings
   ```
3. Navigate to the module's directory:
   ```
   cd $(go list -m -f '{{.Dir}}' github.com/waku-org/waku-go-bindings)
   ```
4. Prepare third_party directory and clone nwaku
   ```
   sudo mkdir third_party
   chown $USER:$USER third_party
   ```
5. Build the dependencies:
   ```
   make -C waku
   ```

Now the module is ready for use in your project.
