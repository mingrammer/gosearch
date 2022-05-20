# GoSearch

Search the Go packages for [pkg.go.dev](https://pkg.go.dev/) via command-line. It supports all search options in [Search Help](https://pkg.go.dev/search-help).

## Installation

```
go get github.com/mingrammer/gosearch
```

## Usage

Basic search.

```
$ gosearch fastjson
```

You can query with multiple search words.

```
$ gosearch logging zero alloc
```

Use `-n` option to see more. (default: 10)

```
$ gosearch -n 20 redis
```

Use `-s` option to search for symbol.

```
$ gosearch -s mux
```

Use `-e` option to search for an exact match.

```
$ gosearch -e go cloud
```

Use `-o` option to combine searches. It will search for each word and combine their results.

```
$ gosearch -o json yaml
```

## Example

Search the mux packages with `gosearch mux`.

```shell
$ gosearch mux
github.com/gorilla/mux (v1.7.3)
├ Package mux implements a request router and dispatcher.
└ Published: Jun 30, 2019 | Imported by: 6513 | License: BSD-3-Clause

k8s.io/apiserver/pkg/server/mux (v0.0.0 (6eed2f5))
├ Package mux contains abstractions for http multiplexing of APIs.
└ Published: 1 day ago | Imported by: 222 | License: Apache-2.0

github.com/containous/mux (v0.0.0 (c33f32e))
├ Package mux implements a request router and dispatcher.
└ Published: Oct 24, 2018 | Imported by: 95 | License: BSD-3-Clause

k8s.io/kubernetes/pkg/genericapiserver/mux (v1.5.8)
├ Package mux contains abstractions for http multiplexing of APIs.
└ Published: Sep 30, 2017 | Imported by: 61 | License: Apache-2.0

k8s.io/kubernetes/pkg/genericapiserver/server/mux (v1.6.0 (alpha.1))
├ Package mux contains abstractions for http multiplexing of APIs.
└ Published: Jan 30, 2017 | Imported by: 48 | License: Apache-2.0

github.com/coreos/etcd/third_party/github.com/gorilla/mux (v0.4.9)
├ Package gorilla/mux implements a request router and dispatcher.
└ Published: Mar 31, 2015 | Imported by: 66 | License: BSD-3-Clause, Apache-2.0

github.com/muxinc/mux-go/examples/common (v0.3.0)
└ Published: Oct 25, 2019 | Imported by: 9 | License: MIT

github.com/muxinc/mux-go (v0.3.0)
└ Published: Oct 25, 2019 | Imported by: 12 | License: MIT

github.com/yinqiwen/gsnova/common/mux (v0.30.0)
└ Published: Oct 29, 2017 | Imported by: 26 | License: BSD-3-Clause

v2ray.com/core/common/mux (v4.19.1+incompatible)
└ Published: Jun  3, 2019 | Imported by: 24 | License: MIT
```
