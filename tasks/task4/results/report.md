1. установлен minikube на ubuntu 24.04 тиа виртуализации docker
2. установлен Kuberctl 
3. создано 2 конфига values-prod.yaml values-staging.yaml
4. для загрузки собранного образа в среду minicube : 
```shell
eval $(minikube docker-env) 
docker build -t booking-service:latest ../booking-service/
```
5. для успешного пинга пришлось пробарсывать порт из кубера на 4567 локально, в скрине указано
6. 