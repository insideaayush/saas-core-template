#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${BASH_VERSION:-}" ]]; then
  exec bash "$0" "$@"
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

NO_INFRA=0
DOWN_AFTER=0
SKIP_UI=0
SKIP_WORKER=0
SKIP_MIGRATIONS=0

usage() {
  cat <<'EOF'
Local end-to-end smoke test (infra + backend + worker + frontend).

Usage:
  bash scripts/smoke-local.sh [options]

Options:
  --no-infra          Do not run docker compose up
  --down              Run docker compose down after the test
  --skip-ui           Do not start Next.js dev server
  --skip-worker       Do not start worker or test jobs
  --skip-migrations   Do not apply SQL migrations
EOF
}

for arg in "$@"; do
  case "$arg" in
    --no-infra) NO_INFRA=1 ;;
    --down) DOWN_AFTER=1 ;;
    --skip-ui) SKIP_UI=1 ;;
    --skip-worker) SKIP_WORKER=1 ;;
    --skip-migrations) SKIP_MIGRATIONS=1 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown argument: $arg" >&2; usage; exit 2 ;;
  esac
done

API_PORT="${PORT:-8080}"
API_BASE=""
UI_BASE="http://localhost:3000"

SMOKE_DB_NAME="${SMOKE_DB_NAME:-saas_core_template_smoke}"
if [[ ! "${SMOKE_DB_NAME}" =~ ^[a-zA-Z0-9_]+$ ]]; then
  echo "invalid SMOKE_DB_NAME (expected [a-zA-Z0-9_]+): ${SMOKE_DB_NAME}" >&2
  exit 2
fi

port_is_open() {
  local port="$1"
  (echo >/dev/tcp/127.0.0.1/"${port}") >/dev/null 2>&1
}

pick_free_port() {
  local base_port="$1"
  local max_tries="${2:-25}"
  local port="${base_port}"
  local i=0

  while (( i < max_tries )); do
    if port_is_open "${port}"; then
      port=$((port + 1))
      i=$((i + 1))
      continue
    fi
    echo "${port}"
    return 0
  done

  return 1
}

API_PORT="$(pick_free_port "${API_PORT}" 25 || true)"
if [[ -z "${API_PORT}" ]]; then
  echo "could not find a free port starting from ${PORT:-8080}" >&2
  exit 1
fi
API_BASE="http://localhost:${API_PORT}"

DATABASE_URL_DEFAULT="postgres://postgres:postgres@localhost:5432/${SMOKE_DB_NAME}?sslmode=disable"
REDIS_URL_DEFAULT="redis://localhost:6379/0"

export DATABASE_URL="${DATABASE_URL:-$DATABASE_URL_DEFAULT}"
export REDIS_URL="${REDIS_URL:-$REDIS_URL_DEFAULT}"
export APP_BASE_URL="${APP_BASE_URL:-$UI_BASE}"
export APP_ENV="${APP_ENV:-development}"
export APP_VERSION="${APP_VERSION:-smoke}"

export OTEL_TRACES_EXPORTER="${OTEL_TRACES_EXPORTER:-none}"
export ANALYTICS_PROVIDER="${ANALYTICS_PROVIDER:-console}"
export ERROR_REPORTING_PROVIDER="${ERROR_REPORTING_PROVIDER:-console}"
export EMAIL_PROVIDER="${EMAIL_PROVIDER:-console}"

export FILE_STORAGE_PROVIDER="${FILE_STORAGE_PROVIDER:-disk}"
export FILE_STORAGE_DISK_PATH="${FILE_STORAGE_DISK_PATH:-./.data/uploads}"

export JOBS_ENABLED="${JOBS_ENABLED:-true}"
export JOBS_WORKER_ID="${JOBS_WORKER_ID:-smoke}"
export JOBS_POLL_INTERVAL="${JOBS_POLL_INTERVAL:-1s}"

API_PID=""
WORKER_PID=""
UI_PID=""

SCRIPT_PGID="$(ps -o pgid= $$ 2>/dev/null | tr -d '[:space:]' || true)"

start_bg() {
  if command -v setsid >/dev/null 2>&1; then
    setsid "$@" &
  else
    "$@" &
  fi
}

kill_process_group() {
  local pid="$1"
  if [[ -z "${pid}" ]]; then
    return 0
  fi
  if ! kill -0 "${pid}" 2>/dev/null; then
    return 0
  fi

  local pgid
  pgid="$(ps -o pgid= "${pid}" 2>/dev/null | tr -d '[:space:]' || true)"
  if [[ -n "${pgid}" && -n "${SCRIPT_PGID}" && "${pgid}" == "${SCRIPT_PGID}" ]]; then
    kill -TERM "${pid}" 2>/dev/null || true
  elif [[ -n "${pgid}" ]]; then
    kill -TERM -- "-${pgid}" 2>/dev/null || true
  else
    kill -TERM "${pid}" 2>/dev/null || true
  fi
}

wait_gone() {
  local pid="$1"
  local timeout_seconds="${2:-5}"
  local start
  start="$(date +%s)"

  while kill -0 "${pid}" 2>/dev/null; do
    local now
    now="$(date +%s)"
    if (( now - start > timeout_seconds )); then
      return 1
    fi
    sleep 1
  done
  return 0
}

