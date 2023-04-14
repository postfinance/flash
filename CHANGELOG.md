## 0.5.1 (2023-04-14)


### Bug Fixes

* **common**: use flash prometheus prefix if appName is empty ([ac6f92ac](https://github.com/postfinance/flash/commit/ac6f92ac))


### Build System

* **deps**: github.com/mattn/go-isatty 0.0.16 -> 0.0.18 ([a519e70b](https://github.com/postfinance/flash/commit/a519e70b))
* **deps**: github.com/prometheus/client_golang 1.14.0 -> 1.15.0 ([8a64034e](https://github.com/postfinance/flash/commit/8a64034e))
* **deps**: github.com/stretchr/testify 1.8.1 -> 1.8.2 ([5e6987d0](https://github.com/postfinance/flash/commit/5e6987d0))
* **deps**: go.uber.org/zap 1.23.0 -> 1.24.0 ([0a1f4574](https://github.com/postfinance/flash/commit/0a1f4574))
* **deps**: gopkg.in/natefinch/lumberjack.v2 2.0.0 -> 2.2.1 ([a42bc762](https://github.com/postfinance/flash/commit/a42bc762))



## 0.5.0 (2022-11-28)


### Build System

* **deps**: github.com/prometheus/client_golang 1.13.0 -> 1.14.0 ([0ce5469d](https://github.com/postfinance/flash/commit/0ce5469d))
* **deps**: github.com/stretchr/testify 1.8.0 -> 1.8.1 ([031c78f0](https://github.com/postfinance/flash/commit/031c78f0))


### New Features

* **common**: add WithoutTimestamps option to log without timestamps ([8d69956f](https://github.com/postfinance/flash/commit/8d69956f))
* **common**: add logfmt support ([da95c21b](https://github.com/postfinance/flash/commit/da95c21b))



## 0.4.0 (2022-09-16)


### Build System

* **deps**: github.com/prometheus/client_golang 1.11.0 -> 1.13.0 ([c210a1cd](https://github.com/postfinance/flash/commit/c210a1cd))
* **deps**: github.com/stretchr/testify 1.7.0 -> 1.8.0 ([ddd957d9](https://github.com/postfinance/flash/commit/ddd957d9))
* **deps**: go.uber.org/zap 1.19.1 -> 1.23.0 ([efe8d230](https://github.com/postfinance/flash/commit/efe8d230))


### New Features

* **common**: use json encoder when no terminal is detected ([176efc60](https://github.com/postfinance/flash/commit/176efc60))



## 0.3.0 (2022-03-18)


### New Features

* **common**: add new option `WithEncoder` to configure encoder ([ca334c91](https://github.com/postfinance/flash/commit/ca334c91))
* **common**: add support for logging into file ([dfe0daa0](https://github.com/postfinance/flash/commit/dfe0daa0))



## 0.2.0 (2020-12-29)

### Bug Fixes

- **common**: adjust stacktrace level on level change with `SetLevel` ([f36e3421](https://github.com/postfinance/flash/commit/f36e3421))

### New Features

- **common**: renamed WithSink option to WithSinks and changed it to be a variadic function (instead of only accepting a string) ([e76679c0](https://github.com/postfinance/flash/commit/e76679c0))

## 0.1.0 (2020-12-23)
