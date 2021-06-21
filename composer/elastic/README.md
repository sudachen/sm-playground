
```bash
docker network create --subnet 192.168.69.0/24 common
docker volume create elastic
docker volume create kibana
docker-compose up -d
```

kibana link is http://192.179.69.102/