cleanup() {
  set +e
  kill_process_group "${UI_PID}"
  kill_process_group "${WORKER_PID}"
  kill_process_group "${API_PID}"

  for pid in "${UI_PID}" "${WORKER_PID}" "${API_PID}"; do
    if [[ -z "${pid}" ]]; then
      continue
    fi
    if wait_gone "${pid}" 5; then
      wait "${pid}" 2>/dev/null || true
      continue
    fi

    kill -KILL "${pid}" 2>/dev/null || true
    wait "${pid}" 2>/dev/null || true
  done

  if [[ "${DOWN_AFTER}" == "1" ]]; then
    docker compose down >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

wait_for() {
  local name="$1"
  local cmd="$2"
  local timeout_seconds="${3:-60}"
  local start
  start="$(date +%s)"

  while true; do
    if eval "${cmd}" >/dev/null 2>&1; then
      echo "ok: ${name}"
      return 0
    fi

    local now
    now="$(date +%s)"
    if (( now - start > timeout_seconds )); then
      echo "timeout waiting for ${name}" >&2
      return 1
    fi
    sleep 1
  done
}

wait_http_ok() {
  local name="$1"
  local url="$2"
  local timeout_seconds="${3:-60}"
  wait_for "${name}" "curl -fsS \"${url}\"" "${timeout_seconds}"
}

require_bin() {
  local bin="$1"
  if ! command -v "${bin}" >/dev/null 2>&1; then
    echo "missing required binary: ${bin}" >&2
    exit 1
  fi
}

require_bin curl
require_bin docker

if [[ "${SKIP_UI}" == "0" ]]; then
  if command -v node >/dev/null 2>&1; then
    NODE_VERSION="$(node -v | tr -d 'v' || true)"
    if [[ "${NODE_VERSION}" == 19.* ]]; then
      NODE_MINOR="$(echo "${NODE_VERSION}" | cut -d. -f2)"
      if [[ "${NODE_MINOR}" -lt 8 ]]; then
        echo "Node.js ${NODE_VERSION} is too old for Next.js; use Node 20+ (or run with --skip-ui)." >&2
        exit 1
      fi
    fi
  fi
fi

if [[ "${NO_INFRA}" == "0" ]]; then
  echo "==> starting infra (docker compose)"
  docker compose up -d postgres redis otel-collector

  wait_for "postgres" "docker compose exec -T postgres pg_isready -U postgres -d postgres" 90
  wait_for "postgres sql" "docker compose exec -T postgres psql -qtA -U postgres -d postgres -c 'SELECT 1' | tr -d '[:space:]' | grep -q '^1$'" 90
  wait_for "redis" "docker compose exec -T redis redis-cli ping | grep -q PONG" 90
fi

echo "==> preparing smoke database (${SMOKE_DB_NAME})"
wait_for "postgres ready for ddl" "docker compose exec -T postgres psql -qtA -U postgres -d postgres -c 'SELECT 1' | tr -d '[:space:]' | grep -q '^1$'" 60
docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U postgres -d postgres >/dev/null <<SQL
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = '${SMOKE_DB_NAME}'
  AND pid <> pg_backend_pid();
DROP DATABASE IF EXISTS ${SMOKE_DB_NAME};
CREATE DATABASE ${SMOKE_DB_NAME};
SQL

if [[ "${SKIP_MIGRATIONS}" == "0" ]]; then
  echo "==> applying migrations"
  (
    cd backend
    DATABASE_URL="${DATABASE_URL}" go run ./cmd/migrate up -dir ./migrations
  )
fi

echo "==> starting api"
start_bg bash -c "cd backend && PORT='${API_PORT}' exec go run ./cmd/api"
API_PID="$!"

wait_http_ok "api /healthz" "${API_BASE}/healthz" 60
wait_http_ok "api /readyz" "${API_BASE}/readyz" 60
wait_http_ok "api /api/v1/meta" "${API_BASE}/api/v1/meta" 60

if [[ "${SKIP_WORKER}" == "0" ]]; then
  echo "==> starting worker"
  start_bg bash -c "cd backend && exec go run ./cmd/worker"
  WORKER_PID="$!"

  echo "==> testing jobs (enqueue -> worker processes -> done)"
  JOB_ID="$(
    docker compose exec -T postgres psql -qtA -U postgres -d "${SMOKE_DB_NAME}" -v ON_ERROR_STOP=1 -c \
      "INSERT INTO jobs (type, payload, status, run_at) VALUES ('send_email', '{\"to\":\"smoke@example.com\",\"subject\":\"Smoke test\",\"text\":\"Hello from smoke test.\"}'::jsonb, 'queued', now()) RETURNING id::text;"
  )"
  JOB_ID="$(echo "${JOB_ID}" | tr -d '[:space:]')"
  if [[ -z "${JOB_ID}" ]]; then
    echo "failed to enqueue job" >&2
    exit 1
  fi

  wait_for "job ${JOB_ID} done" "docker compose exec -T postgres psql -qtA -U postgres -d \"${SMOKE_DB_NAME}\" -c \"SELECT status FROM jobs WHERE id = '${JOB_ID}'\" | tr -d '[:space:]' | grep -q '^done$'" 30
fi

if [[ "${SKIP_UI}" == "0" ]]; then
  echo "==> starting ui"
  (
    cd frontend
    NEXT_PUBLIC_API_URL="${API_BASE}" npm run dev
  ) &
  UI_PID="$!"

  wait_http_ok "ui /" "${UI_BASE}/" 90
  wait_http_ok "ui /pricing" "${UI_BASE}/pricing" 90
fi

echo "==> smoke test passed"
echo "API: ${API_BASE}"
if [[ "${SKIP_UI}" == "0" ]]; then
  echo "UI:  ${UI_BASE}"
fi
