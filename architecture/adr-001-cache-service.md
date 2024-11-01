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

Put all of the logic for aggregating and caching of different noncritical datasources 
in a distinct microservice, which we will call *Cache Service*.

This service will have only one endpoint for now, let's call it *get_full_order_info*.
The *get_full_order_info* endpoint will take in an ID of the order and return 
JSON containing info from all of the noncritical datasources.

Internally this service must implement effective paralel aggregation of data and 
caching.

Should some of fetches for data sources fail, it must return all of the info it could 
get and fill in the missing parts with cache. Some noncritical data sources can 
have policy, where cache older than X seconds can't be used. In this case 
it should not fill in the data, but provide missing indicator in the response instead.

## Consequences

* All of the aggregation and caching of data sources logic can be 
developed independently from the main service as long as JSON schema 
stays compatible, which should improve maintability.

* Use of caching in the service improves reliability. Even 
if some data source service is down, we still can have the chance 
to get the data from cache.

* Efficient parallel aggregation improves both latency and throughput.