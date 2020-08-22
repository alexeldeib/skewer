package skewer

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/pkg/errors"
)

// ResourceClient is the required Azure client interface used to popualate skewer's data.
type ResourceClient interface {
	ListComplete(ctx context.Context, filter string) (compute.ResourceSkusResultIterator, error)
}

type client interface {
	List(ctx context.Context, filter string) ([]compute.ResourceSku, error)
}

type wrappingClient struct {
	client ResourceClient
}

func newWrappingClient(client ResourceClient) *wrappingClient {
	return &wrappingClient{client}
}

// List greedily traverses all returned sku pages
func (w *wrappingClient) List(ctx context.Context, filter string) ([]compute.ResourceSku, error) {
	iter, err := w.client.ListComplete(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "could not list resource skus")
	}

	var skus []compute.ResourceSku
	for iter.NotDone() {
		skus = append(skus, iter.Value())
		if err := iter.NextWithContext(ctx); err != nil {
			return skus, errors.Wrap(err, "could not iterate resource skus")
		}
	}

	return skus, nil
}
