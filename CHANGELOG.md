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
