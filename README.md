# qonto

[![Build Status](https://github.com/dohernandez/qonto/workflows/test-unit/badge.svg)](https://github.com/dohernandez/qonto/actions?query=branch%3Amaster+workflow%3Atest-unit)
[![Coverage Status](https://codecov.io/gh/dohernandez/qonto/branch/master/graph/badge.svg)](https://codecov.io/gh/dohernandez/qonto)
![Code lines](https://sloc.xyz/github/dohernandez/qonto/?category=code)
![Comments](https://sloc.xyz/github/dohernandez/qonto/?category=comments)

Service that performance transfers.

## Table of Contents
- [Table of Contents](#table-of-contents)
- [Overview](#overview)
- [Getting started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Development](#development)
        - [Running the service locally](#running-the-service-locally)
        - [Generate code from proto file](#generate-code-from-proto-file)
    - [Testing](#testing)
    - [Metrics](#metrics)
    - [Migrations](#migrations)
- [Enhancement](#enhancement)
- [Timing](#timing)
- [Proud](#proud)

## Overview

The current architecture of the service is described in the [ARCHITECTURE.md](./ARCHITECTURE.md) document.

For more in-depth explanations and considerations on architectural choices for the service, please refer to our [Architecture Decision Records](./resources/adr) folder.

If you want to submit an architectural change to the service, please create a new entry in the ADR folder [using the template provided](./resources/adr/template.md) and open a new Pull Request for review. Each ADR should have a prefix with the consecutive number and a name. For example `002-implement-server-streaming-rpc-get-leaderboard.md`

## Getting started

### Prerequisites

To run this application on your machine, you must have `docker` && `docker-compose` installed.

The service uses `dep` to manage its dependencies. All the dependencies can be installed using the following `make` command:

```shell
make deps
```

[[table of contents]](#table-of-contents)

### Development

#### Running the service locally

Run app with `docker-compose` dependencies.

First generate an `.env` file which the environment values required by the service such as `APP_GRPC_PORT` and `DATABASE_DSN`. You can run the following `make` command:

```
make envfile
```

This command will generate the `.env` file from the `.env.template`. Make sure the env variables defined in the file `.env` meet your expectation.

**Note:** No need to edit the `.env.template` before running the command, the flow is that you generate the `.env` file from `.env.template` and after that edit the `.env` if needed.

After the `.env` file is generated, you can start the app by running:

```shell
docker-compose up -d
```

Since we are using the option `-d` make sure the service is up and running. Checks the logs of the services:

```shell
docker-compose logs -f app
```

The output should look something like this

```shell
qonto-app  | 2021/11/30 14:55:10 Running build command!
qonto-app  | 2021/11/30 14:55:19 Build ok.
qonto-app  | 2021/11/30 14:55:19 Restarting the given command.
qonto-app  | 2021/11/30 14:55:19 stdout: 2021-11-30T14:55:19.954Z       INFO    qonto/main.go:52        start Metrics server at addr [::]:8010
qonto-app  | 2021/11/30 14:55:19 stdout: 2021-11-30T14:55:19.954Z       INFO    qonto/main.go:52        start GRPC server at addr [::]:8000
qonto-app  | 2021/11/30 14:55:19 stdout: 2021-11-30T14:55:19.964Z       INFO    qonto/main.go:52        start REST server at addr [::]:8080

```

To destroy the app run the command:

```shell
docker-compose down -v
```

[[table of contents]](#table-of-contents)

#### Generate code from proto file

The service implements grpc based on the proto definition. The proto file with the service definition can be found `resources/proto/service.proto`.

To generate the go code based on the proto definition run the `make` command

```shell
make proto-gen-code
```

The Go files generated based on the command can be found `pkg/proto` folder.

[[table of contents]](#table-of-contents)

### Testing

The server follows unit testing. Unit testing make sure the logic of the application is sounds.

Unit tests reside with the application source code, as per Golang recommendation. Use the command `make` command to run the tests:

```shell
make test
```

You can also run

```shell
make lint
``` 

to make sure your changes follow our coding standards.

**Note:** Integration testing is part of the future enhancement.

#### Evans

For manual gRPC API inspection, the service allows gRPC reflection in dev environment.

To install Evans following the instructions from it GitHub page https://github.com/ktr0731/evans#installation.

[[table of contents]](#table-of-contents)

#### REST

Also, you can do test thro REST calls. For that you can use the service REST api documentation which uses Swagger interface.

Launch the service by [Running the service locally](#running-the-service-locally). This will make the service available in http://localhost:8080 (remember that the port is base on the configuration you provide in the `.env` file. This example is based on the `.env.template` configuration) and REST api documentation can be accessible on http://localhost:8080/docs.

[[table of contents]](#table-of-contents)

### Metrics

The service exposes some metrics such as:

- Database
- Go build info
- Current Go process
- Calls started/completed
- Histogram of response latency (seconds).

Metrics are available on http://localhost:8010/metrics

[[table of contents]](#table-of-contents)

### Migrations

Database migrations are stored in [`resources/migrations`](./resources/migrations) folder.

Migrations are run using [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate) tool,
embedded in the service's `Dockerfile` under `/bin/migrate`.

Each migration should have an `<name>.up.sql` and `<name>.down.sql` variants, further information can be seen here: https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md

The layout of the migration name should be as follows:

```
<current-date-string>_<migration-name>.(up|down).sql

Example (created in 2020-01-01 00:00:00):

    20200101000000_my_migration.up.sql
    20200101000000_my_migration.down.sql
```

For creating migration files you can use the following `make` command:

```shell
make create-migration NAME=<migration-name>
```

To run migration up use the following `make` command:

```shell
make migrate
```

**Note:** `DATABASE_DSN` env variable should be defined. In case it does not you can use the following `make` command:

```shell
make env migrate
```

If you run the above command from outside to docker network, make sure to have `127.0.0.1 postgres` in `/etc/hosts`.

[[table of contents]](#table-of-contents)

### Enhancement

- Increase request validation.
- Increase the unit test suite.
- Add grpc Server streaming to allow better communication.
- Add behavioral test (integration testing) to reduce the manual test.

[[table of contents]](#table-of-contents)

### Timing

The service took 5h to 6h to be complete at the current stage. The main reason it took this time was due to adding the basic test for main the structs used in the logic to performance transaction in bulk.


[[table of contents]](#table-of-contents)

### Proud

I feel very proud with the current service stage. I have enjoined a lot while building this service even tho the timing was a bit stressful.


[[table of contents]](#table-of-contents)