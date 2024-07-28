# Workplan

```bash
dpl exec RELEASE_NAME [flags]

Options:
-i, --image string                              Container image to be deployed for the release. It follows 'IMAGE_NAME[:<IMAGE TAG>]' format
-e, --environment string                        Environment to deploy the release
-c, --cluster string                            Cluster to deploy the release
    --kustomize-ref string                      Kustomization file reference (default "kustomization.yaml")
    --kustomize-image-ref string                Kustomization image reference name (default "img")
    --profile string                            Selected profile for deployment (default "kustomize")
    --selector-for-cluster string               Selector for 'cluster' attribute (default "platform.ardikabs.com/cluster")
    --selector-for-environment string           Selector for 'environment' attribute (default "platform.ardikabs.com/environment")
    --selector-for-release string               Selector for 'release' attribute (default "platform.ardikabs.com/release")
-v, --v int                                     Number for the log level verbosity

Environment Variables:
ARGOCD_AUTH_TOKEN               : is the ArgoCD apiKey for your ArgoCD user to be able to authenticate
ARGOCD_SERVER                   : is the address of the ArgoCD server, but without scheme (http{,s}://)
GIT_SECRET                      : is the Git secrets, with the format of Basic Auth credentials. For example: `username:password`
DPL_SELECTOR_FOR_RELEASE        : is the release selector used to specify the resource on Kubernetes, which current supported provider is ArgoCD. It defaults to platform.ardikabs.com/release.
DPL_SELECTOR_FOR_ENVIRONMENT    : is the environment selector used to specify the resource on Kubernetes, which current supported provider is ArgoCD. It defaults to platform.ardikabs.com/environment.
DPL_SELECTOR_FOR_CLUSTER        : is the cluster selector used to specify the resource on Kubernetes, which current supported provider is ArgoCD. It defaults to platform.ardikabs.com/cluster.
```

## Archived Flags

```bash
--repo GIT_REPOSITORY_URL                   The git repository url
--ref GIT_REF                               The git revision (branch, tag, or hash) to check out. If not specified, this defaults to `HEAD` (of the upstream repos default branch).
--workdir GIT_WORKDIR                       The path relatively from git root used as the context, is used to reference file-related
--runner-port DPL_RUNNER_PORT               It defaults to 10080
--renderer DPL_RENDERER                     It specifies the renderer engine to be used, available options are 'kustomize', and 'helm'.
--helm-repo CHART_REPO                      It is the helm chart repository, it could be URL or OCI.
--helm-chart-name CHART_NAME                It is the helm chart name
--helm-chart-version CHART_VERSION          It is the helm chart version
--helm-chart-values CHART_VALUES            It is in the form of string encoded yaml to be passed into Helm
```
