# Testlibs for Go-Waku Bindings

`testlibs` is a dedicated testing framework for the [Go-Waku Bindings](https://github.com/waku-org/waku-go-bindings). It provides an organized structure for writing and executing tests to validate the behavior of Waku nodes and their associated functionalities.

## Overview

The primary goal of `testlibs` is to simplify and enhance the testing process for Go-Waku Bindings by offering:

- **Test Wrappers**: High-level abstractions for managing Waku nodes, including operations like starting, stopping, and destroying nodes.
- **Peer Management**: Tools for validating peer-to-peer connections, such as connecting, disconnecting, and verifying connected peers.
- **Relay Protocol Testing**: Functions to test relay features like subscribing and unsubscribing to pubsub topics and checking relay peer connections.
- **Utility Functions**: Logging and helper functions to streamline debugging and test execution.
## Key Components

### 1. Wrappers

The `wrapper` files define abstractions around the `WakuNode` to streamline node management and interaction during testing. Key functionalities include:

- **Node Lifecycle Management**:
  - `Wrappers_StartWakuNode`: Starts a Waku node with a custom or default configuration.
  - `Wrappers_Stop`: Stops the Waku node gracefully.
  - `Wrappers_Destroy`: Cleans up and destroys the Waku node.
  - `Wrappers_StopAndDestroy`: Combines stop and destroy operations for efficient cleanup.

- **Peer Management**:
  - `Wrappers_ConnectPeer`: Simplifies connecting nodes by taking a `WakuNode` instance instead of a `peerID`, constructing the `peerID` and connection details internally.
  - `Wrappers_DisconnectPeer`: Disconnects a Waku node from a target peer with validation and error handling.
  - `Wrappers_GetConnectedPeers`: Retrieves a list of peers connected to the Waku node.
  - `Wrappers_GetNumConnectedRelayPeers`: Returns the number of relay peers connected for a specific pubsub topic.

- **Relay Subscription**:
  - `Wrappers_RelaySubscribe`: Subscribes the Waku node to a given pubsub topic, ensuring subscription and verifying peer connections.
  - `Wrappers_RelayUnsubscribe`: Unsubscribes from a specific pubsub topic and validates the unsubscription process.

---

### 2. Utilities

The `utilities` package provides helper functions and constants to facilitate testing. Key highlights include:

- **Default Configuration**:
  - `DefaultWakuConfig`: Provides a baseline Waku node configuration with sensible defaults for testing purposes.
  
- **Logging Support**:
  - Centralized logging via `zap` for debugging during tests.
  - Debug-level messages are used extensively to trace the flow and identify issues.

- **Port Management**:
  - `GenerateUniquePort`: Dynamically allocates unique ports for Waku nodes during tests to avoid conflicts.

- **Timeout and Error Handling**:
  - Constants for peer connection timeouts.
  - Enhanced error messaging for debugging failures in Waku node operations.

---

### 3. Test Files

#### Relay Tests (`relay_test.go`)
- **Highlights**:
  - Tests the behavior of the relay protocol.
  - Example: `TestRelaySubscribeToDefaultTopic` validates subscription to the default pubsub topic and ensures connected relay peers increase.

#### Peer Connection Tests (`Peers_connection_test.go`)
- **Highlights**:
  - Simplifies peer connectivity testing using the wrappers.
  - Example: `TestConnectMultipleNodesToSingleNode` verifies multiple nodes can connect to a single node efficiently.

#### Basic Node Tests (`Nodes_basic_test.go`)
- **Highlights**:
  - Validates fundamental node lifecycle management using wrapper APIs.
  - Example: `TestBasicWakuNodes` covers node creation, startup, and cleanup using `Wrappers_StartWakuNode` and `Wrappers_StopAndDestroy`.

## Purpose

This framework is designed to ensure the reliability and robustness of the Go-Waku Bindings, which enable Go applications to interface with the Waku protocol. With `testlibs`, developers can simulate various conditions, verify expected behaviors, and maintain confidence in their implementations.
