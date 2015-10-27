# What is it?

`go-repro` is a reverse proxy that maps upstream hosts to public ports on the local
machine, transparently rewriting requests and responses on the way to account for
the mapping.

It was created as a tool for scenarios where you have a bunch of
(v)hosts on a development machine that are resolved by local
`/etc/hosts` entries. If you want to test such a setup on devices other
than the development machine (e.g. mobile devices or even mobile emulators on the
same machine), you will have to publish the host mapping to the testing device,
either setting up DNS server or fiddling with the `hosts` file of the target
device. `go-repro` presents a different solution for this problem by mapping the
domains to public ports on the development machine that can be easily accessed from
other devices without worrying about name resolution.


Consider for example two vhosts `foo.bar.dev` and `huppe.hoppe.dev` that are
configured on your development machine. If you want to test your application with
the android emulator, you will have to mess with the hosts setup of the emulated
device or setup a DNS server. With `go-repro`, you can map these domains to local
ports. Doing

    go-repro -mappings '0.0.0.0:8081=http://foo.bar.dev,0.0.0.0:8082=http://huppe.hoppe.dev' -rewrite '.'

 will forward these domains to ports `8081` and `8082`, with `go-repro` listening
 on all interfaces. If there are references to the hosts in requests and / or
 responses, `go-repro` will transparently rewrite these in order to account for the
 mapping. In the android emulator, you can access the web application via `10.0.2.2:8081`
 (which is mapped to the host loopback interface by the emulator).

*WARNING* `go-repro` is developed as a debugging tool for web application development.
Do not try to use it in a production environment!

# Installation

 In order to install `go-repro`, you need a working [go](https://golang.org/)
 installation. Doing

    go get github.com/mayflower/go-repro/cli/go-repro

 will install the binary into your `GOPATH`.

# Usage

## Host mappings

 Host mappings are configured with the `-mappings` option. This option takes a comma
 separated list of mapping entries. These are written as `local_ip:local_port=remote_host`
 where `local_ip:local_port` is the local IP/port on which `go-repro` will listen, while `remote_host`
 identifies the associated upstream host, _including_ the protocol.

 The local IP `0.0.0.0` causes the proxy to listen on all interfaces and is replaced
 with the actual IP targeted by the request (as stored in the HTTP host) during
 request rewriting.

## Rewriting

Rewriting considers all configured mappings.

### Headers

 Headers are transparently translated for all requests handled by the proxy. In
 particular, the following headers are handled:

  * `Location` for redirects
  * `Referer` for all requests
  * `origin` and `access-control-allow-origin` for requests that use the HTML5 CORS spec for cross origin requests.

### Body content

 Only routes that match one of the regular expressions provided via the
 `-rewrite` option (a comma separated list of regexes) are considered for body
 rewriting. There are two rewriters, one of which performs plain text replacements
 of all occurences of the remote host (including the protocol!) within the response
 body. The second rewriter handles responses of MIME type `application/json` by
 decoding the JSON and subsequently replacing all occurences of the remote host within
 the JSON structure.

## SSL

SSL encrypted connections to upstream hosts are supported. The `-allow-insecure`
option can be specified in order to ignore any issues during cerificate validation
(useful for self-signed certificates). The connection between client and proxy is
always unencrypted.

## Compression

`gzip` compression is supported. The proxy tries to compress upstream connections
via `gzip` and will return `gzip` compressed content to the client if the client
requested it with the `accept-encoding` header and the upstream server supplied a
compressed response.

## Logging

All rewrite steps are logged as a series of `x-gopro-log` headers, e.g.

    X-Go-Repro-Log:json rewriter: response rewritten
    X-Go-Repro-Log:rewrote access-control-allow-origin
    X-Go-Repro-Log:rewrote origin
    X-Go-Repro-Log:rewrote referer

for a request in which the proxy rewrote referer and CORS headers as well as the
JSON encoded server response.

You can disable logging by specifying the `-no-logging` option.

# Limitations

 * Body rewriting of non-JSON responses is a dumb text replacement on byte level.
   The rewriter is not encoding aware and does not parse HTML.
 * The body of HTML redirects is not proxied. This is a open
   [bug](https://github.com/golang/go/issues/10069) in the go standard library.
 * Only `gzip` compression is supported
 * q values in the `accept-encoding` header are not considered. In particular,
   `q=0` is not correctly handled.

# License

`go-repro` is released under the conditions of the MIT license.
