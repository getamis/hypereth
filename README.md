# hypereth

[![License: LGPL v3](https://img.shields.io/badge/License-LGPL%20v3-blue.svg)](https://www.gnu.org/licenses/lgpl-3.0)
[![Build Status](https://travis-ci.com/getamis/hypereth.svg?branch=master)](https://travis-ci.com/getamis/hypereth)
[![codecov](https://codecov.io/gh/getamis/hypereth/branch/master/graph/badge.svg)](https://codecov.io/gh/getamis/hypereth)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://GitHub.com/Naereen/StrapDown.js/graphs/commit-activity)

The Ethereum ecosystem.

## Getting Started

`hypereth` contains multiple Ethereum-related tools and libraries. All tools are combined into single executable called `hypereth`.

### Tools

* metrics-exporter

  A tool that exports metrics from an Ethereum client. For the details, please see [README](cmd/hypereth/metrics-exporter/README.md)

### Installing

```
$ go get github.com/getamis/hypereth
```

or

```
$ git clone git@github.com:getamis/hypereth.git
```

### Build

```
$ make
```

### Usage

Please see `README.md` for each package.

## Contributing

There are several ways to contribute to this project:

1. **Find bug**: create an issue in our Github issue tracker.
2. **Fix a bug**: check our issue tracker, leave comments and send a pull request to us to fix a bug.
3. **Make new feature**: leave your idea in the issue tracker and discuss with us then send a pull request!

## License

This project is licensed under the GNU Lesser General Public License version 3 - see the [LICENSE](LICENSE) file for details.

Some packages of the project are under other licenses, please see `README.md` for such packages.
