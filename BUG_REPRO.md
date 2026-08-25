# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	farm-sensor-platform/cmd/farmserver	[no test files]
ok  	farm-sensor-platform/internal/adapter	0.001s
ok  	farm-sensor-platform/internal/alerts	0.001s
ok  	farm-sensor-platform/internal/analytics	0.001s
ok  	farm-sensor-platform/internal/api	0.012s
ok  	farm-sensor-platform/internal/audit	0.013s
ok  	farm-sensor-platform/internal/config	0.001s
ok  	farm-sensor-platform/internal/model	0.001s
ok  	farm-sensor-platform/internal/normalize	0.003s
ok  	farm-sensor-platform/internal/observability	0.002s
ok  	farm-sensor-platform/internal/persistence	0.030s
ok  	farm-sensor-platform/internal/processing	0.001s
--- FAIL: TestSensorBatchKeepsFailures (0.02s)
    service_test.go:69: expected partial failure, got {ID:b-bug ReceivedAt:1970-01-01 00:05:00 +0000 UTC Readings:[{ID:b-bug:agrisense:bad BatchID:b-bug Vendor:agrisense ExternalID:bad FieldID:north CapturedAt:1970-01-01 00:01:40 +0000 UTC Moisture:180 Temperature:20 Light:300 Battery:88 Status:pending Reason:moisture value 180.00 outside permitted range} {ID:b-bug:fieldlink:good BatchID:b-bug Vendor:fieldlink ExternalID:good FieldID:north CapturedAt:1970-01-01 00:01:40 +0000 UTC Moisture:40 Temperature:20 Light:100 Battery:90 Status:accepted Reason:}] Total:2 Accepted:2 Pending:0 Failed:false Summary:accepted=2 pending=0}
FAIL
FAIL	farm-sensor-platform/internal/sensors	0.053s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/farmserver): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/farmserver): exit `0`
