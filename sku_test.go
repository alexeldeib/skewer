package skewer

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"
)

func Test_SKU_HasCapability(t *testing.T) {}

func Test_SKU_HasCapabilityWithCapacity(t *testing.T) {}

func Test_SKU_IsResourceType(t *testing.T) {
	cases := map[string]struct {
		sku          compute.ResourceSku
		resourceType string
		expect       bool
	}{
		"nil resourceType should not match anything": {
			sku:          compute.ResourceSku{},
			resourceType: "",
		},
		"empty resourceType should match empty string": {
			sku: compute.ResourceSku{
				ResourceType: to.StringPtr(""),
			},
			resourceType: "",
			expect:       true,
		},
		"empty resourceType should not match non-empty string": {
			sku: compute.ResourceSku{
				ResourceType: to.StringPtr(""),
			},
			resourceType: "foo",
		},
		"populated resourceType should match itself": {
			sku: compute.ResourceSku{
				ResourceType: to.StringPtr("foo"),
			},
			resourceType: "foo",
			expect:       true,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tc.expect, IsResourceType(SKU(tc.sku), ResourceType(tc.resourceType))); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_SKU_GetName(t *testing.T) {
	cases := map[string]struct {
		sku    compute.ResourceSku
		expect string
	}{
		"nil name should return empty string": {
			sku:    compute.ResourceSku{},
			expect: "",
		},
		"empty name should return empty string": {
			sku: compute.ResourceSku{
				Name: to.StringPtr(""),
			},
			expect: "",
		},
		"populated name should return correctly": {
			sku: compute.ResourceSku{
				Name: to.StringPtr("foo"),
			},
			expect: "foo",
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tc.expect, SKU(tc.sku).GetName()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_SKU_AvailabilityZones(t *testing.T) {}
