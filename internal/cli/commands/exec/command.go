package exec

import (
	"os"

	"github.com/ardikabs/dpl/internal/cli/global"
	"github.com/ardikabs/dpl/internal/log"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	params := new(parameters)

	cmd := &cobra.Command{
		Use:     "exec --environment <ENVIRONMENT> --image <IMAGE_NAME[:IMAGE_TAG]> RELEASE_NAME",
		Aliases: []string{"x", "run"},
		Short:   "execute a deployment runner for given release",
		Long: `Execute a deployment runner for given release.

This command deploys the release to the specified environment using the provided image.
By default, it uses a selector to locate the release definition from the platform manager (e.g., ArgoCD),
the selector can be overridden using the '--selector-for-release', '--selector-for-environment', and '--selector-for-cluster' flag,
then modifies the associated repository for the release manifest to automatically use the specified image.

The renderer has a profile that can be selected using the '--profile' flag,
the available profiles are 'kustomize' and 'helm', and by default it uses the 'kustomize' profile.

> Profile "kustomize"
It automatically generates the kustomization file for the release manifest with the specified image provided in this command.

For example, the following command will deploy the release named 'myapp' to the 'staging' environment using the 'ghcr.io/ardikabs/app/myapp:latest' image

$ dpl exec --environment staging --image ghcr.io/ardikabs/app/myapp:b6d7153 myapp

Just after the command is executed, it will try to look the release definition to ArgoCD, as in Application with the following selectors:
- platform.ardikabs.com/release=myapp,platform.ardikabs.com/environment=staging

Afterward, it will caught the repository that contains the release manifest and modified the kustomization file with the specified image,
below are the before and after the kustomization file is modified:

# Before Rendering
cat <<EOF > kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

images:
  - name: main
    newName: ghcr.io/ardikabs/app/myapp
    newTag: dev

resources:
  - deployment.yaml
  - service.yaml
EOF

# After Rendering
cat <<EOF > kustomization.yaml
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1
images:
  # Image 'main' is managed by dpl. DO NOT EDIT.
  # Warning! Direct changes might be overwritten in the next deployment lifecycle.
  - name: main
    newName: ghcr.io/ardikabs/app/myapp
    newTag: b6d7153
resources:
  - deployment.yaml
  - service.yaml
EOF

Finally, it will commit and push the changes to the remote repository,
and trigger a sync to the ArgoCD Application.
`,
		Example: `
# execute a deployment runner for deploying release named myapp
$ dpl exec --environment staging --image ghcr.io/ardikabs/app/myapp:latest myapp`,
	}

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.RunE = runner(params)

	if err := params.Attach(cmd.Flags()); err != nil {
		log.Error(err, "failed to attach command flags")
		os.Exit(1)
	}

	return cmd
}

func runner(params *parameters) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		log.SetLevel(global.GetLogLevel())

		if err := params.ParseArgs(args); err != nil {
			return err
		}

		if err := params.Validate(); err != nil {
			return err
		}

		instance, err := newExecInstance(log.Logger, params)
		if err != nil {
			return err
		}

		return instance.Exec(cmd.Context())
	}
}
