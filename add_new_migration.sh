if [ -z "$1" ]; then
  echo "Usage: $0 <migration_name>"
  exit 1
fi

goose -dir db/migrations create "$1" sql
