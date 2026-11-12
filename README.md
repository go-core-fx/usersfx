<a id="readme-top"></a>

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![Apache 2.0 License][license-shield]][license-url]

<br />
<div align="center">
  <h1>usersfx</h1>
  <p>
    Users domain module for <code>go-core-fx</code> — registration and login by Telegram identity, user lookup, an Uber Fx module and a bun-backed repository.
  </p>
  <a href="https://github.com/go-core-fx/usersfx/issues/new">Report Bug</a>
  &middot;
  <a href="https://github.com/go-core-fx/usersfx/issues/new">Request Feature</a>
</div>

## Table of Contents
- [Table of Contents](#table-of-contents)
- [About The Project](#about-the-project)
  - [Features](#features)
  - [Built With](#built-with)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Usage](#usage)
  - [Module Wiring](#module-wiring)
  - [Providing the Repository](#providing-the-repository)
- [Configuration](#configuration)
- [Development](#development)
  - [Make Targets](#make-targets)
  - [Continuous Integration](#continuous-integration)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)


---

## About The Project

`usersfx` is the users domain for the [go-core-fx](https://github.com/go-core-fx) ecosystem. The root package holds the domain (`User`, `Identity`), the `Service` (`RegisterOrLogin`, `Login`, `GetUser`), the `Repository` port and the sentinel errors; `Module()` registers the [Uber Fx](https://uber-go.github.io/fx/) module `"users"` with a named logger. Persistence ships as the nested [`bun`](./bun) module:

- **Domain and port** — `domain.go` / `errors.go` / `repository.go`; the service never imports a storage driver
- **Fx module** — `Module()` registers module `"users"` and decorates the injected `*zap.Logger` with that name
- **bun repository** — the `github.com/go-core-fx/usersfx/bun` submodule maps `users` / `user_identities` tables (timestamps via `bunfx.TimedModel`), registers a user and identity in one transaction and handles the concurrent-registration race
- **Strict lint config** — a v2.14.0 golangci-lint configuration with a large strict-but-practical linter set
- **Makefile + CI** — formatting, linting, race-enabled tests, coverage, benchmarks, nested-module checks, and GitHub Actions workflows that run them

This is a library, not an application: it reads **no environment variables** of its own and has no configuration file. `DEBUG` is read by the [`go-core-fx/logger`](https://github.com/go-core-fx/logger) dependency, and `TZ`/`DEBUG` are set for the development server by `make air` — see [Configuration](#configuration).

### Features

- **User registration and login by identity** — `Service.RegisterOrLogin` validates the identity, returns an existing user or creates one atomically; duplicate identities from concurrent requests resolve to the existing user
- **User lookup** — `Service.GetUser` by user ID; all failures surface as wrapped sentinel errors (`ErrValidationFailed`, `ErrNotFound`, `ErrIdentityExists`)
- **bun persistence submodule** — `bun.NewRepository(db)` implements the `usersfx.Repository` port over an injected `*bun.DB`; the adapter is bound to the port by the consuming application, not by this module
- **Strict golangci-lint** — the same golden config used across the go-core-fx modules
- **CI pipeline** — lint, race-enabled tests with coverage upload to Codecov, and benchmark tracking with alerting

### Built With

- [![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
- [![Uber Fx](https://img.shields.io/badge/Uber%20Fx-000000?style=for-the-badge)](https://uber-go.github.io/fx/)
- [![bun](https://img.shields.io/badge/github.com%2Fuptrace%2Fbun-00ADD8?style=for-the-badge)](https://bun.uptrace.dev/)
- [![zap](https://img.shields.io/badge/go.uber.org%2Fzap-00ADD8?style=for-the-badge)](https://github.com/uber-go/zap)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.24.3 or newer for the root module (`go.mod` requires `go 1.24.3`); the [`bun`](./bun) submodule requires `go 1.26.1`
- [golangci-lint](https://golangci-lint.run/) v2 (the config targets v2.14.0) for `make lint` and `make fmt`
- GNU Make

### Installation

```bash
go get github.com/go-core-fx/usersfx
```

Add the persistence submodule only if you need it:

```bash
go get github.com/go-core-fx/usersfx/bun
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Usage

### Module Wiring

`Module()` provides the `*Service` and decorates the logger; the application must also provide a `usersfx.Repository` (the port is bound to its adapter in the composition root):

```go
fx.New(
    logger.Module(),
    bunfx.Module(),
    usersfx.Module(),
    // fx.Provide(...): *sql.DB, schema.Dialect, usersfx.Repository
)
```

### Providing the Repository

The bun submodule implements the port over an injected `*bun.DB`:

```go
import usersfx "github.com/go-core-fx/usersfx"
import usersbun "github.com/go-core-fx/usersfx/bun"

fx.Provide(func(db *bun.DB) usersfx.Repository {
    return usersbun.NewRepository(db)
})
```

The consuming application owns the driver and dialect: it opens the `*sql.DB`, supplies the matching `schema.Dialect` to `bunfx.New`, and owns schema migrations for the `users` / `user_identities` tables.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Configuration

This module itself reads no environment variables. The table below lists everything the repository and its direct dependency consume:

| Variable        | Read by                                       | Purpose                                                                                             | Format                                                                     | Default                                        | Example              |
| --------------- | --------------------------------------------- | --------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- | ---------------------------------------------- | -------------------- |
| `DEBUG`         | `github.com/go-core-fx/logger` (`logger.New`) | Selects the zap development logger (console encoding, debug level) instead of the production logger | Boolean accepted by `strconv.ParseBool` (`1`, `t`, `T`, `TRUE`, `true`, …) | unset → treated as `false` → production logger | `1`                  |
| `TZ`            | `make air` (Makefile)                         | Timezone of the development server process                                                          | IANA timezone name                                                         | system local timezone                          | `UTC`                |
| `CODECOV_TOKEN` | Codecov GitHub Action (CI only)               | Uploads coverage reports from the `test` workflow job                                               | Opaque token, stored as a GitHub Actions secret                            | not set                                        | `your-codecov-token` |

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Development

### Make Targets

| Target           | Description                                                            |
| ---------------- | ---------------------------------------------------------------------- |
| `make all`       | Run all checks: `fmt`, `lint`, `coverage`, `nested`                      |
| `make version`   | Display the current version (from `git describe`, defaults to `0.0.0`) |
| `make fmt`       | Run `go generate ./...`, then `golangci-lint fmt`                      |
| `make lint`      | Run `golangci-lint run --timeout=5m`                                   |
| `make test`      | Run tests with race detector, shuffle and atomic coverage              |
| `make coverage`  | Produce `coverage.out` and an HTML report (`coverage.html`)            |
| `make nested`    | Build, vet, test and lint every nested module (auto-discovered `go.mod` directories below the root) |
| `make benchmark` | Run benchmarks into `benchmark.txt`                                    |
| `make air`       | Run the development server with air (`TZ=UTC DEBUG=1`)                 |
| `make deps`      | Download module dependencies                                           |
| `make gen`       | Run `go generate ./...`                                                |
| `make release`   | Create a snapshot release with goreleaser                              |
| `make build`     | Build the project into `bin/`                                          |
| `make clean`     | Remove coverage output, benchmarks, `dist/` and `bin/`                 |
| `make help`      | List all available targets                                             |

### Continuous Integration

The [Go workflow](.github/workflows/go.yml) runs on pushes to `master` and on pull requests (path-filtered to Go sources, `go.mod`/`go.sum`, the lint config and the workflows themselves):

- **Lint** — golangci-lint with the repository config
- **Test** — race-enabled, shuffled tests with coverage uploaded to [Codecov](https://codecov.io/)
- **Benchmark** — runs `go test -bench`, keeps history in the Actions cache, uploads results when benchmarks exist and fails the job on a performance alert
- **E2E Test** — present but disabled (`if: false`)

The [Nested modules workflow](.github/workflows/modules.yml) runs `make nested` on the same triggers: root patterns like `./...` stop at a nested module's boundary, so each module (e.g. [`bun`](./bun)) gets its own build, vet, test and lint pass. Discovery lives in the Makefile — any directory with its own `go.mod` is picked up automatically.

A second workflow marks issues and pull requests stale after 7 days of inactivity and closes them 7 days later.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

Please run `make all` before opening the pull request — it formats, lints and tests the change.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## License

Distributed under the Apache License 2.0. See `LICENSE` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Acknowledgments

- [Uber Fx](https://uber-go.github.io/fx/) — dependency injection framework
- [go-core-fx/logger](https://github.com/go-core-fx/logger) — zap logger module this module decorates with its name
- [uptrace/bun](https://bun.uptrace.dev/) — SQL ORM behind the `bun` submodule
- [golangci-lint](https://golangci-lint.run/) — linter suite behind `make lint`
- [Best-README-Template](https://github.com/othneildrew/Best-README-Template) — README structure

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
[contributors-shield]: https://img.shields.io/github/contributors/go-core-fx/usersfx.svg?style=for-the-badge
[contributors-url]: https://github.com/go-core-fx/usersfx/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/go-core-fx/usersfx.svg?style=for-the-badge
[forks-url]: https://github.com/go-core-fx/usersfx/network/members
[stars-shield]: https://img.shields.io/github/stars/go-core-fx/usersfx.svg?style=for-the-badge
[stars-url]: https://github.com/go-core-fx/usersfx/stargazers
[issues-shield]: https://img.shields.io/github/issues/go-core-fx/usersfx.svg?style=for-the-badge
[issues-url]: https://github.com/go-core-fx/usersfx/issues
[license-shield]: https://img.shields.io/github/license/go-core-fx/usersfx.svg?style=for-the-badge
[license-url]: https://github.com/go-core-fx/usersfx/blob/master/LICENSE
