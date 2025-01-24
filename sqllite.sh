set -e

(
  cd "$(dirname "$0")"
  go build -o %TEMP%\sqlite-go *.go
)