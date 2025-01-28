# Testlibs for Go-Waku Bindings

`testlibs` is a dedicated testing framework for the [Go-Waku Bindings](https://github.com/waku-org/waku-go-bindings). It provides an organized structure for writing and executing tests to validate the behavior of Waku nodes and their associated functionalities.

## Overview

The primary goal of `testlibs` is to simplify and enhance the testing process for Go-Waku Bindings by offering:

- **Test Wrappers**: High-level abstractions for managing Waku nodes, including operations like starting, stopping, and destroying nodes.
- **Peer Management**: Tools for validating peer-to-peer connections, such as connecting, disconnecting, and verifying connected peers.
- **Relay Protocol Testing**: Functions to test relay features like subscribing and unsubscribing to pubsub topics and checking relay peer connections.
- **Utility Functions**: Logging and helper functions to streamline debugging and test execution.

## Key Features

- **Node Lifecycle Management**: Easily manage the lifecycle of Waku nodes (start, stop, destroy) during tests.
- **Peer Connectivity Testing**: Validate Waku node connections, disconnections, and peer interactions.
- **Relay Subscription Verification**: Test the behavior of the relay protocol and ensure proper topic subscription and peer communication.
- **Seamless Integration**: Built to work with Go's `testing` package and `testify` for straightforward assertions and test setup.

## Purpose

This framework is designed to ensure the reliability and robustness of the Go-Waku Bindings, which enable Go applications to interface with the Waku protocol. With `testlibs`, developers can simulate various conditions, verify expected behaviors, and maintain confidence in their implementations.
