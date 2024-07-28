package renderer

type Renderer interface {
	Render(workdir string, releaseName string, params interface{}, opts ...RenderOption) error
}
