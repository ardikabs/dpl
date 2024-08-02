package renderer

func New(profile string) Interface {
	switch profile {
	case "kustomize":
		return &Kustomize{}
	default:
		return nil
	}
}
