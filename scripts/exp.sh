#!/bin/bash

# Number of Shards
nodes=${1:-4}
# Concurrency control Mode
# 1=2PC, 2=IC3, 3=DIC3, 4=IC3+Ordered2PL, 5 = IC3+MessagePulling
mode=${2:-0}
# Number of concurrent requests
con=${3:-10}
# Number of warehouses, cannot be larger than number of nodes
numW=${4:-4}
filePrefix=${5:-""}
verbose="-v"
PREFIX="192.168.20."
ClientAddr=${PREFIX}$((nodes+2))
godir="/usr/local/go/bin/go"

./reset_shards.sh ${nodes} ${mode} ${verbose}

sleep 1
logFile="log/${filePrefix}result.txt"
go run ../runTPCC.go --config=../node --mode=${mode} --id=0 --rate=${con} --w=${numW} ${verbose}