[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/sensu/sensu-opcua-checks)
![Go Test](https://github.com/sensu/sensu-opcua-checks/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/sensu/sensu-opcua-checks/workflows/goreleaser/badge.svg)


# Sensu OPC-UA Checks

## Table of Contents
- [Overview](#overview)
- [Usage examples](#usage-examples)
  - [Help output](#help-output)
  - [Environment variables](#environment-variables)
  - [Annotations](#annotations)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Check definition](#check-definition)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Overview

The Sensu OPC-UA Checks are  [Sensu Checks][6] that query Node values obtained from OPC-UA endpoints.
Note: These Sensu checks are built using the [native Golang OPC-UA implementation][11] 

## Usage examples

### sensu-upcua-metrics
Help:

```
Sensu metrics for OPC-UA nodes

Usage:
  sensu-opcua-metrics [flags]
  sensu-opcua-metrics [command]

Available Commands:
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -e, --endpoint string         OP-CUA endpoint (default "opc.tcp://localhost:4840")
  -n, --nodes strings           Comma separated list of OP-CUA nodes to read (ex "ns=1;s=NODE_NAME1, ns=1;s=NODE_NAME2")
  -t, --tags strings            Comma separated list of additional metrics tags to add (ex: "first=value, second=value")
      --metrics-format string   Metrics Format (currently only prometheus supported) (default "prometheus")
      --dry-run                 Dry-run, do not communicate with endpoint, implies verbose
  -v, --verbose                 Enable verbose logging
  -h, --help                    help for sensu-opcua-metrics
```

### Environment variables

|Argument               |Environment Variable            |
|-----------------------|--------------------------------|
|--endpoint             |OPCUA_ENDPOINT                  |
|--nodes                |OPCUA_NODES                     |
|--tags                 |OPCUA_TAGS                      |
|--metrics-format       |OPCUA_METRICS_FORMAT            |

### Annotations

All arguments for these checks are tunable on a per entity or check basis based
on annotations. The annotations keyspace for this collection of checks is
`sensu.io/plugins/opcua/config`.

**NOTE**: Due to [check token substituion][10], supplying a template value such
as for `description-template` as a check annotation requires that you place the
desired template as a [golang string literal][11] (enlcosed in backticks)
within another template definition.  This does not apply to entity annotations.

#### Examples

To customize the opcua endpoint for a given entity, you could use the following
sensu-agent configuration snippet:

```yml
# /etc/sensu/agent.yml example
annotations:
  sensu.io/plugins/opcua/config/endpoint: 'opc.tpc://server.example.com:4840'
```

## Configuration

### Asset registration

[Sensu Assets][10] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add sensu/sensu-opcua-checks
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][https://bonsai.sensu.io/assets/sensu/sensu-opcua-checks].

### Check definition

#### sensu-upcua-metrics

```yml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: sensu-opcua-metrics
  namespace: default
spec:
  command: >- 
    sensu-opcua-metrics 
    --endpoint "opc.tpc://server.example.com:4840" 
    --nodes "ns=1;s=NODE_NAME1, ns=1;s=NODE_NAME2"
  output_metric_format: prometheus_text
  output_metric_handlers:
  - metric-storage
  subscriptions:
  - opcua
  runtime_assets:
  - sensu/sensu-opcua-checks

```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the sensu-opcua-checks repository:

```
go build
```

## Additional notes

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[2]: https://github.com/sensu-community/sensu-plugin-sdk
[3]: https://github.com/sensu-plugins/community/blob/master/PLUGIN_STYLEGUIDE.md
[4]: https://github.com/sensu-community/check-plugin-template/blob/master/.github/workflows/release.yml
[5]: https://github.com/sensu-community/check-plugin-template/actions
[6]: https://docs.sensu.io/sensu-go/latest/reference/checks/
[7]: https://github.com/sensu-community/check-plugin-template/blob/master/main.go
[8]: https://bonsai.sensu.io/
[9]: https://github.com/sensu-community/sensu-plugin-tool
[10]: https://docs.sensu.io/sensu-go/latest/reference/assets/
[11]: https://github.com/gopcua/opcua
