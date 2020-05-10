# mat

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
