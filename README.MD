# Distributed IC3 Database

This is a prototype of a distributed IC3 protocol

## Installation
```bash
cd scripts
chmod +x *.sh
./install_dependencies.sh
```

## Compile and Run
1. Generate the data
```bash
cd data
go run genData_TPCC.go
```
2. Build Docker Image
```bash
cd docker
ssh-keygen
./build_docker.sh
```
3. Start docker container clusters
```bash
cd scripts
./setup_clusters.sh
```

## Usage
```bash
cd scripts
./exp.sh <dockers> <protocol> <concurrency>
```

All parameters are numeric:

- `dockers`: number of docker containers, by default, four containers are started with `setup_clusters.sh`.
- `protocol`: concurrency protocol, 1 = 2PC, 2 = IC3, 3 = DIC3.
- `concurrency`: number of concurrent requests.