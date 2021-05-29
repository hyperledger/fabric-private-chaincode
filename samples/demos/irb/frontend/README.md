# IRB demo frontend

## Develop

Install npm before

```
make install
make run
```

## Docker

```
make docker
make docker-run
make docker-stop
```


## Configuration

To set the backend endpoints (data-provider, experimenter, and principle investigator), do:

```
cd demo
cp .env .env.local
```

Then update `.env.local` accordingly. This will should not be checked into git.
Note that you need to restart the frontend (if running).