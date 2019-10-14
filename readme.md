# 12 factor GoLang app

## Remember
The app runs on the path `/counter` in your browser!

*this is simple "workaround" because some browsers send a favicon.ico request to the root which ruins the counter timers

## Default environment variables
```
ENV COUNTER_RECHECKAMOUNT=20
ENV COUNTER_HEALTHCHECKTIME=500
ENV COUNTER_REDISHOST="redis:6379"
```
## Docker compose

run `docker-compose up` and visit `localhost:8080/counter`

## K8s

run
```
kubectl create ns redis
kubectl create ns counter
kubectl apply -f redis-deploy.yaml //in redis ns
kubectl apply -f application-deploy.yaml //in counter ns
```
If you run nginx ingress with certmanager, you can run the `ingress.yaml` too
