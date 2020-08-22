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

func newFakeClientFromPath(path string) (*fakeClient, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	wrapper := new(dataWrapper)
	if err := json.Unmarshal(data, wrapper); err != nil {
		return nil, err
	}

	return &fakeClient{
		skus: wrapper.Value,
	}, nil
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

func newFailingFakeResourceClient(err error) *fakeResourceClient {
	return &fakeResourceClient{
		res: newFakeResourceSkusResultIterator(nil),
		err: err,
	}
}

func newSuccessfulFakeResourceClient(skuLists [][]compute.ResourceSku) *fakeResourceClient {
	return &fakeResourceClient{
		res: newFakeResourceSkusResultIterator(skuLists),
		err: nil,
	}
}

type pageList struct {
	cursor int
	pages  []compute.ResourceSkusResult
}

func newPageList(skuLists [][]compute.ResourceSku) *pageList {
	list := &pageList{}
	for i := 0; i+1 < len(skuLists); i++ {
		list.pages = append(list.pages, compute.ResourceSkusResult{
			Value: &skuLists[i],
		})
	}
	return list
}

func (p *pageList) next(context.Context, compute.ResourceSkusResult) (compute.ResourceSkusResult, error) {
	p.cursor++
	if p.cursor >= len(p.pages) {
		return compute.ResourceSkusResult{}, nil
	}
	return p.pages[p.cursor], nil
}

func newFakeResourceSkusResultIterator(skuLists [][]compute.ResourceSku) compute.ResourceSkusResultIterator {
	pages := newPageList(skuLists)
	pageFn := func(ctx context.Context, result compute.ResourceSkusResult) (compute.ResourceSkusResult, error) {
		return pages.next(ctx, result)
	}
	return compute.NewResourceSkusResultIterator(compute.NewResourceSkusResultPage(pageFn))
}

// func newFakeResourceSkusResultPage(skus []compute.ResourceSku) compute.ResourceSkusResultPage {
// 	return compute.NewResourceSkusResultPage(func(context.Context, compute.ResourceSkusResult) (compute.ResourceSkusResult, error) {
// 		return compute.ResourceSkusResult{
// 			Value: &skus,
// 		}, nil
// 	})
// }
//
// func newSinglePageFake(skus []compute.ResourceSku) *fakeResourceClient {
// 	page := newFakeResourceSkusResultPage(skus)
// 	return &fakeResourceClient{
// 		res: compute.NewResourceSkusResultIterator(page),
// 		err: nil,
// 	}
// }

func (f *fakeResourceClient) ListComplete(ctx context.Context, filter string) (compute.ResourceSkusResultIterator, error) {
	if f.err != nil {
		return compute.ResourceSkusResultIterator{}, f.err
	}
	return f.res, nil
}
