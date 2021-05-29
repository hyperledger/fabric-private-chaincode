## Prerequisites:

* run the FPC build

## Worker

### Build it!

```
make build
```

### Run it!

Note that the worker also needs to connect to the storage service in order to collect the patient data.
Therefore, you can define the redis endpoint via `REDIS_HOST` and `REDIS_PORT`.

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
export WORKER_HOST=host.docker.internal
export WORKER_PORT=5000
```