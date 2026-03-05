# clearnode

![Version: 1.0.0](https://img.shields.io/badge/Version-1.0.0-informational?style=flat-square)

Clearnode Helm chart

## Prerequisites

- Kubernetes 1.24+
- Helm 3.0+
- For TLS: cert-manager installed in the cluster
- For Secrets Management (optional):
  - [helm-secrets](https://github.com/jkroepke/helm-secrets/wiki) plugin: `helm plugin install https://github.com/jkroepke/helm-secrets --version v4.6.4`
  - [vals](https://github.com/helmfile/vals): `go install github.com/helmfile/vals/cmd/vals@v0.41.0`

## Installing the Chart

To install the chart with the release name `my-release`:
```bash
helm install my-release git+https://github.com/layer-3/clearnode@chart?ref=main
```

The command deploys Clearnode on the Kubernetes cluster with default configuration. The [Parameters](#parameters) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:
```bash
helm delete my-release
```

## Parameters

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity settings |
| autoscaling.enabled | bool | `false` | Enable autoscaling |
| autoscaling.maxReplicas | int | `100` | Maximum number of replicas |
| autoscaling.minReplicas | int | `2` | Minimum number of replicas |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Target CPU utilization |
| autoscaling.targetMemoryUtilizationPercentage | int | `80` | Target memory utilization |
| config.args | list | `["clearnode"]` | List of arguments to pass to the container |
| config.database.driver | string | `"sqlite"` | Database driver (sqlite, postgres) |
| config.database.host | string | `""` | Database host |
| config.database.name | string | `"clearnode"` | Database name |
| config.database.password | string | `"changeme"` | Database password |
| config.database.path | string | `"clearnet.db?cache=shared"` | Database path (for sqlite) |
| config.database.port | int | `5432` | Database port |
| config.database.sslmode | string | `"disable"` | Database SSL mode (disable, require, verify-ca, verify-full) |
| config.database.user | string | `"changeme"` | Database user |
| config.envSecret | string | `""` | Name of the secret containing environment variables |
| config.extraEnvs | object | `{}` | Additional environment variables as key-value pairs |
| config.logLevel | string | `"info"` | Log level (info, debug, warn, error) |
| config.secretEnvs | object | `{}` | Additional environment variables to be stored in a secret |
| extraLabels | object | `{}` | Additional labels to add to all resources |
| fullnameOverride | string | `""` | Override the full name |
| image.repository | string | `"ghcr.io/layer-3/clearnode"` | Docker image repository |
| image.tag | string | `"0.0.1"` | Docker image tag |
| imagePullSecret | string | `""` | Image pull secret name |
| metrics.enabled | bool | `true` | Enable Prometheus metrics |
| metrics.endpoint | string | `"/metrics"` | Metrics endpoint path |
| metrics.podmonitoring.enabled | bool | `false` | Enable PodMonitoring for Managed Prometheus |
| metrics.port | int | `4242` | Metrics port |
| metrics.scrapeInterval | string | `"30s"` | Metrics scrape interval |
| networking.externalHostname | string | `"clearnode.example.com"` | External hostname for the gateway |
| networking.gateway.className | string | `"envoy-gateway"` | Gateway class name |
| networking.gateway.enabled | bool | `true` | Enable API gateway |
| networking.gateway.ipAddressName | string | `""` | GKE static IP address name (GKE only) |
| networking.ingress.annotations | object | `{}` | Ingress annotations |
| networking.ingress.className | string | `"nginx"` | Ingress class name |
| networking.ingress.enabled | bool | `false` | Enable ingress |
| networking.ingress.grpc | bool | `false` | Enable GRPC for ingress |
| networking.ingress.tls.enabled | bool | `false` | Enable TLS for ingress |
| networking.tlsClusterIssuer | string | `"zerossl-prod"` | TLS cluster issuer |
| nodeSelector | object | `{}` | Node selector |
| probes.liveness.enabled | bool | `false` | Enable liveness probe |
| probes.liveness.type | string | `"tcp"` | Liveness probe type (http, tcp) |
| probes.readiness.enabled | bool | `false` | Enable readiness probe |
| probes.readiness.type | string | `"tcp"` | Readiness probe type (http, tcp) |
| replicaCount | int | `1` | Number of replicas |
| resources.limits | object | `{}` | Resource limits |
| resources.requests | object | `{}` | Resource requests |
| service.http.enabled | bool | `true` | Enable HTTP service |
| service.http.path | string | `"/"` | HTTP service path |
| service.http.port | int | `8000` | HTTP service port |
| serviceAccount | string | `""` | Service account name |
| tolerations | list | `[]` | Tolerations |

## Gateway Configuration

By default, the chart creates an API Gateway and configures it to use TLS via cert-manager. To use this feature:

1. Create a cert-manager ClusterIssuer
2. Configure `gateway.tlsClusterIssuer` with the issuer name
3. Set `gateway.externalHostname` to your domain name

> **Warning**: The Gateway currently does not support configurations with a static IP address. Ensure that your setup uses a dynamic DNS or hostname for proper functionality. Alternatively, you can configure an ingress resource to use a static IP address if required.

## Managing Secrets

For managing sensitive values like API keys and credentials, you can use `helm-secrets` with `vals`:

1. Set up the required environment variable:
   ```bash
   export HELM_SECRETS_BACKEND=vals
   ```

2. Create a values file with your secrets (e.g., `secrets.yaml`) and refer to secrets using the vals syntax:
   ```yaml
   apiKey: ref+awssecrets://my-secret/api-key
   database:
     password: ref+vault://secret/data/database?key=password
   ```
  
3. When deploying or upgrading, reference your secrets file with the `secrets://` prefix:
   ```bash
   helm upgrade --install my-release git+https://github.com/layer-3/clearnode@chart?ref=main \
     -f values.yaml \
     -f secrets://secrets.yaml
   ```

The vals tool supports [multiple backends](https://github.com/helmfile/vals/tree/main?tab=readme-ov-file#supported-backends) including:
- AWS Secrets Manager and SSM Parameter Store
- Google Cloud Secret Manager
- HashiCorp Vault
- Azure Key Vault
- And many more

For detailed usage, consult the [helm-secrets documentation](https://github.com/jkroepke/helm-secrets/wiki).

## Troubleshooting

### Common Issues

- **Database Connection Issues**: Ensure the database connection URL is correct and the database is accessible from the cluster
- **TLS Certificate Issues**: Check cert-manager logs for problems with certificate issuance
- **Blockchain Connection Issues**: Verify RPC endpoint URLs are correct and accessible

For more detailed debugging, check the application logs:

```bash
kubectl logs -l app=clearnode
```

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)