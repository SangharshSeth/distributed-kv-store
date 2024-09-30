# Distributed Key-Value Store Package

This package implements a distributed key-value store that supports basic CRUD operations, partitioning, concurrent access control, and persistence through an append-only log (AOF). It can be imported and used in Go applications to manage key-value data efficiently.

## Features

- **Partitioning**: Data is divided into partitions, with keys hashed to specific partitions for balanced distribution.
- **Concurrency**: Read-Write mutexes ensure thread-safe access to each partition.
- **Persistence**: Commands are logged to an Append-Only File (AOF) for recovery after restarts.
- **Statistics**: Tracks connection and command metrics.

## Installation

To install the package, use the following command:

```bash
go get github.com/SangharshSeth/distributed-kv-store
```
