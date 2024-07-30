package exec

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ardikabs/dpl/internal/types"
	"github.com/joeshaw/envdecode"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

type parameters struct {
	ReleaseName            string
	Image                  string
	Environment            string
	Cluster                string
	Profile                string `env:"DPL_PROFILE,default=kustomize"`
	SelectorForRelease     string `env:"DPL_SELECTOR_FOR_RELEASE,default=platform.ardikabs.com/release"`
	SelectorForEnvironment string `env:"DPL_SELECTOR_FOR_ENVIRONMENT,default=platform.ardikabs.com/environment"`
	SelectorForCluster     string `env:"DPL_SELECTOR_FOR_CLUSTER,default=platform.ardikabs.com/cluster"`
	KustomizationFileRef   string `env:"KUSTOMIZE_FILE_REF,default=kustomization.yaml"`
	KustomizationImageRef  string `env:"KUSTOMIZE_IMAGE_REF,default=img"`
	ArgoCDAuthToken        string `env:"ARGOCD_AUTH_TOKEN"`
	ArgoCDHost             string `env:"ARGOCD_HOST"`
	GitSecret              string `env:"GIT_SECRET,required"`
	IsTriggerRestart       bool

	gitSecret       types.GitSecret
	imageDefinition types.ImageDefinition
}

func (p *parameters) Attach(flagset *flag.FlagSet) error {
	if err := envdecode.Decode(p); err != nil {
		return err
	}

	flagset.StringVarP(&p.Image, "image", "i", p.Image, "Container image to be deployed for the release")
	flagset.StringVarP(&p.Environment, "environment", "e", p.Environment, "Environment to deploy the release")
	flagset.StringVarP(&p.Cluster, "cluster", "c", p.Cluster, "Cluster to deploy the release")
	flagset.StringVar(&p.Profile, "profile", p.Profile, "Selected profile for deployment")
	flagset.StringVar(&p.KustomizationFileRef, "kustomize-file-ref", p.KustomizationFileRef, "Kustomization file reference")
	flagset.StringVar(&p.KustomizationImageRef, "kustomize-image-ref", p.KustomizationImageRef, "Kustomization image reference name")
	flagset.StringVar(&p.SelectorForRelease, "selector-for-release", p.SelectorForRelease, "Selector for 'release' attribute")
	flagset.StringVar(&p.SelectorForEnvironment, "selector-for-environment", p.SelectorForEnvironment, "Selector for 'environment' attribute")
	flagset.StringVar(&p.SelectorForCluster, "selector-for-cluster", p.SelectorForCluster, "Selector for 'cluster' attribute")
	flagset.BoolVar(&p.IsTriggerRestart, "restart", p.IsTriggerRestart, "Restart the release")

	return nil
}

func (p *parameters) ParseArgs(args []string) error {
	if len(args) != 1 {
		return errors.New("RELEASE_NAME argument is required")
	}

	p.ReleaseName = args[0]
	return nil
}

func (p *parameters) Validate() error {
	if err := p.validateRequiredFlags(); err != nil {
		return err
	}

	if err := p.validateAndSetGitSecret(); err != nil {
		return err
	}

	if err := p.validateAndSetImageDefinition(); err != nil {
		return err
	}

	return nil
}

func (p *parameters) validateRequiredFlags() error {
	if p.Image == "" {
		return errors.New("image is required. Please set --image flag")
	}

	if p.Environment == "" {
		return errors.New("environment is required. Please set --environment flag")
	}

	if p.ArgoCDHost == "" {
		return errors.New("ArgoCD Host is required. Please set ARGOCD_HOST environment variable")
	}

	if p.ArgoCDAuthToken == "" {
		return errors.New("ArgoCD Auth Token is required. Please set ARGOCD_AUTH_TOKEN environment variable")
	}

	if p.GitSecret == "" {
		return errors.New("git secret is required. Please set GIT_SECRET environment variable")
	}

	return nil
}

func (p *parameters) validateAndSetImageDefinition() error {
	parts := strings.Split(p.Image, ":")
	if len(parts) == 1 {
		p.imageDefinition = types.ImageDefinition{Name: parts[0], Tag: "latest"}
		return nil
	}

	if len(parts) == 2 {
		p.imageDefinition = types.ImageDefinition{Name: parts[0], Tag: parts[1]}
		return nil
	}

	return errors.New("invalid image format, it should be in format <image-name>:<tag>")
}

func (p *parameters) validateAndSetGitSecret() error {
	parts := strings.Split(p.GitSecret, ":")
	if len(parts) != 2 {
		return errors.New("invalid git secret format, it should be in format <username:password>")
	}
	p.gitSecret = types.GitSecret{Username: parts[0], Password: parts[1]}
	return nil
}

func (p *parameters) GetGitSecret() types.GitSecret {
	return p.gitSecret
}

func (p *parameters) GetImageDefinition() types.ImageDefinition {
	return p.imageDefinition
}

func markFlagsAsRequired(flagset *flag.FlagSet, flags ...string) error {
	for _, name := range flags {
		if err := cobra.MarkFlagRequired(flagset, name); err != nil {
			return fmt.Errorf("failed to mark '%s' as required flag", name)
		}
	}

	return nil
}
