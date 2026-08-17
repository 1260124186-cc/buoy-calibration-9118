# Buoy Calibration Service

Buoy Calibration Service is a small HTTP service for ocean-observation teams.
It keeps calibration profiles for buoy sensors, records a bounded set of
measurement samples for each field run, and seals a run into a correction
report that can be consumed by downstream research tools.

## Workflows

1. Create a calibration profile with a scale and bias for one sensor type.
2. Open a field run and append timestamped observations from that profile.
3. Seal the run to calculate corrected minimum, maximum, and average values.

## Layout

- `cmd/server`: HTTP server entry point.
- `internal/api`: request parsing and JSON responses.
- `internal/calibration`: correction and report calculations.
- `internal/model`: shared domain types and validation errors.
- `internal/service`: workflow orchestration.
- `internal/store`: concurrency-safe in-memory repository.

## Run

```bash
go run ./cmd/server
```

The server listens on `:8080` by default. Set `BUOY_ADDR` to override the
address.

## API

```text
POST /profiles
POST /runs
POST /runs/{id}/samples
POST /runs/{id}/seal
GET  /runs/{id}/report
GET  /healthz
```

All state is kept in memory so a field team can use the service for a single
calibration session without external infrastructure.

## Verify

```bash
go build ./...
go test ./...
```
