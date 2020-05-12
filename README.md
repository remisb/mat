# mat

Restaurant menu voting and management REST api service implemented on golang.

Application is done to be fully configurable and to mach ["The Twelve-Factor App"] requirements. 
All application work related properties are externalized and app work can be configured with
configuration file, CLI flags or environment apps.

TBD 

## Project setup / development and usage instructions

* Install Docker on your local machine - [Get Docker](https://docs.docker.com/get-docker/)
* Install and setup golang. v 1.14 required - [golang Gettins Started](https://golang.org/doc/install)
* Git [Getting Started - Install Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
* Install Gnu make (optional) - makes life a little bit simpler.

## Features

- HTTP routing and middleware support is provided with help of [go-chi/chi](https://github.com/go-chi/chi) router. As a 
bonus advantage `chi` router has wide set of [core middlewares](https://github.com/go-chi/chi#middlewares) with 
additional [auxiliary middlewares & packages](https://github.com/go-chi/chi#auxiliary-middlewares--packages).
- Database support using Postgres.
- CRUD based pattern.
- Account signup and user management.
- Testing patterns.
- Use of Docker, Docker Compose, and Makefiles.
- TBD - ntegration with CircleCI for enterprise-level CI/CD.


## Requirement

This project contains two services and uses 3rd party services such as MongoDB and Zipkin. Docker is required to run this software on your local machine.

## Downloading The Project

You can use git clone to clone this repository to your computer.
```bash
git clone https://github.com/remisb/mat.git
```

## Installing Docker

Once the project is cloned, it is important to validate the installation. 
This project requires the use of Docker since images are created and run in a Docker-Compose environment.

[Installing Docker](https://docs.docker.com/get-docker/)

TBD If you are having problems installing docker reach out or jump on `Gopher Slack` for help.

## Running Tests

With Docker installed, navigate to the root of the project and run the test suite.

NOTE - at the current time. Few tests are failing. Those failing test cases will be reviewed and updated. 
Nevertheless application is working correctly.
 
```bash
$ make test
```

## Building And Running

A makefile has also been provide to allow building, running and testing the software easier.

## Building The Project

Navigate to the root of the project and use the makefile to build all of the services.

```bash
$ cd $GOPATH/src/github.com/ardanlabs/service
$ make all
```

## Running the project

Navigate to the root of the project and use the makefile to run all of the services.

```bash
$ cd mat
$ make up
```

The make up command will leverage Docker Compose to run all the services, including the 3rd party services. 
The first time you run this command, Docker will download the required images for the 3rd party services.

Default configuration is set for the developer environment which should be valid for most systems. 
Use the `docker-compose.yaml` file to configure the services differently if necessary.

## Stopping the project

You can hit C in the terminal window running make up. Once that shutdown sequence is complete, it is important to 
run the make down command.

```bash
$ <ctrl>C
$ make down
```

Running make down will properly stop and terminate the Docker Compose session.

### Provided make targets

* `make up`- starts docker compose containers used for development
* `make down` - 
* `make migrate` - Migrate attempts to bring the schema for db up to date with the migrations
* `make seed`  - Seed runs the set of seed-data queries against db. The queries are ran in a transaction and rolled back if any fail.
 

## Project is based on the following tools, technologies and packages

* [Docker container](https://www.docker.com)
* [PostgreSql](https://www.postgresql.org/)
* [viper - Go configuration with fangs](https://github.com/spf13/viper)
* [Go-chi - lightweight, idiomatic and composable router for building Go HTTP services](https://github.com/go-chi/chi)
* [Go-chi - JWT authentication middleware for Go HTTP services](https://github.com/go-chi/jwtauth)
* [uber zap logger - Blazing fast, structured, leveled logging in Go](Uber Zap logger)
* [GNU Make](https://www.gnu.org/software/make/)
* [httpexpect - End-to-end HTTP and REST API testing for Go.](https://github.com/gavv/httpexpect)

Project consist two administration and management CLI executable files `admin` and `rest-api`. 
Source code of those executables is located in:

* cmd/mat-admin - admin application
* cmd/rest-api - restaurant REST-API microservice 

## How to setup and start development

Project uses PostgreSql database for data storage.

Project setup and common development task can be invoked through makefile targets. 
Bellow is provided a list of default makefile targets.

```make
down                           Stops Docker containers and removes containers, networks, volumes, and images created by up.
keys                           Generate private key file to private.pem file
migrate                        Migrate attempts to bring the schema for db up to date with the migrations defined.
seed                           Seed runs the set of seed-data queries against db. The queries are ran in a transaction and rolled back if any fail.
up                             Builds, (re)creates, starts, and attaches to Docker containers for a service.
```

## Database setup

Start PostgreSql docker container. 
All tasks are dependent on the postgreSql docker container, before performing any other task `make up` 
should be executed and DB container should accept incomming DB connections on the default PostgreSql port 5432. 

```bash
> make up
```

To create default DB schema.

```bash
> make migrate
``` 

To fill database with default testing data. 

```bash
> make seed
```

## ToDo

- [x] Finish logging to external file
- [ ] write documentation
- [ ] write config file documentation
- [ ] write CLI flag documentation
- [ ] write ENV variable documentation 
- [ ] update / improve app flag documentation
- [ ] add swagger documentation
- [ ] extend / update makefile to improve development experience
- [ ] review / cleanup / update makefile

## Requirements

- 1 dfaslkdf
- 2 sadkfjasd;lkf

## Background

REST API service based on Chi Routes and mix of solutions from  Mat Ryer book Go Programming Blueprints and ardanlabs service repository.

References to:
* Mat Rayer
* [go-chi](https://github.com/go-chi/chi) router
* [ardanlabs service](https://github.com/ardanlabs/service)


### Development tips

In case if db docker container is not cleaned up properly. That may happen in the test development process, you may have to `stop` and `remove` postgresql db container.

Bash shell
```bash
docker stop $(docker ps -aq)
docker rm $(docker ps -aq)
```

Fish shell
```fish
docker stop (docker ps -aq); docker rm (docker ps -aq)
```
