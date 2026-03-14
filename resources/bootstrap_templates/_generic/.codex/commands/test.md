# /test [scope]

Use `go-test-runner` to run tests/lint in the minimal order.
Default scope: all (`go test ./...`).
Return:
- failing packages/tests only
- first relevant error lines
- minimal fix hypothesis
