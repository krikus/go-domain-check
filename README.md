#go-domain-check

## Description

Util that helps validating domains settings by passing options

## TODO

[] Support multiple --ip and --cname options passed together

### Execution

```
Usage:
  go-domain-check [OPTIONS]

Application Options:
      --cname=         Pass a CNAME entry that domain needs to set as well
      --ip=            Pass an IP range that domain needs to be in (can be in CIDR format)
      --tls            Check if domain has valid certificate
  -v, --verbose        Show more info on the console
  -f, --file=          File path with all domains to be checked.
                       Each domain should be in a new line
  -p, --print-valid    Prints valid domains
  -i, --print-invalid  Prints invalid domains
  -c, --concurrency=   How many checks to perform in the same time (default: 100)
  -e, --exit           Exit with code upon detecting invalid domain

```

### Input sources

The binary can be used with those scenarios:

[x] console pipelines: `cat domains.list | go-domain-check --tls -p`
[x] standard input: `go-domain-check --tls -p` + keyboard magic :)
[x] multiple inline params: `go-domain-check --tls -p domain1.com domain2.com`

#### Examples

Check if the domain has valid certificate installed:

`./go-domain-check --tls -e example.com && echo valid || echo invalid`

Check if the domain has valid CNAME entry

`./go-domain-check --cname example.com. test.example.com && echo valid || echo invalid`

Check if domain is pointing to CIDR address

`./go-domain-check --ip 127.0.0.1/24 --ip 1.2.3.4/24 test.example.com`
