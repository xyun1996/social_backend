# Worker HTTP Contract

Base purpose: async job queue inspection and lifecycle transitions for prototype worker execution.

## Health

- `GET /healthz`

## Enqueue Job

- `POST /v1/jobs`
- Request

```json
{
  "type": "invite.expire",
  "payload": "{\"invite_id\":\"inv-1\"}"
}
```

- Response `200`: job object

## List Jobs

- `GET /v1/jobs?status=queued&type=invite.expire`
- Response `200`

```json
{
  "status": "queued",
  "type": "invite.expire",
  "count": 1,
  "jobs": []
}
```

## Claim Job

- `POST /v1/jobs/claim`
- Request

```json
{
  "worker_id": "worker-a",
  "type": "invite.expire"
}
```

- Response `200`: claimed job object

## Complete Job

- `POST /v1/jobs/{jobID}/complete`
- Request

```json
{
  "worker_id": "worker-a"
}
```

- Response `200`: completed job object

## Fail Job

- `POST /v1/jobs/{jobID}/fail`
- Request

```json
{
  "worker_id": "worker-a",
  "last_error": "temporary failure"
}
```

- Response `200`: failed job object

## Run One Job

- `POST /v1/jobs/run-once`
- Request

```json
{
  "worker_id": "worker-a",
  "type": "invite.expire"
}
```

- Response `200`: execution summary

## Run Until Empty

- `POST /v1/jobs/run-until-empty`
- Request

```json
{
  "worker_id": "worker-a",
  "type": "invite.expire",
  "limit": 100
}
```

- Response `200`: execution summary across processed jobs

## Background Runner

- Process-level option controlled by `WORKER_AUTO_RUN=true`
- `WORKER_AUTO_RUN_INTERVAL_MS` controls the drain loop interval
