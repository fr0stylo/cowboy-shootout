# Cowboy shootout simulation

## Description

This repository simulates cowboy shootout. Cowboys shoots each other with one second intervals. Last standing cowboy wins.
Few rules:
 * Cowboy can't shoot himself
 * Cowboy won't shoot dead cowboy (why should he waste any bullets?)
 * Cowboy takes random target
 * All cowboys work in isolated process 

## Development

### Tooling

Tools used:
  * Go 1.17
  * Docker
  * Docker compose
  * Make (optional)

Persistance layer:
  * File (as long term persistance)
  * Redis (as distributed execution time storage)

Communication:
  * Redis Pub/Sub

### Running simulation

#### Docker compose

Application can be run using docker compose. 
In repository root hit `docker compose up` command. This wil initiate docker compose and build application and will create redis instance. 
All logs will be seen in docker compose logs.

#### Running locally

In order to run locally you'll need to have access to redis instance. 

If running redis locally on `6379` port run `make run` this will run service locally

Else run `REDIS_CONNECTION_STRING=<redisURL> make run`


### Design

`starter` is control plane for cowboys. Cowboys are spawned dynamically with help of `shooters.json`. On initialization shooters are read and pushed to redis. Redis key is cowboy name. Each key spawns new cowboy process `shooter` which act like cowboy and exits when health reaches 0, all communication are done via redis pub/sub. Logging is done in central place - `starter`. All logs are pushed through redis pub/sub. All this implementation is design in a way that `shooter` can by spawned dynamically and can run in any sort of containerized environment (kubernetes, dockerswarn, ECS, etc....)


