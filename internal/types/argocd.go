package types

type ArgoConfig struct {
	Host      string
	Insecure  bool
	PlainText bool
	GRPCWeb   bool

	Secret ArgoSecret
}

type ArgoSecret struct {
	Token string
}
