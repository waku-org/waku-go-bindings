# Waku Go Bindings

This repository provides Go bindings for the Waku library, enabling seamless integration with Go projects.

## Installation

To build the required dependencies for this module, the `make` command needs to be executed. If you are integrating this module into another project via `go get`, ensure that you navigate to the `waku-go-bindings` module directory and run `make`.

### Steps to Install

Follow these steps to install and set up the module:

1. Retrieve the module using `go get`:
   ```
   go get -u github.com/waku-org/waku-go-bindings
   ```
2. Navigate to the module's directory:
   ```
   cd $(go list -m -f '{{.Dir}}' github.com/waku-org/waku-go-bindings)
   ```
3. Build the dependencies:
   ```
   sudo make
   ```

Now the module is ready for use in your project.
