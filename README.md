# SnapDB

SnapDB is a professional-grade database backup and restore tool with both a traditional CLI and an interactive shell.

```bash
snapdb backup mysql
snapdb restore mysql-20260609T020000
snapdb list
snapdb version
```

Interactive mode starts when `snapdb` is run with no arguments:

```text
>>> snapdb version
SnapDB v0.1.0
>>> snapdb help
>>> exit
```

Interactive commands must start with `snapdb`. Use `history` inside the shell to view commands from the current session.

## Features

- Cobra-based CLI with interactive shell parity.
- Backup support for MySQL/MariaDB, PostgreSQL, MongoDB, and Redis through native tools.
- Restore support for MySQL, PostgreSQL, and MongoDB.
- Backup metadata with IDs, timestamps, file sizes, checksums, status, and profile names.
- Backup management commands: `list`, `delete`, `stats`, `verify`, and `cleanup`.
- Config profiles through `snapdb config init/show/set`.
- Compression formats: `gzip`, `zstd`, and `zip`.
- AES-256-GCM encryption with a passphrase supplied through an environment variable.
- Upload staging for `local`, `s3`, `gcs`, and `azure` providers under SnapDB local state.
- Schedule definitions through `snapdb schedule create/list/delete`.

## Install

```bash
go build -o snapdb .
./snapdb version
```

Optional database tools must be installed separately depending on the database:

- MySQL/MariaDB: `mysqldump`, `mysql`
- PostgreSQL: `pg_dump`, `psql`
- MongoDB: `mongodump`, `mongorestore`
- Redis: `redis-cli`

## Configuration

Initialize local SnapDB state:

```bash
snapdb config init
```

Create a profile:

```bash
export SNAPDB_MYSQL_PASSWORD='your-password'

snapdb config set \
  --name local-mysql \
  --default \
  --type mysql \
  --host localhost \
  --port 3306 \
  --user root \
  --password-env SNAPDB_MYSQL_PASSWORD \
  --dbname appdb
```

Show config:

```bash
snapdb config show
```

SnapDB stores config and metadata under `~/.snapdb` with restricted file permissions. Profiles store `passwordEnv` rather than plaintext passwords.

## Backup

Using a profile:

```bash
snapdb backup --profile local-mysql
```

Using direct flags:

```bash
snapdb backup mysql \
  --host localhost \
  --port 3306 \
  --user root \
  --password-env SNAPDB_MYSQL_PASSWORD \
  --dbname appdb
```

Compression:

```bash
snapdb backup postgres --profile local-postgres --compress --format gzip
snapdb backup postgres --profile local-postgres --compress --format zstd
snapdb backup postgres --profile local-postgres --compress --format zip
```

Encryption:

```bash
export SNAPDB_ENCRYPTION_PASSWORD='strong passphrase'
snapdb backup mysql --profile local-mysql --encrypt
```

Upload staging:

```bash
snapdb backup mysql --profile local-mysql --upload s3
snapdb backup mongodb --profile local-mongo --upload gcs
snapdb backup postgres --profile local-postgres --upload azure
```

Provider uploads are currently staged under `~/.snapdb/cloud/<provider>` so the command flow is testable locally before wiring production cloud credentials.

## Restore

Restore by backup ID:

```bash
snapdb restore mysql-20260609T020000 --profile local-mysql
```

Restore by file:

```bash
snapdb restore ~/.snapdb/backups/appdb_mysql-20260609T020000.sql --profile local-mysql
```

Encrypted files are decrypted to a temporary file during restore when `SNAPDB_ENCRYPTION_PASSWORD` is set.

## Backup Management

```bash
snapdb list
snapdb verify <backup-id-or-file>
snapdb stats
snapdb delete <backup-id>
snapdb cleanup --older-than 720h
```

Example list output:

```text
ID                    DATABASE  TYPE      CREATED           SIZE     STATUS
mysql-20260609T020000 appdb     mysql     2026-06-09 02:00  42.0 MiB completed
```

## Scheduling

Create schedule definitions:

```bash
snapdb schedule create --every 24h --profile local-mysql
snapdb schedule create --cron "0 2 * * *" --profile local-postgres
```

Manage schedules:

```bash
snapdb schedule list
snapdb schedule delete <schedule-id>
```

Schedules are persisted as definitions in `~/.snapdb/schedules.json`. A long-running scheduler/daemon can be added on top of this stored model.

## Security Notes

- SnapDB does not write database passwords into config files; use `--password-env`.
- Encrypted backups use AES-256-GCM with keys derived by scrypt.
- Commands use structured process execution instead of shell command strings.
- Config, metadata, schedules, backups, and staged cloud uploads are written with owner-only permissions where applicable.
- Command output avoids printing raw connection strings or credentials.

## Roadmap

- Production cloud SDK integrations for S3, GCS, and Azure Blob Storage.
- Long-running schedule runner or OS scheduler integration.
- Rich REPL tab completion using a terminal line editor.
- Database-specific incremental and differential backup implementations.
- More automated command tests and fixture-based restore tests.