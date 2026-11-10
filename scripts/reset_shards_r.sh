#!/bin/bash
nodes=${1:-8}
mode=${2:-0}
reorder=${3:-1}
verbose=${4:-""}
PREFIX="192.168.20."
godir="/usr/local/go/bin/go"
r="-r"
if [ $reorder -ne 1 ]; then
    r=""
fi

for (( nodeId=0; nodeId<nodes; nodeId++ )); do
  addr=$PREFIX$((nodeId + 2))
  ssh -o StrictHostKeyChecking=no root@${addr} "pkill -f startServer"
  ssh -o StrictHostKeyChecking=no root@${addr} "cd /root/ShardDB; \
          cp data/*.db server/db/;"

  for (( serverId=0; serverId<10; serverId++ )); do
    id=$((nodeId*10 + serverId))
    logFile="shardLog${id}.txt"
    ssh -o StrictHostKeyChecking=no root@${addr} "cd /root/ShardDB; \
          rm -f ${logFile}; touch ${logFile}; \
          ${godir} run startServer.go --config=node --id=${id} --mode=${mode} ${r} ${verbose} &>> ${logFile} &"
  done
done
