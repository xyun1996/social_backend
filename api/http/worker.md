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
