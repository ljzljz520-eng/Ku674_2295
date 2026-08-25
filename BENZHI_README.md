# Ku674_2295

Go project for module `farm-sensor-platform`.

## Standard commands

```bash
GOTOOLCHAIN=local go build ./...
GOTOOLCHAIN=local go test -count=1 ./...
GOTOOLCHAIN=local go test -race -count=20 ./...
```

## Run

```bash
GOTOOLCHAIN=local go run ./cmd/farmserver -listen :8080 -db farm-sensors.db
```

The service handles fixed-fixture vendor adapters and exposes `GET /health`, `GET /v1/batches`, `GET /v1/pending`, `GET /v1/dashboard`, `GET /v1/report`, `GET /v1/fields`, `GET /v1/trends`, and `GET /v1/alerts`.

`go.etcd.io/bbolt` stores readings, batches, pending items, and audit entries. `TestPersistenceSurvivesReopen` verifies the close/reopen lifecycle. `TestSensorBatchKeepsFailures` is the intentional lost-update regression; its expected failure is recorded in `BUG_REPRO.md`.

## Docker validation

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh my-go-task linux/arm64
./build_benzhi_docker.sh my-go-task linux/amd64
docker run -it my-go-task:latest
```

## Known initial failures

See `BUG_REPRO.md` for the exact command and output captured during packaging.
