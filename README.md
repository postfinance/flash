[![Go Report Card](https://goreportcard.com/badge/github.com/postfinance/flash)](https://goreportcard.com/report/github.com/postfinance/flash)
[![Coverage Status](https://coveralls.io/repos/github/postfinance/flash/badge.svg?branch=main)](https://coveralls.io/github/postfinance/flash?branch=main)
[![Build Status](https://github.com/postfinance/flash/workflows/build/badge.svg)](https://github.com/postfinance/flash/actions)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/postfinance/flash)](https://pkg.go.dev/github.com/postfinance/flash)
# flash
Creates an opinionated zap logger.

## Adapting it as a `logr.Logger` instance

Shall you at some point need to pass a `logr.Logger` instance in your code, e.g. while writing code that uses the 
Kubernetes client library etc., you can use `go-logr/zapr` to wrap the logger, as follows:

```golang
package main

import (
	"github.com/go-logr/zapr"
	"github.com/postfinance/flash"
)

func main() {
	l := flash.New()
	z := zapr.NewLogger(l.Desugar())
        z.V(0).Info("I'm a zap logger complying with logr.Logger interface !")
}
```

## Logger Without Stacktrace
### New Logger
> Debug / Info / Error
```
2020-10-09T09:29:06.363+0200    INFO    test/main.go:26 message {"StackTrace": "off", "debug": false}
2020-10-09T09:29:06.364+0200    ERROR   test/main.go:27 message {"StackTrace": "off", "debug": false}
```

### Enable Debug
> Debug / Info / Error
```
2020-10-09T09:29:06.364+0200    DEBUG   test/main.go:33 message {"StackTrace": "off", "debug": true}
2020-10-09T09:29:06.364+0200    INFO    test/main.go:34 message {"StackTrace": "off", "debug": true}
2020-10-09T09:29:06.364+0200    ERROR   test/main.go:35 message {"StackTrace": "off", "debug": true}
```

### Disable Debug
> Debug / Info / Error
```
2020-10-09T09:29:06.364+0200    INFO    test/main.go:42 message {"StackTrace": "off", "debug": false}
2020-10-09T09:29:06.364+0200    ERROR   test/main.go:43 message {"StackTrace": "off", "debug": false}
```

> Fatal
```
2020-10-09T09:29:06.364+0200    FATAL   test/main.go:47 message {"StackTrace": "on", "debug": false}
exit status 1
```

## Logger With Stacktrace
### New Logger
> Debug / Info / Error
```
2020-10-09T09:29:11.889+0200    INFO    test/main.go:26 message {"StackTrace": "off", "debug": false}
2020-10-09T09:29:11.889+0200    ERROR   test/main.go:27 message {"StackTrace": "off", "debug": false}
```

### Enable Debug (Stacktrace on Error or above)
> Debug / Info / Error
```
2020-10-09T09:29:11.889+0200    DEBUG   test/main.go:33 message {"StackTrace": "off", "debug": true}
2020-10-09T09:29:11.889+0200    INFO    test/main.go:34 message {"StackTrace": "off", "debug": true}
2020-10-09T09:29:11.889+0200    ERROR   test/main.go:35 message {"StackTrace": "off", "debug": true}
main.main
        /export/home/sauterm/tmp/test/main.go:35
runtime.main
        /export/home/sauterm/.gimme/versions/go1.15.2.linux.amd64/src/runtime/proc.go:204
```

###  Disable Debug (Stacktrace on Fatal only)
> Debug / Info / Error
```
2020-10-09T09:29:11.889+0200    INFO    test/main.go:42 message {"StackTrace": "off", "debug": false}
2020-10-09T09:29:11.889+0200    ERROR   test/main.go:43 message {"StackTrace": "off", "debug": false}
```

> Fatal
```
2020-10-09T09:29:11.889+0200    FATAL   test/main.go:47 message {"StackTrace": "on", "debug": false}
main.main
        /export/home/sauterm/tmp/test/main.go:47
runtime.main
        /export/home/sauterm/.gimme/versions/go1.15.2.linux.amd64/src/runtime/proc.go:204
exit status 1
```

