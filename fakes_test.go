package skewer

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
)

type dataWrapper struct {
	Value []compute.ResourceSku `json:"value,omitempty"`
}

type fakeClient struct {
	skus []compute.ResourceSku
	err  error
}

func newDataWrapper(path string) (*dataWrapper, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	wrapper := new(dataWrapper)
	if err := json.Unmarshal(data, wrapper); err != nil {
		return nil, err
	}

	return wrapper, nil
}

func (f *fakeClient) List(ctx context.Context, filter string) ([]compute.ResourceSku, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.skus, nil
}

type fakeResourceClient struct {
	res compute.ResourceSkusResultIterator
	err error
}

func (f *fakeResourceClient) ListComplete(ctx context.Context, filter string) (compute.ResourceSkusResultIterator, error) {
	if f.err != nil {
		return compute.ResourceSkusResultIterator{}, f.err
	}
	return f.res, nil
}

func newFailingFakeResourceClient(err error) (*fakeResourceClient, error) {
	iterator, err := newFakeResourceSkusResultIterator(nil)
	if err != nil {
		return nil, err
	}

	return &fakeResourceClient{
		res: iterator,
		err: err,
	}, nil
}

func newSuccessfulFakeResourceClient(skuLists [][]compute.ResourceSku) (*fakeResourceClient, error) {
	iterator, err := newFakeResourceSkusResultIterator(skuLists)
	if err != nil {
		return nil, err
	}

	return &fakeResourceClient{
		res: iterator,
		err: nil,
	}, nil
}

func newFakeResourceSkusResultIterator(skuLists [][]compute.ResourceSku) (compute.ResourceSkusResultIterator, error) {
	pages := newPageList(skuLists)
	pageFn := func(ctx context.Context, result compute.ResourceSkusResult) (compute.ResourceSkusResult, error) {
		return pages.next(ctx, result)
	}
	newPage := compute.NewResourceSkusResultPage(pageFn)
	if err := newPage.NextWithContext(context.Background()); err != nil {
		return compute.ResourceSkusResultIterator{}, err
	}
	return compute.NewResourceSkusResultIterator(newPage), nil
}

func chunk(skus []compute.ResourceSku, count int) [][]compute.ResourceSku {
	divided := [][]compute.ResourceSku{}
	size := (len(skus) + count - 1) / count
	for i := 0; i < len(skus); i += size {
		end := i + size

		if end > len(skus) {
			end = len(skus)
		}

		divided = append(divided, skus[i:end])
	}
	return divided
}

type pageList struct {
	cursor int
	pages  []compute.ResourceSkusResult
}

func newPageList(skuLists [][]compute.ResourceSku) *pageList {
	list := &pageList{}
	for i := 0; i < len(skuLists); i++ {
		list.pages = append(list.pages, compute.ResourceSkusResult{
			Value: &skuLists[i],
		})
	}
	return list
}

func (p *pageList) next(context.Context, compute.ResourceSkusResult) (compute.ResourceSkusResult, error) {
	if p.cursor >= len(p.pages) {
		return compute.ResourceSkusResult{}, nil
	}
	old := p.cursor
	p.cursor++
	return p.pages[old], nil
}
