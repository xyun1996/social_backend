Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

go test ./test/load/... -run TestHotPathSmoke -count=1 -v
