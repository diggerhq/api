The repository for digger official API.

# How to run for development

1. Create the environment files for local development:
   2. `echo "DATABASE_URL=postgres://postgres:23q4RSDFSDFS@127.0.0.1:5432/postgres" > .env`
   3. `echo "DATABASE_URL=postgres://postgres:23q4RSDFSDFS@postgres:5432/postgres" > .env.docker-compose`
2. Start the Docker containers for the database and the API via `docker-compose up` or `docker-compose up -d` which should make it available from http://localhost:3100   
3. You can also run the API by typing `make start` which should make it available from http://localhost:3000