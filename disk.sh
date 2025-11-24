#!/bin/bash

mkdir -p results

if ! command -v mpstat &> /dev/null; then
    echo "Error: mpstat not found."
    exit 1
fi

DB_CORE=2
SERVER_CORES="0,1,3,4,5,6,7,8,9,10,11,12,13,14,15"

docker rm -f cs744-db 2>/dev/null
docker run --name cs744-db --cpuset-cpus="2" -e POSTGRES_PASSWORD=thok -e POSTGRES_LOGGING_LEVEL="FATAL" -p 5433:5432 -d postgres > /dev/null
sleep 5 

fuser -k 8080/tcp > /dev/null 2>&1
echo "Starting Server on CPU 1..."
taskset -c $SERVER_CORES go run . -quiet &
SERVER_PID=$!
sleep 2

echo "Seeding data..."
for i in {0..100}; do
  curl -s -X POST -d "{\"key\":\"popular-$i\", \"value\":\"cached-data\"}" http://localhost:8080/kv/ > /dev/null
done

LOADS=(1 3 5 7 9 11 13 15)
DURATION=180
WORKLOAD="put_all"
OUTPUT_FILE="results/${WORKLOAD}_results.csv"

echo "Clients,Throughput,ResponseTime(ms),DB_IO_Wait(%)" > $OUTPUT_FILE

echo "Running Workload: $WORKLOAD"

for CLIENTS in "${LOADS[@]}"; do
    echo "Processing $CLIENTS clients..."
    
    mpstat -P ALL 1 $DURATION > cpu_stats.txt &
    MPSTAT_PID=$!
    
    RESULT=$(go run loadgen/load_generator.go \
        -clients $CLIENTS \
        -duration $DURATION \
        -workload $WORKLOAD)
    
    wait $MPSTAT_PID
    
    THROUGHPUT=$(echo "$RESULT" | grep "throughput" | cut -d',' -f2)
    RESPONSE_TIME=$(echo "$RESULT" | grep "latency" | cut -d',' -f2)
    
    DB_WAIT=$(grep "Average" cpu_stats.txt | awk '$2=="2" {print $6}')

    echo "Result: $THROUGHPUT req/s | Latency: $RESPONSE_TIME ms | DB Wait: $DB_WAIT%"
    
    echo "$CLIENTS,$THROUGHPUT,$RESPONSE_TIME,$DB_WAIT" >> $OUTPUT_FILE
    
    rm cpu_stats.txt
    sleep 5 
done

echo "Stopping..."
kill $SERVER_PID
