<h1><img src="logo.svg" alt="drawing" style="width: 150px; vertical-align: middle"/>APKover</h1>

APKover is a tool to extract the coverage a Wolfi package as the percentage of statements of the packaged software
that are covered by the [Melange](https://github.com/chainguard-dev/melange/) tests.

It leverages the language specific instrumentation features to generate integration tests coverage,
by re-building the package to enable the software measure the coverage data at runtime.
Finally, it reports the overall package test coverage as a percentage to standard output.

A minimum threshold can be set to make APKover exit 1 when a package does not meet the coverage criteria.

```
Usage:
  apkover [flags]

Flags:
  -c, --config string      path to package config file
      --fail-under float   The minimum accepted coverage, expressed as percentage (e.g. 80 for 80% of coverage). Fail if it's under the specified threshold.
  -h, --help               help for apkover
      --language string    main language of the package (default "go")
      --log-level string   log level (default "info")
  -o, --output string      output format (text, json, yaml) (default "text")
```

## Requirements
- GNU `make`
- [Melange](http://github.com/chainguard-dev/melange/)
- A local copy of a [Wolfi packages repository](https://github.com/wolfi-dev/os/)

## Quickstart

Specify a Wolfi package Melange config to the `--config` flag.

```shell
$ apkover --config wolfi-dev/os/crane.yaml
2025-02-16T19:11:15+01:00 INF Updating the build pipeline to instrument the package package=crane
2025-02-16T19:11:15+01:00 INF Updating the test pipeline to generate coverage data package=crane
2025-02-16T19:11:15+01:00 INF Writing the package config to disk package=crane
2025-02-16T19:11:15+01:00 INF Re-building the package instrumented package=crane steps=2
2025-02-16T19:11:15+01:00 INF Running tests and writing coverage data package=crane steps=7

 Test Coverage: [█████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 14%
```

### Fail on low coverage

Now imagine you want to ensure a package test coverage and you want a quality gate in continuous integration.
You can specify a minimum threshold as a percentage to the `--fail-under` CLI flag to make `apkover` exit 1 when the
coverage is below that percentage:

```shell
$ apkover --config wolfi-dev/os/crane.yaml --fail-under 20 2>/dev/null

 Test Coverage: [█████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 14%

❌ test coverage is below the minimum required
exit status 1
```

### Monitoring the coverage over time

The coverage data can be easily consumed as metrics and monitored over time using machine readable formats like JSON:

```shell
$ apkover --config wolfi-dev/os/crane.yaml --output=json 2>/dev/null | jq
{
  "pkg_name": "crane",
  "pkg_version": "0.20.3",
  "pkg_epoch": 2,
  "cov_float": 0.14199999809265137
}
```

## Support

### Languages

The supported language now is:
* `go`

Support for additional language is yet to be implemented.

#### Go

APKover leverages the Go [Coverage profiling support for integration tests](https://go.dev/doc/build-cover)
to build the Go binary instrumented for measurement and coverage data analysis with the `covdata` Go tool.

### Packages

Currently only the main package of a Melange pipeline is supported.

Support for subpackages is yet to be implemented.

