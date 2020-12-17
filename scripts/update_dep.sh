#!/usr/bin/env bash
set -Eeuo pipefail

updates=($(go list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}{{end}}' -m all))

for update in "${updates[@]}"; do
  echo "run: go get ${update}"
  go get "${update}"
done
