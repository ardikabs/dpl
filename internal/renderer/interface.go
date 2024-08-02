package renderer

type Interface interface {
	Render(workdir string, releaseName string, params interface{}, opts ...RenderOption) error
}
