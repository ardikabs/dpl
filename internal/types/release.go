package types

type Release struct {
	ID string

	Name        string
	Cluster     string
	Environment string
	Image       ImageDefinition

	GitURL      string
	GitPath     string
	GitRevision string
}
