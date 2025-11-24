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

echo "Seeding popular keys..."
for i in {0..100}; do
  curl -s -X POST -d "{\"key\":\"popular-$i\", \"value\":\"cached-data\"}" http://localhost:8080/kv/ > /dev/null
done

LOADS=(2 4 8 16 32 64 96 128 160 192)
DURATION=10

run_workload() {
    WORKLOAD=$1
    OUTPUT_FILE="results/${WORKLOAD}_results.csv"
    
    echo "Clients,Throughput,ResponseTime(ms),Metric_Value(%)" > $OUTPUT_FILE
    
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
        
        if [ "$WORKLOAD" == "get_popular" ]; then
            USR=$(grep "Average" cpu_stats.txt | awk '$2=="1" {print $3}')
            SYS=$(grep "Average" cpu_stats.txt | awk '$2=="1" {print $5}')
            SFT=$(grep "Average" cpu_stats.txt | awk '$2=="1" {print $8}')
            METRIC=$(awk "BEGIN {print $USR + $SYS + $SFT}")
            echo "Result: $THROUGHPUT req/s | Latency: $RESPONSE_TIME ms | Target: Server Active CPU ($METRIC%)"
            
        else
            METRIC=$(grep "Average" cpu_stats.txt | awk '$2=="2" {print $6}')
            echo "Result: $THROUGHPUT req/s | Latency: $RESPONSE_TIME ms | Target: Disk I/O Wait ($METRIC%)"
        fi

        echo "$CLIENTS,$THROUGHPUT,$RESPONSE_TIME,$METRIC" >> $OUTPUT_FILE
        
        rm cpu_stats.txt
        sleep 5 
    done
}

run_workload "get_popular"
run_workload "put_all"

echo "Stopping..."
kill $SERVER_PID
