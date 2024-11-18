package azure_utils

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// GetTagsFromVirtualMachines: Given a credential, call the Azure API and page virtual machines.
// Returns a map of VM ID to tags.
func GetVirtualMachineTags(cred *azidentity.ClientSecretCredential, subscriptionId string) (map[string]map[string]interface{}, error) {
	vmClient, err := armcompute.NewVirtualMachinesClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate virtual machine client: %v", err)
	}

	ctx := context.Background()
	vmTags := make(map[string]map[string]interface{})

	pager := vmClient.NewListAllPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get next page of VMs: %v", err)
		}
		tagRes := make(map[string]interface{})
		for _, vm := range page.Value {
			for key, ptr := range vm.Tags {
				if ptr != nil {
					tagRes[key] = *ptr
				}
			}
			vmTags[*vm.ID] = tagRes
		}
	}
	return vmTags, nil
}
