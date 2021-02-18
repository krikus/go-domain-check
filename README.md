#go-domain-check

## Description

Util that helps validating domains settings by passing options

### Execution

```
Usage:
  go-domain-check [OPTIONS]

Application Options:
      --cname=         Pass a CNAME entry that domain needs to set as well
      --tls            Check if domain has valid certificate
  -v, --verbose        Show more info on the console
  -f, --file=          File path with all domains to be checked.
                       Each domain should be in a new line
  -p, --print-valid    Prints valid domains
  -i, --print-invalid  Prints invalid domains
  -e, --exit           Exit with code upon detecting invalid domain

Help Options:
  -h, --help           Show this help message

```

#### Examples

Check if the domain has valid certificate installed:

`./go-domain-check --tls -e example.com && echo valid || echo invalid`

Check if the domain has valid CNAME entry

`./go-domain-check --cname example.com. test.example.com && echo valid || echo invalid`
