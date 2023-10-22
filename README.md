# exSTATic Server
This is a work in progress project to make [exSTATic](https://github.com/KamWithK/exSTATic) into a website

## Architecture
exSTATic uses the following services:
* [Fly.io](https://fly.io) for hosting
* [Turso](https://turso.tech) for production database (distributed [LibSQL](https://turso.tech/libsql))
* [LibSQL](https://turso.tech/libsql) for local/test file databases ([SQLite](https://www.sqlite.org) fork)
* [Atlas](https://atlasgo.io) for database migrations
* [Google OAuth 2.0](https://developers.google.com/identity/protocols/oauth2) for social login

## Local Development Overview
### Environment Variables
* The `.env` file is read by docker compose
* Copy the `.sample.env` file to `.env` and fill in any credentials

### Database Migrations
* To check the database migration status run `atlas migrate status --env local`
* To create a migration run `atlas migrate diff --env local MIGRATION_NAME`
* To execute migrations run `atlas migrate apply --env local`

Note: Local file databases are used for easy testability

### Running
* Run `air` and navigate to `http://localhost:8080/` for a live reloading setup
* Run `go test ./...` in the `backend` directory to run all tests
