FROM redpandadata/redpanda:latest

CMD ["redpanda", "start", \
     "--kafka-addr", "PLAINTEXT://0.0.0.0:9092", \
     "--advertise-kafka-addr", "PLAINTEXT://buy-now-kafka.internal:9092", \
     "--pandaproxy-addr", "PLAINTEXT://0.0.0.0:8082", \
     "--advertise-pandaproxy-addr", "PLAINTEXT://buy-now-kafka.internal:8082", \
     "--smp", "1", \
     "--memory", "400M", \
     "--overprovisioned", \
     "--node-id", "0", \
     "--check=false"]
