# Azure Virtual Machine Tags Plugin

This is a plugin for checking Azure Virtual Machine Tags against policies. Policies are evaluated (and should be defined) for individual machines in order to identify machines from violations.

The plugin requires a config with the following values to be passed:

```go
{
    SubscriptionId: "" // The Azure subscription ID
    TenantId:       "", // The client's tenant ID
    ClientID:       "", // The client ID
}
```

It also requires the environment variable `AZURE_CLIENT_SECRET` to be set.

## Prerequisites

* GoReleaser https://goreleaser.com/install/

## Building

Once you are ready to serve the plugin, you need to build the binaries which can be used by the agent.

```shell
goreleaser release --snapshot --clean
```

## Usage

You can use this plugin by passing it to the compliiance agent

```shell
agent --plugin=[PATH_TO_YOUR_BINARY]
```

## Releasing

Once you are ready to release your plugin, you need only create a release in Github, and the plugin binaries
will be added as artifacts on the release page