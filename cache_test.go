package skewer

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_Cache_List(t *testing.T)               {}
func Test_Cache_GetVirtualMachines(t *testing.T) {}

// func Test_Cache_GetAvailabilityZones(t *testing.T) {}

func Test_Filter(t *testing.T) {}

func Test_Map(t *testing.T) {}

func Test_Cache_Get(t *testing.T) {
	cases := map[string]struct {
		sku          string
		resourceType ResourceType
		have         []compute.ResourceSku
		found        bool
	}{
		"should find": {
			sku:          "foo",
			resourceType: "bar",
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("other"),
					ResourceType: to.StringPtr("baz"),
				},
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr("bar"),
				},
			},
			found: true,
		},
		"should not find": {
			sku:          "foo",
			resourceType: "bar",
			have: []compute.ResourceSku{
				{
					Name: to.StringPtr("other"),
				},
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cache := &Cache{
				data: wrapResourceSKUs(tc.have),
			}

			val, found := cache.Get(context.Background(), tc.sku, tc.resourceType)
			if tc.found != found {
				t.Errorf("expected %t but got %t when trying to Get resource with name %s and resourceType %s",
					tc.found,
					found,
					tc.sku,
					tc.resourceType,
				)
			} else if found {
				if val.Name == nil {
					t.Fatalf("expected name to be %s, but was nil", tc.sku)
					return
				}
				if *val.Name != tc.sku {
					t.Fatalf("expected name to be %s, but was %s", tc.sku, *val.Name)
				}
				if val.ResourceType == nil {
					t.Fatalf("expected name to be %s, but was nil", tc.sku)
					return
				}
				if *val.ResourceType != string(tc.resourceType) {
					t.Fatalf("expected kind to be %s, but was %s", tc.resourceType, *val.ResourceType)
				}
			}

		})
	}
}

