# ABCI Example

Example of Tendermint bundled with ABCI app

## Prerequisites

- Go version >= 1.12.5

  - [Install Go](https://golang.org/dl/) by following [installation instructions.](https://golang.org/doc/install)

## Run

### Node 1

```sh
ABCI_DB_DIR_PATH=DB1 go run ./abci --home ./config/tendermint/node1 node
```

### Node 2

```sh
ABCI_DB_DIR_PATH=DB2 go run ./abci --home ./config/tendermint/node2 node
```

### Node 3

```sh
ABCI_DB_DIR_PATH=DB3 go run ./abci --home ./config/tendermint/node3 node
```

### Node 4

```sh
ABCI_DB_DIR_PATH=DB4 go run ./abci --home ./config/tendermint/node4 node
```

**Environment variable options**

- `ABCI_DB_DIR_PATH`: Directory path for ABCI app persistence data files [Default: `./DB`]
- `ABCI_DB_TYPE`: Database type (same options as Tendermint's `db_backend`) [Default: `goleveldb`]
- `ABCI_LOG_LEVEL`: Log level. Allowed values are `error`, `warn`, `info` and `debug` [Default: `debug`]
- `ABCI_LOG_TARGET`: Where should logger writes logs to. Allowed values are `console` or `file` (eg. `ABCI.log`) [Default: `console`]
- `ABCI_LOG_FILE_PATH`: File path for log file (use when `ABCI_LOG_TARGET` is set to `file`) [Default: `./abci-<PID>-<CURRENT_DATETIME>.log`]

## Test

```sh
go test ./test -count=1 -v
```

## Reset chain

```sh
rm -rf DB*/

go run ./abci --home ./config/tendermint/node1 unsafe_reset_all

go run ./abci --home ./config/tendermint/node2 unsafe_reset_all

go run ./abci --home ./config/tendermint/node3 unsafe_reset_all

go run ./abci --home ./config/tendermint/node4 unsafe_reset_all
```

## Generate go from protobuf

```sh
./scripts/protogen.sh
```