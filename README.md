# Gator
This project is a CLI tool for managing feeds RSS.
It accepts different user and follows feeds.
It's made following Boot.dev course

# Installation
You need to have Go and Postgres installed to be able to use Gator.
After installing Go and Postgres, you should install `goose` (a migration tool) using `go install github.com/pressly/goose/v3/cmd/goose@latest`.

Create a config file name `.gatorconfig.json` at the root of your machine with the following content:

```json
{
  "db_url": "postgres://Postgres:@localhost:5432/DB_NAME?sslmode=disable",
  "current_user_name": "Your 	"
}
```
> Replace `Postgres` with your Postgres username, `DB_NAME` with your database name.
The `sslmode=disable` part is optional but recommended for local development.

Then create the database named `DB_NAME` in Postgres.

Then run `goose postgres "postgres://Postgres:@localhost:5432/DB_NAME" up` to create the database schema.
> Replace `Postgres` with your Postgres username and `DB_NAME` with your database name.


## Usage
To use Gator, run the following command (if you built it from source):
```
gator command [arguments]
```
Or 
```
go run . command [arguments]
```

List of commands:
- `register <username>`: Register a new user.
- `login <username>`: Log in as an existing user.
- `reset`: Reset the database. (WARNING: This will delete all data)
- `users`: List all users. It shows the current user with an `*`.
- `feeds`: List all feeds.
- `addfeed <feed_url>`: Add a feed.
- `follow <feed_url>`: Follow a feed.
- `unfollow <feed_url>`: Unfollow a feed.
- `browse <limit>`: List all posts. Need to be a number.
- `agg <time_request>`: Fetch the older rss feed fetched, then aggregate to posts. It take a number (only one digit) and a unit of time (e.g. `1h` for 1 hour, `1m` for 1 minute, `1s` for 1 second).