func Test_Cache_GetAvailabilityZones(t *testing.T) {
	cases := map[string]struct {
		have []compute.ResourceSku
		want []string
	}{
		"should find 1 result": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"1"},
						},
					},
				},
			},
			want: []string{"1"},
		},
		"should find 2 results": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"1"},
						},
					},
				},
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"2"},
						},
					},
				},
			},
			want: []string{"1", "2"},
		},
		"should not find due to location mismatch": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"foobar",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("foobar"),
							Zones:    &[]string{"1"},
						},
					},
				},
			},
			want: nil,
		},
		"should not find due to location restriction": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"1"},
						},
					},
					Restrictions: &[]compute.ResourceSkuRestrictions{
						{
							Type:   compute.Location,
							Values: &[]string{"baz"},
						},
					},
				},
			},
			want: nil,
		},
		"should not find due to zone restriction": {
			have: []compute.ResourceSku{
				{
					Name:         to.StringPtr("foo"),
					ResourceType: to.StringPtr(string(VirtualMachines)),
					Locations: &[]string{
						"baz",
					},
					LocationInfo: &[]compute.ResourceSkuLocationInfo{
						{
							Location: to.StringPtr("baz"),
							Zones:    &[]string{"1"},
						},
					},
					Restrictions: &[]compute.ResourceSkuRestrictions{
						{
							Type: compute.Zone,
							RestrictionInfo: &compute.ResourceSkuRestrictionInfo{
								Zones: &[]string{"1"},
							},
						},
					},
				},
			},
			want: nil,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cache := NewStaticCache(wrapResourceSKUs(tc.have), WithLocation("baz"))

			zones := cache.GetAvailabilityZones(context.Background())
			if diff := cmp.Diff(zones, tc.want, []cmp.Option{
				cmpopts.EquateEmpty(),
				cmpopts.SortSlices(func(a, b string) bool {
					return a < b
				}),
			}...); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

// func TestCacheGetZonesWithVMSize(t *testing.T) {
// 	cases := map[string]struct {
// 		have []compute.ResourceSku
// 		want []string
// 	}{
// 		"should find 1 result": {
// 			have: []compute.ResourceSku{
// 				{
// 					Name:         to.StringPtr("foo"),
// 					ResourceType: to.StringPtr(string(VirtualMachines)),
// 					Locations: &[]string{
// 						"baz",
// 					},
// 					LocationInfo: &[]compute.ResourceSkuLocationInfo{
// 						{
// 							Location: to.StringPtr("baz"),
// 							Zones:    &[]string{"1"},
// 						},
// 					},
// 				},
// 			},
// 			want: []string{"1"},
// 		},
// 		"should find 2 results": {
// 			have: []compute.ResourceSku{
// 				{
// 					Name:         to.StringPtr("foo"),
// 					ResourceType: to.StringPtr(string(VirtualMachines)),
// 					Locations: &[]string{
// 						"baz",
// 					},
// 					LocationInfo: &[]compute.ResourceSkuLocationInfo{
// 						{
// 							Location: to.StringPtr("baz"),
// 							Zones:    &[]string{"1", "2"},
// 						},
// 					},
// 				},
// 			},
// 			want: []string{"1", "2"},
// 		},
// 		"should not find due to size mismatch": {
// 			have: []compute.ResourceSku{
// 				{
// 					Name:         to.StringPtr("foobar"),
// 					ResourceType: to.StringPtr(string(VirtualMachines)),
// 					Locations: &[]string{
// 						"baz",
// 					},
// 					LocationInfo: &[]compute.ResourceSkuLocationInfo{
// 						{
// 							Location: to.StringPtr("baz"),
// 							Zones:    &[]string{"1"},
// 						},
// 					},
// 				},
// 			},
// 			want: nil,
// 		},
// 		"should not find due to location mismatch": {
// 			have: []compute.ResourceSku{
// 				{
// 					Name:         to.StringPtr("foo"),
// 					ResourceType: to.StringPtr(string(VirtualMachines)),
// 					Locations: &[]string{
// 						"foobar",
// 					},
// 					LocationInfo: &[]compute.ResourceSkuLocationInfo{
// 						{
// 							Location: to.StringPtr("foobar"),
// 							Zones:    &[]string{"1"},
// 						},
// 					},
// 				},
// 			},
// 			want: nil,
// 		},
// 		"should not find due to location restriction": {
// 			have: []compute.ResourceSku{
// 				{
// 					Name:         to.StringPtr("foo"),
// 					ResourceType: to.StringPtr(string(VirtualMachines)),
// 					Locations: &[]string{
// 						"baz",
// 					},
// 					LocationInfo: &[]compute.ResourceSkuLocationInfo{
// 						{
// 							Location: to.StringPtr("baz"),
// 							Zones:    &[]string{"1"},
// 						},
// 					},
// 					Restrictions: &[]compute.ResourceSkuRestrictions{
// 						{
// 							Type:   compute.Location,
// 							Values: &[]string{"baz"},
// 						},
// 					},
// 				},
// 			},
// 			want: nil,
// 		},
// 		"should not find due to zone restriction": {
// 			have: []compute.ResourceSku{
// 				{
// 					Name:         to.StringPtr("foo"),
// 					ResourceType: to.StringPtr(string(VirtualMachines)),
// 					Locations: &[]string{
// 						"baz",
// 					},
// 					LocationInfo: &[]compute.ResourceSkuLocationInfo{
// 						{
// 							Location: to.StringPtr("baz"),
// 							Zones:    &[]string{"1"},
// 						},
// 					},
// 					Restrictions: &[]compute.ResourceSkuRestrictions{
// 						{
// 							Type: compute.Zone,
// 							RestrictionInfo: &compute.ResourceSkuRestrictionInfo{
// 								Zones: &[]string{"1"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			want: nil,
// 		},
// 	}

// 	for name, tc := range cases {
// 		tc := tc
// 		t.Run(name, func(t *testing.T) {
// 			t.Parallel()

// 			cache := &Cache{
// 				data: tc.have,
// 			}

// 			zones, err := cache.GetZonesWithVMSize(context.Background(), "foo", "baz")
// 			if err != nil {
// 				t.Error(err)
// 			}
// 			if diff := cmp.Diff(zones, tc.want, []cmp.Option{cmpopts.EquateEmpty()}...); diff != "" {
// 				t.Fatalf(diff)
// 			}
// 		})
// 	}
// }
