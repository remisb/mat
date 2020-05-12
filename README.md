# mat

Restaurant menu voting and management REST api service implemented on golang.

Application is done to be fully configurable and to mach ["The Twelve-Factor App"] requirements. 
All application work related properties are externalized and app work can be configured with
configuration file, CLI flags or environment apps.

TBD 

## Requirement

* Docker container
* Golang 1.14

## Project setup / development and usage instructions

* Install Docker on your local machine - [Get Docker](https://docs.docker.com/get-docker/)
* Install and setup golang. v 1.14 required - [golang Gettins Started](https://golang.org/doc/install)
* Git [Getting Started - Install Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
* Install Gnu make (optional) - makes life a little bit simpler.

### Provided make targets

* TBD

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
