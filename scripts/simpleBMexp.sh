#!/bin/bash

# Number of Shards
nodes=${1:-4}
# Concurrency control Mode
# 1=2PC, 2=IC3, 3=DIC3, 4=IC3+Ordered2PL, 5 = IC3+MessagePulling
mode=${2:-0}
# Number of concurrent requests
con=${3:-10}
reorder=${4:-1}
filePrefix=${5:-""}
verbose=""
PREFIX="192.168.20."
ClientAddr=${PREFIX}$((nodes+2))
godir="/usr/local/go/bin/go"

./reset_shards_r.sh ${nodes} ${mode} ${reorder} ${verbose}

sleep 1
logFile="log/${filePrefix}result.txt"
go run ../runSimpleBM.go --config=../node --mode=${mode} --id=0 --rate=${con} ${verbose}