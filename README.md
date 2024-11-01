## Files and directory structure

`architecture` &mdash; contains all ADRs for this project

`cache-service` &mdash; source code of the microservice for 
caching data sources (see [ADR-001](architecture/adr-001-cache-service.md))

`order-service` &mdash; source code of the microservice responsible 
for assigning and aquiring of orders (see [ADR-002](architecture/adr-002-order-service.md))

`docker` &mdash; all the dockerfiles for this project, e. g. for `cache-service` and `order-service`

`infra` &mdash; everything needed for getting the project up in the cloud

`tests` &mdash; integration tests

`compose.yaml` &mdash; docker-compose file for local testing