# Backend API Contract (Current Stage)

This document describes the new and updated backend endpoints added in the current stage:
- `proposals`
- `kinopoisk`
- `statistics`
- admin-only movie creation flow
- weekly pack mini app helpers (`current`, `me/limits`)

All responses are JSON.

## Auth And Roles

JWT auth uses `Authorization: Bearer <access_token>`.

Protected routes additionally require:
- user `status = active`
- allowed `role_name` (`member` / `admin` depending on route)

## Proposals

### Create proposal
- `POST /proposals`
- Access: `member`, `admin`
- Purpose: create a movie proposal for moderation

Request body:
```json
{
  "title": "Interstellar",
  "description": "Sci-fi drama",
  "small_description": "Space mission",
  "duration": 169,
  "release_date": 2014,
  "country": "US",
  "poster": "https://example.com/poster.jpg",
  "rating_kp": 8.6,
  "source": "manual"
}
```

Rules:
- `source` must be one of: `manual`, `kinopoisk` (empty defaults to `manual`)
- status on create is always `pending`

Response:
```json
{
  "proposal": {
    "id": 42,
    "status": "pending",
    "source": "manual"
  }
}
```

### My proposals
- `GET /proposals/me`
- Access: `member`, `admin`
- Purpose: list proposals created by current user

Response:
```json
{
  "proposals": []
}
```

### Admin list proposals
- `GET /admin/proposals`
- `GET /admin/proposals?status=pending|approved|rejected`
- Access: `admin`
- Purpose: moderation queue and history

Response:
```json
{
  "proposals": []
}
```

### Admin moderate proposal
- `PATCH /admin/proposals/:id/status`
- Access: `admin`
- Purpose: approve or reject pending proposal

Request body:
```json
{
  "status": "approved",
  "moderation_comment": "looks good"
}
```

Rules:
- allowed status: `approved`, `rejected`
- `pending` is terminal input for create only
- already moderated proposals cannot be moderated again
- approve creates or links a film in one transaction

Response:
```json
{
  "proposal": {
    "id": 42,
    "status": "approved",
    "film_id": 55
  }
}
```

## Kinopoisk

### Search
- `GET /kinopoisk/search?q=<query>&limit=<1..20>`
- alias query param also accepted: `query`
- Access: `member`, `admin`
- Purpose: search in provider API without creating records

Response:
```json
{
  "results": [
    {
      "provider_movie_id": "301",
      "title": "The Matrix",
      "description": "Neo story",
      "small_description": "Neo",
      "duration": 136,
      "release_date": 1999,
      "country": "US",
      "poster": "https://example.com/poster.jpg",
      "rating_kp": 8.5,
      "genres": ["sci-fi"],
      "source": "kinopoisk"
    }
  ]
}
```

Notes:
- API key is used only on backend.
- If provider is unavailable, endpoint returns generic provider error (no internal detail leakage).
- If backend is not configured with provider key, endpoint returns configuration error.

## Statistics

### Summary
- `GET /statistics/summary`
- Access: public read-only
- Purpose: aggregate counters and averages calculated from raw tables

Response:
```json
{
  "summary": {
    "movies_total": 0,
    "franchises_total": 0,
    "users_active": 0,
    "proposals_pending": 0,
    "proposals_approved": 0,
    "proposals_rejected": 0,
    "expectations_total": 0,
    "expectations_numeric": 0,
    "expectations_refuse": 0,
    "expectations_average": 0,
    "reviews_total": 0,
    "reviews_average": 0,
    "weekly_packs_total": 0,
    "weekly_packs_voting": 0,
    "weekly_packs_closed": 0,
    "weekly_pack_votes_total": 0,
    "weekly_pack_movies_total": 0,
    "generated_at": "2026-05-29T12:00:00Z"
  }
}
```

## Movies (Admin Create Flow)

### Canonical create endpoint
- `POST /admin/movies`
- Access: `admin`
- Purpose: explicit admin-only movie creation flow

### Legacy alias
- `POST /film/`
- Access: `admin`
- Status: legacy alias kept for backward compatibility

## Weekly Packs (Mini App Helpers)

### Current voting pack
- `GET /weekly-packs/current`
- Access: public read-only
- Purpose: return current weekly pack in `voting` status

Response:
```json
{
  "weekly_pack": {
    "id": 17,
    "status": "voting"
  }
}
```

Rules:
- if no active pack exists, backend returns `404`.

### My vote limits for pack
- `GET /weekly-packs/:id/votes/me/limits`
- Access: `member`, `admin`
- Purpose: return remaining score limits for current user in one pack

Response:
```json
{
  "limits": {
    "pack_id": 17,
    "limits": [
      {"score": 3, "limit": 1, "used": 1, "remaining": 0},
      {"score": 2, "limit": 2, "used": 0, "remaining": 2},
      {"score": 1, "limit": 2, "used": 1, "remaining": 1},
      {"score": -2, "limit": 1, "used": 0, "remaining": 1}
    ],
    "zero_score_unlimited": true,
    "zero_score_used": 4
  }
}
```

## Error Contract

Current API uses:
```json
{
  "error": "human readable message"
}
```

Common status mapping in this stage:
- `400` invalid request data
- `401` auth required / invalid token
- `403` role or status forbidden
- `404` not found
- `409` logical conflict (where implemented)
- `500` internal server/config error
- `502` upstream provider failure

## Stage Scope

Covered in this stage:
- proposal lifecycle with moderation audit
- provider search integration
- raw-data summary statistics
- role-name based authorization for updated routes
- weekly pack mini app helpers (`current`, `me/limits`)

Deferred to later stages:
- telegram notification workflows
- full mini app UI integration
- extended analytics dimensions
- import pipeline hardening
