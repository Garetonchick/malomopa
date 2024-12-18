import uuid
import random
from datetime import datetime, timedelta
import numpy as np
import argparse
import json

def generate_data(num_entries, start_time, end_time, cost_distribution, bool_combination_dist):
    """
    Generator for ScyllaDB data.

    Parameters:
    - num_entries (int): Number of entries to generate.
    - start_time (datetime): Start time for 'created_at'.
    - end_time (datetime): End time for 'created_at'.
    - cost_distribution (dict): Parameters for cost distribution, e.g., {"type": "normal", "mean": 50, "stddev": 10}.
    - bool_combination_dist (dict): Distribution percentages for (is_acquired, is_cancelled).
    """
    if sum(bool_combination_dist.values()) != 100:
        raise ValueError("Boolean combination percentages must sum up to 100.")

    # Create a list of weighted boolean combinations
    bool_combinations = []
    for bool_pair, percentage in bool_combination_dist.items():
        bool_combinations.extend([bool_pair] * percentage)

    # Generate random timestamps in the given range
    time_range = (end_time - start_time).total_seconds()

    def generate_cost():
        """Generate cost based on distribution."""
        if cost_distribution["type"] == "normal":
            return max(0, np.random.normal(cost_distribution["mean"], cost_distribution["stddev"]))
        elif cost_distribution["type"] == "uniform":
            return random.uniform(cost_distribution["low"], cost_distribution["high"])
        elif cost_distribution["type"] == "exponential":
            return np.random.exponential(cost_distribution["scale"])
        else:
            raise ValueError("Unsupported cost distribution type.")

    for _ in range(num_entries):
        order_id = str(uuid.uuid4())
        executor_id = str(uuid.uuid4())
        created_at = start_time + timedelta(seconds=random.uniform(0, time_range))
        cost = round(generate_cost(), 2)
        is_acquired, is_cancelled = random.choice(bool_combinations)
        payload = bytes(random.getrandbits(8) for _ in range(10))

        yield {
            "order_id": order_id,
            "executor_id": executor_id,
            "created_at": created_at.isoformat(),
            "cost": cost,
            "payload": payload.hex(),
            "is_acquired": is_acquired,
            "is_cancelled": is_cancelled
        }

def parse_args():
    """Parse command-line arguments."""
    parser = argparse.ArgumentParser(description="Data generator for ScyllaDB based on input parameters.")
    parser.add_argument("--num_entries", type=int, required=True, help="Number of entries to generate.")
    parser.add_argument("--start_time", type=str, required=True, help="Start time for 'created_at' in format YYYY-MM-DDTHH:MM:SS.")
    parser.add_argument("--end_time", type=str, required=True, help="End time for 'created_at' in format YYYY-MM-DDTHH:MM:SS.")
    parser.add_argument("--cost_distribution", type=str, required=True,
                        help="JSON string for cost distribution, e.g., '{\"type\": \"normal\", \"mean\": 50, \"stddev\": 10}'.")
    parser.add_argument("--bool_distribution", type=str, required=True,
                        help="JSON string for boolean distribution, e.g., '{\"(True, True)\": 25, \"(True, False)\": 25, \"(False, True)\": 25, \"(False, False)\": 25}'.")
    parser.add_argument("--output_file", type=str, default="output.json", help="Output file to save generated data (default: output.json).")
    return parser.parse_args()

def main():
    args = parse_args()

    # Parse start and end times
    start_time = datetime.fromisoformat(args.start_time)
    end_time = datetime.fromisoformat(args.end_time)

    # Parse JSON inputs
    cost_distribution = json.loads(args.cost_distribution)
    bool_combination_dist = {eval(k): v for k, v in json.loads(args.bool_distribution).items()}

    # Generate data and save to file
    generator = generate_data(args.num_entries, start_time, end_time, cost_distribution, bool_combination_dist)

    with open(args.output_file, "w") as file:
        for entry in generator:
            file.write(json.dumps(entry) + "\n")

    print(f"Generated {args.num_entries} entries and saved to {args.output_file}")

if __name__ == "__main__":
    main()
