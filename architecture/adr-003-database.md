# Architecture Decision Record: Database

## Status

Proposed

## Context

As per [ADR-002](adr-002-order-service.md) we need persistent storage
for storing information about orders and their assigned executors.

* This storage must be reliable. 
* It must be able to store data for 2 years.
* We expect every order to be about 75 KiB in size.
Taking in account RPS for the *assign order* endpoint, we 
estimate to store about 1321 TiB.

Moreover we have additional requirement to conduct an experiment. 
Specifically, we need to research how much perfomance of the Order Service 
can be improved by sacrificing consistency.

### Option: Traditional relational SQL databases

These are databases that were written in the 80-90s, they are not distributed (i.e. one instance runs on one node), as a rule, they support strict consistency guarantees (ACID), and have a rich query language (SQL). Some of the most popular databases of this family are PostgreSQL, MySQL, Oracle. I considered PostgreSQL specifically, since I have at least some expertise in this database, plus good documentation and the availability of ready-made libraries for working with the database on golang.

### Option: NoSQL databases

These databases were invented in the 00-10s, are distributed, and easily scale horizontally (unlike the DBs discussed above). In terms of the CAP theorem, these DBs choose AP, and, as a rule, with default settings, these databases have consistency in the end. However, they have a rather truncated query language, and do not have transactions (they exist only at the level of one row). Examples of such databases: Cassandra, ScyllaDB, DynamoDB.

### Option: NewSQL databases

A rapidly growing area, these databases both have ACID guarantees and are distributed by design. They support full SQL. Examples: YDB, CockroachDB, YTsaurus.

## Decision

I chose the NoSQL database ScyllaDB. Now I will explain why I did not choose the other two options.

Why I did not choose traditional old relational DBs:
1) They are not distributed by design, and the storage volume required in the problem statement is quite large and will not fit on one node.
2) In the data and query scheme we developed, all handlers must change or insert rows into the database, that is, these are modifying operations. Therefore, the scheme with master-slave replication in this case is meaningless. If we make a sharded Postgres, and each shard will consist of one node and will be the master, then it turns out that we do not use the assumption that the consistency is not strict.
3) In the data and query scheme we developed, all sorts of transactions and other bells and whistles that SQL, rich in them, provides are not needed.

Why I didn't choose NewSQL databases:
1) We need strong consistency, it's not free anyway
2) No expertise
3) No need for distributed transactions

Why I chose NoSQL:
1) Eventual consistency by default, but you can configure it so that there will be strong consistency
2) It's easy to scale horizontally, and we need it
3) We don't need transactions, we only need transactionality at the level of one separate row of the database.

Why I chose ScyllaDB:
1) There is a ready-made golang gocql driver
2) Good documentation of CQL (Cassandra Query Language)
3) ScyllaDB is written in C++, Cassandra in Java, so in terms of performance the choice is obvious

To perform the experiment, we can tweak the parameters of the quorums for reading and writing. In each session with ScyllaDB we can configure the consistency level as it suits us.

## Consequences

* Information about orders will be persisted in reliable way. 
* We will be able to store information about the order for up to 2 years.
* We can relatively easily conduct our experiment simply
by editing quorum parameters. 