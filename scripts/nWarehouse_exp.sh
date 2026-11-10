#!/bin/bash

# Number of Shards
nodes=${1:-2}
logFileName=${2:-"log"}
concurrency=80
rm -f "${logFileName}.txt"
touch "${logFileName}.txt"

mode="1 2 3"
for SH in $mode; do
  echo "Running on mode ${SH}"
  ./exp.sh "${nodes}" "${SH}" "${concurrency}" ${nodes} | tee -a "${logFileName}.txt"
  sleep 1
done