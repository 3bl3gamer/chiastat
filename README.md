## DB setup
```bash
sudo su - postgres
createuser chiastat -P  # with password "chia"
createdb chiastat_db -O chiastat --echo
psql chiastat_db -c "CREATE SCHEMA chiastat AUTHORIZATION chiastat"

go run migrations/*.go init
go run migrations/*.go
```
