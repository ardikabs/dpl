package manager

import (
	"errors"
	"fmt"
	"strings"
)

type labelsGetter func(labels map[string]string) string

func createLabelGetter(key string) labelsGetter {
	return func(labels map[string]string) string {
		v, ok := labels[key]
		if !ok {
			return ""
		}

		return v
	}
}

type ListReleaseRequest struct {
	releaseGetter     labelsGetter
	environmentGetter labelsGetter
	clusterGetter     labelsGetter
	selectors         []string

	Selector string
}

func (r *ListReleaseRequest) GetReleaseFrom(labels map[string]string) string {
	if r.releaseGetter == nil {
		return ""
	}

	return r.releaseGetter(labels)
}

func (r *ListReleaseRequest) GetClusterFrom(labels map[string]string) string {
	if r.clusterGetter == nil {
		return ""
	}

	return r.clusterGetter(labels)
}

func (r *ListReleaseRequest) GetEnvironmentFrom(labels map[string]string) string {
	if r.environmentGetter == nil {
		return ""
	}

	return r.environmentGetter(labels)
}

type ListReleaseRequestBuilder struct {
	req *ListReleaseRequest
}

func NewListReleaseRequestBuilder() *ListReleaseRequestBuilder {
	return &ListReleaseRequestBuilder{
		req: &ListReleaseRequest{},
	}
}

func (b *ListReleaseRequestBuilder) SetReleaseSelector(key, value string) *ListReleaseRequestBuilder {
	b.req.releaseGetter = createLabelGetter(key)

	if value == "" {
		return b
	}

	b.req.selectors = append(b.req.selectors, key, value)
	return b
}

func (b *ListReleaseRequestBuilder) SetEnvironmentSelector(key, value string) *ListReleaseRequestBuilder {
	b.req.environmentGetter = createLabelGetter(key)

	if value == "" {
		return b
	}

	b.req.selectors = append(b.req.selectors, key, value)
	return b
}

func (b *ListReleaseRequestBuilder) SetClusterSelector(key, value string) *ListReleaseRequestBuilder {
	b.req.clusterGetter = createLabelGetter(key)

	if value == "" {
		return b
	}

	b.req.selectors = append(b.req.selectors, key, value)
	return b
}

func (b *ListReleaseRequestBuilder) Build() (*ListReleaseRequest, error) {
	if len(b.req.selectors)%2 != 0 {
		return nil, fmt.Errorf("%w, selector must be in the form of key-value pairs: %v", errors.New("invalid selectors"), b.req.selectors)
	}

	var s []string
	for i := 0; i < len(b.req.selectors); i += 2 {
		s = append(s, b.req.selectors[i]+"="+b.req.selectors[i+1])
	}

	b.req.Selector = strings.Join(s, ",")

	return b.req, nil
}
