import uuid
import random
from datetime import datetime, timedelta
import numpy as np
import argparse
import json
from cassandra.cluster import Cluster
from cassandra.auth import PlainTextAuthProvider

def connect_to_scylla(hosts, keyspace, username, password):
    """Connect to the ScyllaDB cluster and return a session."""
    auth_provider = PlainTextAuthProvider(username, password)
    print(hosts)
    cluster = Cluster(hosts, auth_provider=auth_provider)
    session = cluster.connect()
    session.set_keyspace(keyspace)
    print(f"Connected to ScyllaDB keyspace: {keyspace}")
    return session


def create_table(session, table_name):
    """Create a table if it doesn't already exist."""
    query = f"""
    CREATE TABLE IF NOT EXISTS {table_name} (
        order_id TEXT PRIMARY KEY,
        executor_id TEXT,
        created_at TIMESTAMP,
        acquired_at TIMESTAMP,
        cost FLOAT,
        payload BLOB,
        is_acquired BOOLEAN,
        is_cancelled BOOLEAN
    );
    """
    session.execute(query)
    print(f"Ensured table '{table_name}' exists.")

def generate_data(num_entries, start_time, end_time, cost_distribution, bool_combination_dist):
    """Data generator as per configuration."""
    if sum(bool_combination_dist.values()) != 100:
        raise ValueError("Boolean combination percentages must sum up to 100.")

    # Create weighted boolean combinations
    bool_combinations = []
    for bool_pair, percentage in bool_combination_dist.items():
        bool_combinations.extend([bool_pair] * percentage)

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
        acquired_at = start_time + timedelta(seconds=random.uniform(0, time_range))
        if created_at > acquired_at:
            acquired_at, created_at = created_at, acquired_at
        cost = round(generate_cost(), 2)
        is_acquired, is_cancelled = random.choice(bool_combinations)
        payload = bytes(random.getrandbits(8) for _ in range(10))

        yield (order_id, executor_id, created_at, acquired_at, cost, payload, is_acquired, is_cancelled)

def insert_data(session, table_name, data_generator):
    """Insert generated data into the specified ScyllaDB table."""
    query = f"""
    INSERT INTO {table_name} (order_id, executor_id, created_at, acquired_at, cost, payload, is_acquired, is_cancelled)
    VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
    """

    count = 0
    for row in data_generator:
        session.execute(query, row)
        count += 1
        if count % 100 == 0:
            print(f"Inserted {count} rows...")

    print(f"Successfully inserted {count} rows into '{table_name}'.")

def parse_args():
    """Parse command-line arguments."""
    parser = argparse.ArgumentParser(description="Data generator for ScyllaDB and uploader.")
    parser.add_argument("--hosts", type=str, required=True, help="Comma-separated list of ScyllaDB hosts (IP addresses or hostnames).")
    parser.add_argument("--keyspace", type=str, required=True, help="Keyspace name in ScyllaDB.")
    parser.add_argument("--table", type=str, default="orders", help="Table name (default: 'orders').")
    parser.add_argument("--username", type=str, required=True, help="ScyllaDB username.")
    parser.add_argument("--password", type=str, required=True, help="ScyllaDB password.")
    parser.add_argument("--num_entries", type=int, required=True, help="Number of entries to generate.")
    parser.add_argument("--start_time", type=str, required=True, help="Start time for 'created_at' in format YYYY-MM-DDTHH:MM:SS.")
    parser.add_argument("--end_time", type=str, required=True, help="End time for 'created_at' in format YYYY-MM-DDTHH:MM:SS.")
    parser.add_argument("--cost_distribution", type=str, required=True,
                        help="JSON string for cost distribution, e.g., '{\"type\": \"normal\", \"mean\": 50, \"stddev\": 10}'.")
    parser.add_argument("--bool_distribution", type=str, required=True,
                        help="JSON string for boolean distribution, e.g., '{\"(True, True)\": 25, \"(True, False)\": 25, \"(False, True)\": 25, \"(False, False)\": 25}'.")
    return parser.parse_args()

def main():
    args = parse_args()

    # Parse input arguments
    hosts = args.hosts.split(",")
    keyspace = args.keyspace
    table_name = args.table
    username = args.username
    password = args.password
    start_time = datetime.fromisoformat(args.start_time)
    end_time = datetime.fromisoformat(args.end_time)
    cost_distribution = json.loads(args.cost_distribution)
    bool_combination_dist = {eval(k): v for k, v in json.loads(args.bool_distribution).items()}

    # Connect to ScyllaDB
    session = connect_to_scylla(hosts, keyspace, username, password)

    # Ensure table exists
    create_table(session, table_name)

    # Generate data
    print("Generating data...")
    data_generator = generate_data(args.num_entries, start_time, end_time, cost_distribution, bool_combination_dist)

    # Insert data into ScyllaDB
    print("Inserting data into ScyllaDB...")
    insert_data(session, table_name, data_generator)

    print("Data generation and upload complete.")

if __name__ == "__main__":
    main()
