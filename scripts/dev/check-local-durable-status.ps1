$ErrorActionPreference = "Stop"

$env:OPS_BASE_URL = if ($env:OPS_BASE_URL) { $env:OPS_BASE_URL } else { "http://localhost:8088" }
$env:REQUIRE_MYSQL_SUMMARY = if ($env:REQUIRE_MYSQL_SUMMARY) { $env:REQUIRE_MYSQL_SUMMARY } else { "true" }
$env:REQUIRE_REDIS_SUMMARY = if ($env:REQUIRE_REDIS_SUMMARY) { $env:REQUIRE_REDIS_SUMMARY } else { "true" }
$env:EXPECTED_MYSQL_SERVICES = if ($env:EXPECTED_MYSQL_SERVICES) { $env:EXPECTED_MYSQL_SERVICES } else { "identity,social,invite,chat,party,guild" }

go run ./scripts/dev/cmd/check_local_durable_status
