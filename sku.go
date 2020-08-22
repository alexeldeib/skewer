package skewer

import (
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/pkg/errors"
)

// SKU wraps the Azure compute SKUs with richer functionality
type SKU compute.ResourceSku

func wrapResourceSKUs(in []compute.ResourceSku) []SKU {
	out := make([]SKU, len(in))
	for index, value := range in {
		out[index] = SKU(value)
	}
	return out
}

// ResourceType is an enum representing existing resources..
type ResourceType string

const (
	// VirtualMachines is the .
	VirtualMachines ResourceType = "virtualMachines"
	// Disks is a convenience constant to filter resource SKUs to only include disks.
	Disks ResourceType = "disks"
)

// Supported models an enum of possible boolean values for resource support in the Azure API.
type Supported string

const (
	// CapabilitySupported is an enum value for the string "True" returned when a SKU supports a binary capability.
	CapabilitySupported Supported = "True"
	// CapabilityUnupported is an enum value for the string "True" returned when a SKU does not support a binary capability.
	CapabilityUnupported Supported = "False"
)

const (
	// EphemeralOSDisk identifies the capability for ephemeral os support.
	EphemeralOSDisk = "EphemeralOSDiskSupported"
	// AcceleratedNetworking identifies the capability for accelerated networking support.
	AcceleratedNetworking = "AcceleratedNetworkingEnabled"
	//VCPUs identifies the capability for the number of vCPUS.
	VCPUs = "vCPUs"
	// MemoryGB identifies the capability for memory capacity.
	MemoryGB = "MemoryGB"
	// HyperVGenerations identifies the hyper-v generations this vm sku supports.
	HyperVGenerations = "HyperVGenerations"
)

// HasCapability return true for a capability which can be either
// supported or not. Examples include "EphemeralOSDiskSupported",
// "UltraSSDAvavailable" "EncryptionAtHostSupported",
// "AcceleratedNetworkingEnabled", and "RdmaEnabled"
func (s SKU) HasCapability(name string) bool {
	if s.Capabilities != nil {
		for _, capability := range *s.Capabilities {
			if capability.Name != nil && *capability.Name == name {
				if capability.Value != nil && strings.EqualFold(*capability.Value, string(CapabilitySupported)) {
					return true
				}
			}
		}
	}
	return false
}

// HasCapabilityWithSeparator return true for a capability which may be
// exposed as a comma-separated list. We check that the list contains
// the desired substring. An example is "HyperVGenerations" which may be
// "V1,V2"
func (s SKU) HasCapabilityWithSeparator(name string, value string) bool {
	if s.Capabilities != nil {
		for _, capability := range *s.Capabilities {
			if capability.Name != nil && *capability.Name == name {
				if capability.Value != nil && strings.Contains(*capability.Value, string(value)) {
					return true
				}
			}
		}
	}
	return false
}

// HasCapabilityWithCapacity returns true when the provided resource
// exposes a numeric capability and the maximum value exposed by that
// capability exceeds the value requested by the user. Examples include
// "MaxResourceVolumeMB", "OSVhdSizeMB", "vCPUs",
// "MemoryGB","MaxDataDiskCount", "CombinedTempDiskAndCachedIOPS",
// "CombinedTempDiskAndCachedReadBytesPerSecond",
// "CombinedTempDiskAndCachedWriteBytesPerSecond", "UncachedDiskIOPS",
// and "UncachedDiskBytesPerSecond"
func (s SKU) HasCapabilityWithCapacity(name string, value int64) (bool, error) {
	if s.Capabilities != nil {
		for _, capability := range *s.Capabilities {
			if capability.Name != nil && *capability.Name == name {
				if capability.Value != nil {
					intVal, err := strconv.ParseInt(*capability.Value, 10, 64)
					if err != nil {
						return false, errors.Wrapf(err, "failed to parse string '%s' as int64", *capability.Value)
					}
					if intVal >= value {
						return true, nil
					}
				}
				return false, nil
			}
		}
	}
	return false, nil
}

// IsResourceType returns true when the wrapped SKU has the provided
// value as its resource type. This may be used to filter using values
// such as "virtualMachines", "disks", "availabilitySets", "snapshots",
// and "hostGroups/hosts".
func IsResourceType(s SKU, t ResourceType) bool {
	return s.ResourceType != nil && strings.EqualFold(*s.ResourceType, string(t))
}

// GetResourceType returns the name of this resource sku. It normalizes pointers
// to the empty string for comparison purposes. For example,
// "virtualMachines" for a virtual machine.
func (s SKU) GetResourceType() string {
	if s.ResourceType == nil {
		return ""
	}
	return *s.ResourceType
}

// GetName returns the name of this resource sku. It normalizes pointers
// to the empty string for comparison purposes. For example,
// "Standard_D8s_v3" for a virtual machine.
func (s SKU) GetName() string {
	if s.Name == nil {
		return ""
	}
	return *s.Name
}

// GetLocation returns the first found location on this SKU resource.
// Typically only one should be listed (multiple SKU results will be returned for multiple regions).
// We fallback to locationInfo although this appears to be duplicate info.
func (s SKU) GetLocation() string {
	if s.Locations != nil {
		for _, location := range *s.Locations {
			return location
		}
	}

	// TODO(ace): probably should remove
	if s.LocationInfo != nil {
		for _, locationInfo := range *s.LocationInfo {
			if locationInfo.Location != nil {
				return *locationInfo.Location
			}
		}
	}

	return ""
}

// AvailabilityZones returns the list of Availability Zones which have this resource SKU available and unrestricted.
func (s SKU) AvailabilityZones(location string) map[string]bool {
	for _, locationInfo := range *s.LocationInfo {
		if strings.EqualFold(*locationInfo.Location, location) {
			// Use map for easy deletion and iteration
			availableZones := make(map[string]bool)

			// add all zones
			for _, zone := range *locationInfo.Zones {
				availableZones[zone] = true
			}

			if s.Restrictions != nil {
				for _, restriction := range *s.Restrictions {
					// Can't deploy to any zones in this location. We're done.
					if restriction.Type == compute.Location {
						availableZones = nil
						break
					}

					// remove restricted zones
					for _, restrictedZone := range *restriction.RestrictionInfo.Zones {
						delete(availableZones, restrictedZone)
					}
				}
			}

			return availableZones
		}
	}

	return nil
}

// Equal returns true when two skus have the same location, type, and name.
func (s SKU) Equal(other SKU) bool {
	return strings.EqualFold(s.GetResourceType(), other.GetResourceType()) &&
		strings.EqualFold(s.GetName(), other.GetName()) &&
		strings.EqualFold(s.GetLocation(), other.GetLocation())
}
