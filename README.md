# skewer [![codecov](https://codecov.io/gh/alexeldeib/skewer/branch/master/graph/badge.svg)](https://codecov.io/gh/alexeldeib/skewer)

This package wraps the Azure SDK for Go clients to simplify working with Azure's Resource SKU APIs.

## Usage

This package requires an existing, authorized Azure client. Here is a
complete example using the simplest methods.

```go
package main

import (
    "context"
    "fmt"

    "github.com/Azure/go-autorest/autorest/azure/auth"
    "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"

    "github.com/alexeldeib/skewer"
)

func main() {
    authorizer, err := auth.NewAuthorizerFromEnvironment()
    if err != nil {
        fmt.Printf("failed to get authorizer: %s", err)
        os.Exit(1)
    }
	client := compute.NewResourceSkusClient(sub)
    client.Authorizer = authorizer
    // Now we can use the client...
    resourceSkuIterator, err := client.ListComplete(context.Background(), "eastus")
    if err != nil {
        fmt.Printf("failed to list skus: %s", err)
            os.Exit(1)
        }
    // or instantiate a cache for this package!
    cache, err := skewer.NewCache(context.Background(), skewer.WithLocation("eastus"), skewer.WithResourceClient(client))
    if err != nil {
        fmt.Printf("failed to instantiate sku cache: %s", err)
        os.Exit(1)
    }
}
```

Once we have a cache, we can query against its contents:
```go
sku, found := cache.Get(context.Background, "standard_d4s_v3", skewer.VirtualMachines)
if !found {
    return fmt.Errorf("expected to find virtual machine sku standard_d4s_v3")
}
// Check for capabilities
if sku.IsEphemeralOSDiskSupported() {
    fmt.Println("SKU %s supports ephemeral OS disk!", sku.GetName())
}

cpu, err := sku.VCPU()
if err != nil {
    return fmt.Errorf("failed to parse cpu from sku: %s", err)
}
memory, err := sku.Memory()
if err != nil {
    return fmt.Errorf("failed to parse memory from sku: %s", err)
}

fmt.Printf("vm sku %s has %d vCPU cores and %.2fGi of memory", sku.GetName(), cpu, memory)
```
