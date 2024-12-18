python3 main.py \
    --num_entries 1000 \
    --start_time "2024-01-01T00:00:00" \
    --end_time "2024-01-02T00:00:00" \
    --cost_distribution '{"type": "normal", "mean": 50, "stddev": 10}' \
    --bool_distribution '{"(True, True)": 25, "(True, False)": 25, "(False, True)": 25, "(False, False)": 25}' \
    --output_file output.json
