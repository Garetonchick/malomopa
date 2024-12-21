#! /bin/bash

python3 main.py \
    --hosts "scylla-node1,scylla-node2,scylla-node3" \
    --keyspace "data" \
    --table "orders" \
    --username "cassandra" \
    --password "cassandra" \
    --num_entries 1000 \
    --start_time "2024-01-01T00:00:00" \
    --end_time "2024-01-02T00:00:00" \
    --cost_distribution '{"type": "normal", "mean": 50, "stddev": 10}' \
    --bool_distribution '{"(True, True)": 0, "(True, False)": 33, "(False, True)": 33, "(False, False)": 34}'
