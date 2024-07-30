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

type ListReleases []*Release

func (l ListReleases) GetGitURL() string {
	if len(l) == 0 {
		return ""
	}

	return l[0].GitURL
}

func (l ListReleases) GetGitRevision() string {
	if len(l) == 0 {
		return ""
	}

	return l[0].GitRevision
}
