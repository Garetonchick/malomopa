# Architecture Decision Record: Order Service

## Status

Proposed

## Context

When the decision is made to assign some executor to the order, we need to 
somehow make the executor know that he is now assigned to this order. But 
there some things that can go wrong:

* If we simply send to the executor message with information about the order it might be lost
* It can be decided that this order must be canceled
* Order can be updated with new info or reassigned to some other executor

## Decision

Our solution is the `Order Service`. This microservice must serve 4 
functions, which we will go over in the following subsections.

### Assign order

The service must implement an endpoint for assigning an order to 
the executor. It takes in ID of an order and ID of an executor 
and should work like this:

1. Fetches information about the order and the executor using Cache Service (see [ADR-001](adr-001-cache-service.md))
2. Computes cost of the order using fetched info
3. Persists information about the order and it's assigned executor in 
the database (see [ADR-003](adr-003-database.md))
4. Returns success/failure indicator

If data for at least one of the critical data sources is not available, it should
return failure indicator.

We account for 300 RPS load for this endpoint.

### Update order

The service must implement an endpoint for reassining an 
order to some other executor. As in the *assign order* this 
update should be persisted in the database. It should 
take in the same parameters as in *assign order*.

If the executor has already aquired the order, it should not 
reassign the order, but return failure indicator instead.

### Aqcuire order

The service must implement an endpoint which will be 
called by executor to aqcuire an order. It must take 
in ID of an executor and have the following behaviour:

1. Get order for this executor which 
was assigned in *assign order* or *update order* from 
the database.
2. Put and indicator that this executor acquired the order in the database.
3. Return info about acquired order

In case there is no order assigned to the requested executor, it 
it must return immediately with an indicator of this.

The executor is expected to poll this endpoint every minute 
before he gets an order. We account 
for the number of executors to be 250k, which 
gives us about 4.2k RPS of estimated load for this endpoint.

### Cancel order

The service must implement an endpoint which will be 
called to cancel an order. It must take in an ID of the 
order and must do the following:

1. Delete information about the order and it's assigned executor from the database
2. Return information about the deleted order

In case order has been already acquired by the executor, it must 
return an indicator of failure.

There must be a safety check in place i. e. after 10 minutes passed 
since an assignment of the order, it can't be canceled and 
failure indicator must be returned. 

We expect about 40% of assigned orders to be canceled. 
Taking in account RPS for *assign order* we account for 
120 RPS load for this endpoint.

## Consequences

* This service can be scaled independently from the service 
which makes assignment decisions, which is good because loads for these 
services can be very different.
* It improves reliability, because even if assignment service fails, 
executors can still get their assigned orders from Order Service.