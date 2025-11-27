## Run Postgres

```
docker run --name casdoor-postgres -d \
-p 5432:5432 \
-e POSTGRES_USER=postgres \
-e POSTGRES_PASSWORD=postgres \
-e POSTGRES_DB=casdoor \
postgres:15-alpine
```


## After creating a proto file, we will do this:

```aiignore
cd ./proto
buf dep update
buf lint
buf generate
```