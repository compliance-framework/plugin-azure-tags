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