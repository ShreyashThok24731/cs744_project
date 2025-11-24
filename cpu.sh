#!/bin/bash

mkdir -p results

if ! command -v mpstat &> /dev/null; then
    echo "Error: mpstat not found."
    exit 1
fi

docker rm -f cs744-db 2>/dev/null
docker run --name cs744-db --cpuset-cpus="2" -e POSTGRES_PASSWORD=thok -p 5433:5432 -d postgres > /dev/null
sleep 5

fuser -k 8080/tcp > /dev/null 2>&1
echo "Starting Server on CPU 1..."
taskset -c 1 go run . -quiet &
SERVER_PID=$!
sleep 2

echo "Seeding data..."
for i in {0..100}; do
  curl -s -X POST -d "{\"key\":\"popular-$i\", \"value\":\"cached-data\"}" http://localhost:8080/kv/ > /dev/null
done

LOADS=(2 4 8 16 24 32 50 64)
DURATION=300
WORKLOAD="get_popular"
OUTPUT_FILE="results/${WORKLOAD}_results.csv"

echo "Clients,Throughput,ResponseTime(ms),Server_CPU_Active(%)" > $OUTPUT_FILE

echo "Running Workload: $WORKLOAD"

for CLIENTS in "${LOADS[@]}"; do
    echo "Processing $CLIENTS clients..."
    
    mpstat -P 1 1 $DURATION > cpu_stats.txt &
    MPSTAT_PID=$!
    
    RESULT=$(go run loadgen/load_generator.go \
        -clients $CLIENTS \
        -duration $DURATION \
        -workload $WORKLOAD)
    
    wait $MPSTAT_PID
    
    THROUGHPUT=$(echo "$RESULT" | grep "throughput" | cut -d',' -f2)
    RESPONSE_TIME=$(echo "$RESULT" | grep "latency" | cut -d',' -f2)
    
    C1_USR=$(grep "Average" cpu_stats.txt | awk '$2=="1" {print $3}')
    C1_SYS=$(grep "Average" cpu_stats.txt | awk '$2=="1" {print $5}')
    C1_SOFT=$(grep "Average" cpu_stats.txt | awk '$2=="1" {print $8}')
    SERVER_ACTIVE=$(awk "BEGIN {print $C1_USR + $C1_SYS + $C1_SOFT}")

    echo "Result: $THROUGHPUT req/s | Latency: $RESPONSE_TIME ms | Server CPU: $SERVER_ACTIVE%"
    
    echo "$CLIENTS,$THROUGHPUT,$RESPONSE_TIME,$SERVER_ACTIVE" >> $OUTPUT_FILE
    
    rm cpu_stats.txt
    sleep 5 
done

echo "Stopping..."
kill $SERVER_PID
