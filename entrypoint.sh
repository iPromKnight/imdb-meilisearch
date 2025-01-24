#!/bin/sh

cleanup() {
    echo "Stopping Meilisearch and Go app..."
    kill "$MEILI_PID" "$GO_PID"
    wait "$MEILI_PID" "$GO_PID"
    exit 0
}

export MEILI_NO_ANALYTICS=true
export MEILISEARCH_HOST=${MEILISEARCH_HOST:-"http://127.0.0.1:7700"}
export LOG_DEBUG=${LOG_DEBUG:-false}
export SEED_ON_STARTUP=${SEED_ON_STARTUP:-false}

if [ "$LOG_DEBUG" = "true" ]; then
    export DEBUG_PARAM="-d"
fi

if [ "$MEILISEARCH_HOST" = "http://127.0.0.1:7700" ]; then
    echo "Using internal Meilisearch instance"
    /usr/local/bin/meilisearch --http-addr "0.0.0.0:7700" &
    MEILI_PID=$!
fi

echo "Waiting for Meilisearch to start..."
until curl -s "$MEILISEARCH_HOST/health" | grep -q '"status":"available"'; do
    echo "Meilisearch not ready yet, retrying in 1 second..."
    sleep 1
done
echo "Meilisearch is up and running"

if [ "$SEED_ON_STARTUP" = "true" ]; then
    echo "Seeding data..."
    /usr/local/bin/imdb-meilisearch seed
    echo "Data seeded"
fi

/usr/local/bin/imdb-meilisearch daemon $DEBUG_PARAM &
GO_PID=$!

trap cleanup INT TERM

if [ "$MEILISEARCH_HOST" = "http://127.0.0.1:7700" ]; then
  wait "$MEILI_PID" "$GO_PID"
else
  wait "$GO_PID"
fi

echo "One of the processes exited. Stopping the container..."
cleanup
