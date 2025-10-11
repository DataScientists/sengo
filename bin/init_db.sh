
#!/bin/bash
docker compose cp ./docker/postgres_data/sql postgres:/var/lib/postgresql/data/
docker compose exec postgres bash -c "psql -U root -f /var/lib/postgresql/data/sql/reset_database.sql"
