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

### Option: Non SQL databases

TBD

## Decision

We will use the PostgreSQL DBMS. 
It is both reliable by using replication and is able to store large amounts of data.
To play it safe, we will make our storage 120% of estimated bytes we need to store, 
i. e. about 1585 TiB. We will make use of 3 replicas, which will mean total 
4755 TiB of storage size. If the database is sharded, we will need 
about 159 pods, assuming each of them can store up to 10 TiB of data.

To conduct the experiment we can switch PostgreSQL between synchronous and asynchronous replication modes.

## Consequences

* Information about orders will be persisted in reliable way. 
* We will be able to store information about the order for up to 2 years.
* We can relatively easily conduct our experiment simply
by switching between synchronous/asynchronous replication modes. 