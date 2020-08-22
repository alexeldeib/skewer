package skewer

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	expectedVirtualMachinesCount = 377
	expectedAvailabilityZones    = []string{"1", "2", "3"}
)

func Test_Data(t *testing.T) {
	dataWrapper, err := newDataWrapper("./testdata/eastus.json")
	if err != nil {
		t.Error(err)
	}

	fakeClient := &fakeClient{
		skus: dataWrapper.Value,
	}

	resourceClient, err := newSuccessfulFakeResourceClient([][]compute.ResourceSku{
		dataWrapper.Value,
	})
	if err != nil {
		t.Error(err)
	}

	chunkedClient, err := newSuccessfulFakeResourceClient(chunk(dataWrapper.Value, 10))
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()

	cases := map[string]struct {
		newCacheFunc NewCacheFunc
	}{
		"resourceClient": {
			newCacheFunc: func(_ context.Context, _ ResourceClient, _ ...CacheOption) (*Cache, error) {
				return NewCache(ctx, resourceClient, WithLocation("eastus"))
			},
		},
		"wrappedClient": {
			newCacheFunc: func(_ context.Context, _ ResourceClient, _ ...CacheOption) (*Cache, error) {
				return NewCacheWithWrappedClient(ctx, fakeClient, WithLocation("eastus"))
			},
		},
		"chunkedClient": {
			newCacheFunc: func(_ context.Context, _ ResourceClient, _ ...CacheOption) (*Cache, error) {
				return NewCache(ctx, chunkedClient, WithLocation("eastus"))
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cache, err := tc.newCacheFunc(ctx, resourceClient)
			if err != nil {
				t.Error(err)
			}
			t.Run("virtual machines", func(t *testing.T) {
				t.Run("expect 377 virtual machine skus", func(t *testing.T) {
					if len(cache.GetVirtualMachines(ctx)) != expectedVirtualMachinesCount {
						t.Errorf("expected %d virtual machine skus but found %d", expectedVirtualMachinesCount, len(cache.GetVirtualMachines(ctx)))
					}
				})
				t.Run("standard_d4s_v3", func(t *testing.T) {
					sku, found := cache.Get(ctx, "standard_d4s_v3", VirtualMachines)
					if !found {
						t.Errorf("expected to find virtual machine sku standard_d4s_v3")
					}
					if !sku.HasCapability(EphemeralOSDisk) {
						t.Errorf("expected standard_d4s_v3 to support ephemeral os")
					}
					if fits, err := sku.HasCapabilityWithCapacity("OSVhdSizeMB", 1047552); !fits || err != nil {
						t.Errorf("expected standard_d4s_v3 to support ephemeral os, got '%t', error: %s", fits, err)
					}
					hasV1 := !sku.HasCapabilityWithSeparator(HyperVGenerations, "V1")
					hasV2 := !sku.HasCapabilityWithSeparator(HyperVGenerations, "V2")
					if hasV1 && hasV2 {
						t.Errorf("expected standard_d4s_v3 to support hyper-v generation v1 and v2, got v1: '%t' , v2: '%t'", hasV1, hasV2)
					}
				})
			})

			t.Run("availability zones", func(t *testing.T) {
				if diff := cmp.Diff(cache.GetAvailabilityZones(ctx), expectedAvailabilityZones, []cmp.Option{
					cmpopts.EquateEmpty(),
					cmpopts.SortSlices(func(a, b string) bool {
						return a < b
					}),
				}...); diff != "" {
					t.Errorf("expected and actual availability zones mismatch: %s", diff)
				}
			})
		})
	}
}
