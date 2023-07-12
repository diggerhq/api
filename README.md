The repository for digger official API.

# How to run for development

```
# run the database
docker-compse up
# create .env file with creds
echo "DATABASE_URL=postgres://postgres:23q4RSDFSDFS@127.0.0.1:5432/postgres" >> .env
make start
```
