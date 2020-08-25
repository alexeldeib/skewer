package skewer

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	expectedVirtualMachinesCount = 377
	expectedAvailabilityZones    = []string{"1", "2", "3"}
)

// nolint:gocyclo,funlen
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

	chunkedResourceClient, err := newSuccessfulFakeResourceClient(chunk(dataWrapper.Value, 10))
	if err != nil {
		t.Error(err)
	}

	resourceProviderClient, err := newSuccessfulFakeResourceProviderClient([][]compute.ResourceSku{
		dataWrapper.Value,
	})
	if err != nil {
		t.Error(err)
	}

	chunkedResourceProviderClient, err := newSuccessfulFakeResourceProviderClient(chunk(dataWrapper.Value, 10))
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()

	cases := map[string]struct {
		newCacheFunc NewCacheFunc
	}{
		"resourceClient": {
			newCacheFunc: func(_ context.Context, _ ...CacheOption) (*Cache, error) {
				return NewCache(ctx, WithResourceClient(resourceClient), WithLocation("eastus"))
			},
		},
		"chunkedResourceClient": {
			newCacheFunc: func(_ context.Context, _ ...CacheOption) (*Cache, error) {
				return NewCache(ctx, WithResourceClient(chunkedResourceClient), WithLocation("eastus"))
			},
		},
		"resourceProviderClient": {
			newCacheFunc: func(_ context.Context, _ ...CacheOption) (*Cache, error) {
				return NewCache(ctx, WithResourceProviderClient(resourceProviderClient), WithLocation("eastus"))
			},
		},
		"chunkedResourceProviderClient": {
			newCacheFunc: func(_ context.Context, _ ...CacheOption) (*Cache, error) {
				return NewCache(ctx, WithResourceProviderClient(chunkedResourceProviderClient), WithLocation("eastus"))
			},
		},
		"wrappedClient": {
			newCacheFunc: func(_ context.Context, _ ...CacheOption) (*Cache, error) {
				return NewCache(ctx, WithClient(fakeClient), WithLocation("eastus"))
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			cache, err := tc.newCacheFunc(ctx)
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
					errCapabilityValueNil := &ErrCapabilityValueParse{}
					errCapabilityNotFound := &ErrCapabilityNotFound{}

					sku, found := cache.Get(ctx, "standard_d4s_v3", VirtualMachines)
					if !found {
						t.Errorf("expected to find virtual machine sku standard_d4s_v3")
					}
					if name := sku.GetName(); !strings.EqualFold(name, "standard_d4s_v3") {
						t.Errorf("expected standard_d4s_v3 to have name standard_d4s_v3, got: '%s'", name)
					}
					if resourceType := sku.GetResourceType(); resourceType != VirtualMachines {
						t.Errorf("expected standard_d4s_v3 to have resourceType virtual machine, got: '%s'", resourceType)
					}
					if cpu, err := sku.VCPU(); cpu != 4 || err != nil {
						t.Errorf("expected standard_d4s_v3 to have 4 vCPUs and parse successfully, got value '%d' and error '%s'", cpu, err)
					}
					if memory, err := sku.Memory(); memory != 16 || err != nil {
						t.Errorf("expected standard_d4s_v3 to have 16GB of memory and parse successfully, got value '%f' and error '%s'", memory, err)
					}
					if quantity, err := sku.GetCapabilityIntegerQuantity("ShouldNotBePresent"); quantity != -1 || !errors.As(err, &errCapabilityNotFound) {
						t.Errorf("expected standard_d4s_v3 not to have a non-existent capability, got value '%d' and error '%s'", quantity, err)
					}
					if quantity, err := sku.GetCapabilityIntegerQuantity("PremiumIO"); quantity != -1 || !errors.As(err, &errCapabilityValueNil) {
						t.Errorf("expected standard_d4s_v3 to fail parsing value for boolean premiumIO as int, got value '%d' and error '%s'", quantity, err)
					}
					if !sku.HasZonalCapability(UltraSSDAvailable) {
						t.Errorf("expected standard_d4s_v3 to support ultra ssd")
					}
					if sku.HasZonalCapability("NotExistingCapability") {
						t.Errorf("expected standard_d4s_v3 not to support non-existent capability")
					}
					if !sku.HasCapability(EphemeralOSDisk) {
						t.Errorf("expected standard_d4s_v3 to support ephemeral os")
					}
					if !sku.HasCapability(AcceleratedNetworking) {
						t.Errorf("expected standard_d4s_v3 to support accelerated networking")
					}
					if !sku.HasCapability(EncryptionAtHost) {
						t.Errorf("expected standard_d4s_v3 to support encryption at host")
					}
					if !sku.IsAvailable("eastus") {
						t.Errorf("expected standard_d4s_v3 to be available in eastus")
					}
					if sku.IsRestricted("eastus") {
						t.Errorf("expected standard_d4s_v3 to be unrestricted in eastus")
					}
					if sku.IsAvailable("westus2") {
						t.Errorf("expected standard_d4s_v3 not to be available in westus2")
					}
					if sku.IsRestricted("westus2") {
						t.Errorf("expected standard_d4s_v3 not to be restricted in westus2")
					}
					if isSupported, err := sku.HasCapabilityWithCapacity("MaxResourceVolumeMB", 32768); !isSupported || err != nil {
						t.Errorf("expected standard_d4s_v3 to  fit 32GB temp disk, got '%t', error: %s", isSupported, err)
					}
					if isSupported, err := sku.HasCapabilityWithCapacity("MaxResourceVolumeMB", 32769); isSupported || err != nil {
						t.Errorf("expected standard_d4s_v3 not to fit 32GB  +1 byte temp disk, got '%t', error: %s", isSupported, err)
					}
					hasV1 := !sku.HasCapabilityWithSeparator(HyperVGenerations, "V1")
					hasV2 := !sku.HasCapabilityWithSeparator(HyperVGenerations, "V2")
					if hasV1 || hasV2 {
						t.Errorf("expected standard_d4s_v3 to support hyper-v generation v1 and v2, got v1: '%t' , v2: '%t'", hasV1, hasV2)
					}
				})

				t.Run("standard_d2_v2", func(t *testing.T) {
					errCapabilityValueNil := &ErrCapabilityValueParse{}
					errCapabilityNotFound := &ErrCapabilityNotFound{}

					sku, found := cache.Get(ctx, "Standard_D2_v2", VirtualMachines)
					if !found {
						t.Errorf("expected to find virtual machine sku standard_d2_v2")
					}
					if name := sku.GetName(); !strings.EqualFold(name, "standard_d2_v2") {
						t.Errorf("expected standard_d2_v2 to have name standard_d2_v2, got: '%s'", name)
					}
					if resourceType := sku.GetResourceType(); resourceType != VirtualMachines {
						t.Errorf("expected standard_d2_v2 to have resourceType virtual machine, got: '%s'", resourceType)
					}
					if cpu, err := sku.VCPU(); cpu != 2 || err != nil {
						t.Errorf("expected standard_d2_v2 to have 2 vCPUs and parse successfully, got value '%d' and error '%s'", cpu, err)
					}
					if memory, err := sku.Memory(); memory != 7 || err != nil {
						t.Errorf("expected standard_d2_v2 to have 7GB of memory and parse successfully, got value '%f' and error '%s'", memory, err)
					}
					if quantity, err := sku.GetCapabilityIntegerQuantity("ShouldNotBePresent"); quantity != -1 ||
						!errors.As(err, &errCapabilityNotFound) ||
						err.Error() != "ShouldNotBePresentCapabilityNotFound" {
						t.Errorf("expected standard_d2_v2 not to have a non-existent capability, got value '%d' and error '%s'", quantity, err)
					}
					if quantity, err := sku.GetCapabilityIntegerQuantity("PremiumIO"); quantity != -1 ||
						!errors.As(err, &errCapabilityValueNil) ||
						err.Error() != "PremiumIOCapabilityValueParse: failed to parse string 'False' as int64, error: 'strconv.ParseInt: parsing \"False\": invalid syntax'" { // nolint:lll
						t.Errorf("expected standard_d2_v2 to fail parsing value for boolean premiumIO as int, got value '%d' and error '%s'", quantity, err)
					}
					if sku.HasZonalCapability(UltraSSDAvailable) {
						t.Errorf("expected standard_d2_v2 not to support ultra ssd")
					}
					if sku.HasZonalCapability("NotExistingCapability") {
						t.Errorf("expected standard_d2_v2 not to support non-existent capability")
					}
					if sku.HasCapability(EphemeralOSDisk) {
						t.Errorf("expected standard_d2_v2 not to support ephemeral os")
					}
					if !sku.HasCapability(AcceleratedNetworking) {
						t.Errorf("expected standard_d2_v2 to support accelerated networking")
					}
					if sku.HasCapability(EncryptionAtHost) {
						t.Errorf("expected standard_d2_v2 not to support encryption at host")
					}
					if !sku.IsAvailable("eastus") {
						t.Errorf("expected standard_d2_v2 to be available in eastus")
					}
					if sku.IsRestricted("eastus") {
						t.Errorf("expected standard_d2_v2 to be unrestricted in eastus")
					}
					if sku.IsAvailable("westus2") {
						t.Errorf("expected standard_d2_v2 not to be available in westus2")
					}
					if sku.IsRestricted("westus2") {
						t.Errorf("expected standard_d2_v2 not to be restricted in westus2")
					}
					if isSupported, err := sku.HasCapabilityWithCapacity("MemoryGB", 1000); isSupported || err != nil {
						t.Errorf("expected standard_d2_v2 not to have 1000GB of memory, got '%t', error: %s", isSupported, err)
					}
					hasV1 := !sku.HasCapabilityWithSeparator(HyperVGenerations, "V1")
					hasV2 := sku.HasCapabilityWithSeparator(HyperVGenerations, "V2")
					if hasV1 || hasV2 {
						t.Errorf("expected standard_d2_v2 to support hyper-v generation v1 but not v2, got v1: '%t' , v2: '%t'", hasV1, hasV2)
					}
				})

				t.Run("standard_D13_v2_promo", func(t *testing.T) {
					sku, found := cache.Get(ctx, "standard_D13_v2_promo", VirtualMachines)
					if !found {
						t.Errorf("expected to find virtual machine sku standard_D13_v2_promo")
					}
					if sku.IsAvailable("eastus") {
						t.Errorf("expected standard_D13_v2_promo to be unavailable in eastus")
					}
					if !sku.IsRestricted("eastus") {
						t.Errorf("expected standard_D13_v2_promo to be restricted in eastus")
					}
					if sku.IsAvailable("westus2") {
						t.Errorf("expected standard_D13_v2_promo not to be available in westus2")
					}
					if sku.IsRestricted("westus2") {
						t.Errorf("expected standard_D13_v2_promo not to be restricted in westus2")
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
