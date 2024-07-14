#!/bin/bash

# Run lint on all jsonnet files in the repository
RESULT=0;
for f in $(find . -name 'vendor' -prune -o -name '*.libsonnet' -print -o -name '*.jsonnet' -print); do
  # jsonnet fmt -i "$$f"
  echo "Linting ${f}"
  jsonnetfmt -- "${f}" | diff -u "${f}" -
  RESULT=$((RESULT+$?))
done

echo "Linting complete"
exit $RESULT
