#!/bin/bash

# Number of Shards
nodes=${1:-2}
# Concurrency control Mode
# 1=2PC, 2=IC3, 3=DIC3, 4=IC3+Ordered2PL, 5 = IC3+MessagePulling
mode=${2:-1}
logFileName=${3:-"log"}
rm -f "${logFileName}.txt"
touch "${logFileName}.txt"

concurrency="10 20 40 60 80 100"
for SH in $concurrency; do
  echo "Running on ${SH} concurrency"
  ./exp.sh "${nodes}" "${mode}" "${SH}" ${nodes} | tee -a "${logFileName}.txt"
  sleep 1
done