# Architecture Decision Record: Cache Service

## Status

Proposed

## Context

Data sources for this project are divided in the two categories &mdash; critical
and noncritical. To satisfy our reliablity constraints,
noncritical data sources can be cached. But there are complications:

* We expect the number of noncritical data sources to grow significantly in the future
* Different data sources can have different cache strategies
* Implementing fetching and caching of data from data sources
is not trivial and requires quite a bit of work to make it parallel and effective

## Decision

Put all of the logic for aggregating and caching of different datasources
into a distinct module, which we will call *Cache Service*.

This service will have only one purpose &mdash; to provide *Order Service*
with information from all the datasources.
As parameters it will take in IDs of the order and the executor and return
info from all of the datasources needed for *assign_order*
endpoint (see [ADR-002](adr-002-order-service.md)).

Internally this service must implement effective paralel aggregation of data and
caching.

Should both fetch and cache retrieval for one of
the data sources fail, it must return
immidiately with an error.

Some noncritical data sources can
have policy, where cache older than duration X can't be used. In this case
an error should be returned if we can't fetch this data source and
cache is stale.

## Consequences

* All of the aggregation and caching of data sources logic can be
developed independently from the main service as long as interface stays
the same, which should improve maintability.

* Use of caching in the service improves reliability. Even
if one of data source's services is down, we still can have the chance
to get the data from cache.

* Efficient parallel aggregation improves both latency and throughput.
