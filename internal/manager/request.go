package manager

import (
	"fmt"
	"strings"

	"k8s.io/utils/ptr"
)

type labelsGetter func(labels map[string]string) string

func createMetaGetter(key string) labelsGetter {
	return func(labels map[string]string) string {
		v, ok := labels[key]
		if !ok {
			return ""
		}

		return v
	}
}

type ReleaseRequest struct {
	releaseGetter     labelsGetter
	environmentGetter labelsGetter
	clusterGetter     labelsGetter

	Selector *string
}

func (r *ReleaseRequest) GetReleaseFrom(labels map[string]string) string {
	if r.releaseGetter == nil {
		return ""
	}

	return r.releaseGetter(labels)
}

func (r *ReleaseRequest) GetClusterFrom(labels map[string]string) string {
	if r.clusterGetter == nil {
		return ""
	}

	return r.clusterGetter(labels)
}

func (r *ReleaseRequest) GetEnvironmentFrom(labels map[string]string) string {
	if r.environmentGetter == nil {
		return ""
	}

	return r.environmentGetter(labels)
}

type ReleaseRequestBuilder struct {
	releaseSelectorKey   string
	releaseSelectorValue string

	environmentSelectorKey   string
	environmentSelectorValue string

	clusterSelectorKey   string
	clusterSelectorValue string
}

func NewReleaseRequestBuilder() *ReleaseRequestBuilder {
	return &ReleaseRequestBuilder{
		releaseSelectorKey:     "platform.ardikabs.com/release",
		environmentSelectorKey: "platform.ardikabs.com/environment",
		clusterSelectorKey:     "platform.ardikabs.com/cluster",
	}
}

type ReleaseRequestBuilderOptions struct {
	SelectorKeyForRelease     string
	SelectorKeyForEnvironment string
	SelectorKeyForCluster     string
}

func NewReleaseRequestBuilderWithOptions(opts *ReleaseRequestBuilderOptions) *ReleaseRequestBuilder {
	b := NewReleaseRequestBuilder()
	b.releaseSelectorKey = opts.SelectorKeyForRelease
	b.environmentSelectorKey = opts.SelectorKeyForEnvironment
	b.clusterSelectorKey = opts.SelectorKeyForCluster
	return b
}

func (b *ReleaseRequestBuilder) SetReleaseSelector(value string) *ReleaseRequestBuilder {
	b.releaseSelectorValue = value
	return b
}

func (b *ReleaseRequestBuilder) SetEnvironmentSelector(value string) *ReleaseRequestBuilder {
	b.environmentSelectorValue = value
	return b
}

func (b *ReleaseRequestBuilder) SetClusterSelector(value string) *ReleaseRequestBuilder {
	b.clusterSelectorValue = value
	return b
}

func (b *ReleaseRequestBuilder) Build() *ReleaseRequest {
	req := &ReleaseRequest{
		environmentGetter: createMetaGetter(b.environmentSelectorKey),
		releaseGetter:     createMetaGetter(b.releaseSelectorKey),
		clusterGetter:     createMetaGetter(b.environmentSelectorKey),
	}

	var selectors []string

	if b.releaseSelectorValue != "" {
		selectors = append(selectors, fmt.Sprintf("%s=%s", b.releaseSelectorKey, b.releaseSelectorValue))
	}

	if b.environmentSelectorValue != "" {
		selectors = append(selectors, fmt.Sprintf("%s=%s", b.environmentSelectorKey, b.environmentSelectorValue))
	}

	if b.clusterSelectorValue != "" {
		selectors = append(selectors, fmt.Sprintf("%s=%s", b.clusterSelectorKey, b.clusterSelectorValue))
	}

	if len(selectors) > 0 {
		req.Selector = ptr.To(strings.Join(selectors, ","))
	}

	return req
}
