package types

type ImageDefinition struct {
	Name string
	Tag  string
}

func (i ImageDefinition) String() string {
	return i.Name + ":" + i.Tag
}
