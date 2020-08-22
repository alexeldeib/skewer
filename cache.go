package skewer

import (
	"context"
	"fmt"
	"strings"
)

// Cache stores a list of known skus, possibly fetched with a provided client
type Cache struct {
	location string
	filter   string
	client   client
	data     []SKU
}

// CacheOption describes functional options to customize the listing behavior of the cache.
type CacheOption func(c *Cache)

// WithLocation is a functional option to filter skus by location
func WithLocation(location string) CacheOption {
	return func(c *Cache) {
		c.location = location
		c.filter = fmt.Sprintf("location eq '%s'", location)
	}
}

// NewCacheFunc describes the live cache instantiation signature. Used
// for testing.
type NewCacheFunc func(ctx context.Context, client ResourceClient, opts ...CacheOption) (*Cache, error)

// NewCache instantiates a cache of resource sku data with a ResourceClient
// client, optionally with additional filtering by location. The
// accepted client interface matches the real Azure clients (it returns
// a paginated iterator).
func NewCache(ctx context.Context, client ResourceClient, opts ...CacheOption) (*Cache, error) {
	c := &Cache{
		client: newWrappingClient(client),
	}

	for _, optionFn := range opts {
		optionFn(c)
	}

	if err := c.refresh(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

// NewCacheWithWrappedClient instantiates a cache of resource sku data
// with an easier to mock interface than the live azure client,
// optionally with additional filtering by location.
func NewCacheWithWrappedClient(ctx context.Context, client client, opts ...CacheOption) (*Cache, error) {
	c := &Cache{
		client: client,
	}

	for _, optionFn := range opts {
		optionFn(c)
	}

	if err := c.refresh(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

// NewStaticCacheFn returns a function that initializes a cache with data and no ability to refresh. Used for testing.
func NewStaticCacheFn(data []SKU, opts ...CacheOption) NewCacheFunc {
	return func(ctx context.Context, client ResourceClient, opts ...CacheOption) (*Cache, error) {
		return NewStaticCache(data, opts...), nil
	}
}

// NewStaticCache initializes a cache with data and no ability to refresh. Used for testing.
func NewStaticCache(data []SKU, opts ...CacheOption) *Cache {
	c := &Cache{
		data: data,
	}

	for _, optionFn := range opts {
		optionFn(c)
	}

	return c
}

func (c *Cache) refresh(ctx context.Context) error {
	data, err := c.client.List(ctx, c.filter)
	if err != nil {
		return err
	}

	c.data = wrapResourceSKUs(data)

	return nil
}

// Get returns the first matching resource of a given name and type in a location.
func (c *Cache) Get(ctx context.Context, name, resourceType string) (SKU, bool) {
	filtered := Filter(c.data, []FilterFn{
		ResourceTypeFilter(resourceType),
		NameFilter(name),
	}...)

	if len(filtered) < 1 {
		return SKU{}, false
	}

	return filtered[0], true
}

// List returns all resource types for this location.
func (c *Cache) List(ctx context.Context) []SKU {
	return c.data
}

// GetVirtualMachines returns the list of all virtual machines *SKUs in a given azure location.
func (c *Cache) GetVirtualMachines(ctx context.Context) []SKU {
	return Filter(c.data, ResourceTypeFilter(VirtualMachines))
}

// GetVirtualMachineAvailabilityZones returns all virtual machine zones available in a given location.
func (c *Cache) GetVirtualMachineAvailabilityZones(ctx context.Context) []string {
	return c.GetAvailabilityZones(ctx, ResourceTypeFilter(VirtualMachines))
}

// GetVirtualMachineAvailabilityZonesForSize returns all virtual machine zones available in a given location.
func (c *Cache) GetVirtualMachineAvailabilityZonesForSize(ctx context.Context, size string) []string {
	return c.GetAvailabilityZones(ctx, ResourceTypeFilter(VirtualMachines), NameFilter(size))
}

// GetAvailabilityZones returns the list of all availability zones in a given azure location.
func (c *Cache) GetAvailabilityZones(ctx context.Context, filters ...FilterFn) []string {
	allZones := make(map[string]bool)

	Map(c.data, func(s *SKU) SKU {
		if All(s, filters) {
			for zone := range s.AvailabilityZones(c.location) {
				allZones[zone] = true
			}
		}
		return SKU{}
	})

	result := make([]string, 0, len(allZones))
	for zone := range allZones {
		result = append(result, zone)
	}

	return result
}

// Equal compares two caches.
func (c *Cache) Equal(other *Cache) bool {
	if c == nil && other != nil {
		return false
	}
	if c != nil && other == nil {
		return false
	}
	if c != nil && other != nil {
		return c.location == other.location &&
			c.filter == other.filter
	}
	if len(c.data) != len(other.data) {
		return false
	}
	for i := range c.data {
		if c.data[i] != other.data[i] {
			return false
		}
	}
	return true
}

// All returns true if the provided sku meets all provided conditions.
func All(sku *SKU, conditions []FilterFn) bool {
	for _, condition := range conditions {
		if !condition(sku) {
			return false
		}
	}
	return true
}

// Filter returns a new slice containing all values in the slice that
// satisfy all filterFn predicates.
func Filter(skus []SKU, filterFn ...FilterFn) []SKU {
	if skus == nil {
		return nil
	}

	filtered := make([]SKU, 0)
	for i := range skus {
		if All(&skus[i], filterFn) {
			filtered = append(filtered, skus[i])
		}
	}

	return filtered
}

// Map returns a new slice containing the results of applying the
// mapFn to each value in the original slice.
func Map(skus []SKU, fn MapFn) []SKU {
	if skus == nil {
		return nil
	}

	mapped := make([]SKU, 0, len(skus))
	for i := range skus {
		mapped = append(mapped, fn(&skus[i]))
	}

	return mapped
}

// FilterFn is a convenience type for filtering.
type FilterFn func(*SKU) bool

// ResourceTypeFilter produces a filter function for any resource type.
func ResourceTypeFilter(resourceType string) func(*SKU) bool {
	return func(s *SKU) bool {
		return IsResourceType(s, resourceType)
	}
}

// NameFilter produces a filter function for the name of a resource sku.
func NameFilter(name string) func(*SKU) bool {
	return func(s *SKU) bool {
		return strings.EqualFold(s.GetName(), name)
	}
}

// MapFn is a convenience type for mapping.
type MapFn func(*SKU) SKU
