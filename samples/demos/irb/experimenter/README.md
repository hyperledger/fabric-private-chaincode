## Prerequisites:

* run the FPC build

## Worker

### Build it!

```
make build
```

### Run it!

```
cd worker
make docker-run
```

## Client Service

Start the service; by default it listens on 3001

```
cd client/service
go run.
```

The client connects to a worker. You can define the worker host and port by setting:
```
export WORKER_HOST=REDIS_HOST=host.docker.internal
export WORKER_PORT=5000
```