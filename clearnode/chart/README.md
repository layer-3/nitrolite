# clearnode

![Version: 1.0.0](https://img.shields.io/badge/Version-1.0.0-informational?style=flat-square)

Clearnode Helm chart

## Prerequisites

- Kubernetes 1.24+
- Helm 3.0+
- For TLS: cert-manager installed in the cluster

## Installing the Chart

To install the chart with the release name `my-release`:
```bash
helm install my-release git+https://github.com/layer-3/nitrolite@clearnode/chart?ref=main
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
| fullnameOverride | string | `""` | Override the full name |
| config.args | list | `["clearnode"]` | List of arguments to pass to the container |
| config.logLevel | string | `"info"` | Log level (info, debug, warn, error) |
| config.database.driver | string | `"sqlite"` | Database driver (sqlite, postgres) |
| config.database.path | string | `"clearnet.db?cache=shared"` | Database path (for sqlite) |
| config.database.host | string | `""` | Database host |
| config.database.port | int | `5432` | Database port |
| config.database.name | string | `"clearnode"` | Database name |
| config.database.user | string | `"changeme"` | Database user |
| config.database.password | string | `"changeme"` | Database password |
| config.database.sslmode | string | `"disable"` | Database SSL mode (disable, require, verify-ca, verify-full) |
| config.gcpSaSecret | string | `""` | Name of the secret containing GCP SA Credentials (Optional) |
| config.extraEnvs | object | `{}` | Additional environment variables as key-value pairs |
| config.secretEnvs | object | `{}` | Additional environment variables to be stored in a secret |
| config.envSecret | string | `""` | Name of the secret containing environment variables |
| config.blockchains | string | `""` | Blockchains configuration |
| config.assets | string | `""` | Assets configuration |
| config.actionGateway | string | `""` | Action Gateway configuration |
| replicaCount | int | `1` | Number of replicas |
| image.repository | string | `"ghcr.io/layer-3/nitrolite/clearnode"` | Docker image repository |
| image.tag | string | `"v1.0.0-rc.0"` | Docker image tag |
| service.http.enabled | bool | `true` | Enable HTTP service |
| service.http.port | int | `7824` | HTTP service port |
| service.http.path | string | `"/ws"` | HTTP service path (used by ingress) |
| metrics.enabled | bool | `false` | Enable Prometheus metrics |
| metrics.podmonitoring.enabled | bool | `false` | Enable PodMonitoring for Managed Prometheus |
| metrics.port | int | `4242` | Metrics port |
| metrics.endpoint | string | `"/metrics"` | Metrics endpoint path |
| metrics.scrapeInterval | string | `"30s"` | Metrics scrape interval |
| probes.liveness.enabled | bool | `false` | Enable liveness probe |
| probes.liveness.type | string | `"tcp"` | Liveness probe type (http, tcp) |
| probes.readiness.enabled | bool | `false` | Enable readiness probe |
| probes.readiness.type | string | `"tcp"` | Readiness probe type (http, tcp) |
| resources.limits | object | `{}` | Resource limits |
| resources.requests | object | `{}` | Resource requests |
| serviceAccount.create | bool | `false` | Create a ServiceAccount resource |
| serviceAccount.annotations | object | `{}` | Annotations to add to the ServiceAccount (e.g. for GKE Workload Identity) |
| autoscaling.enabled | bool | `false` | Enable autoscaling |
| autoscaling.minReplicas | int | `2` | Minimum number of replicas |
| autoscaling.maxReplicas | int | `100` | Maximum number of replicas |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Target CPU utilization |
| autoscaling.targetMemoryUtilizationPercentage | int | `80` | Target memory utilization |
| networking.tlsClusterIssuer | string | `"zerossl-prod"` | TLS cluster issuer |
| networking.externalHostname | string | `"clearnode.example.com"` | External hostname for the gateway |
| networking.gateway.enabled | bool | `true` | Enable API gateway |
| networking.gateway.className | string | `"envoy-gateway"` | Gateway class name |
| networking.gateway.ipAddressName | string | `""` | GKE static IP address name (GKE only) |
| networking.ingress.enabled | bool | `false` | Enable ingress |
| networking.ingress.className | string | `"nginx"` | Ingress class name |
| networking.ingress.annotations | object | `{}` | Ingress annotations |
| networking.ingress.grpc | bool | `false` | Enable GRPC for ingress |
| networking.ingress.tls.enabled | bool | `false` | Enable TLS for ingress |
| imagePullSecret | string | `""` | Image pull secret name |
| ghcrPullDockerConfigJson | string | `""` | Base64-encoded .dockerconfigjson for GHCR pull secret (provided via SOPS-encrypted secrets) |
| nodeSelector | object | `{}` | Node selector |
| tolerations | list | `[]` | Tolerations |
| affinity | object | `{}` | Affinity settings |
| extraLabels | object | `{}` | Additional labels to add to all resources |
| debug.enabled | bool | `false` | Enable debug deployment (idle container for exec debugging) |
| debug.resources | object | see values.yaml | Resource requests/limits for debug container |
| stressTest.enabled | bool | `false` | Enable stress test pods (helm test) |
| stressTest.wsURL | string | `""` | Default WebSocket URL for all pods (defaults to in-cluster service) |
| stressTest.privateKey | string | `""` | Default private key for signing (optional, ephemeral key used if not set) |
| stressTest.connections | int | `10` | Default number of connections per test |
| stressTest.timeout | string | `"10m"` | Default per-test timeout |
| stressTest.maxErrorRate | string | `"0.01"` | Default max error rate threshold (0.01 = 1%) |
| stressTest.pods | list | see values.yaml | List of stress test pods to run |

## Gateway Configuration

By default, the chart creates an API Gateway and configures it to use TLS via cert-manager. To use this feature:

1. Create a cert-manager ClusterIssuer
2. Configure `networking.tlsClusterIssuer` with the issuer name
3. Set `networking.externalHostname` to your domain name

> **Warning**: The Gateway currently does not support configurations with a static IP address. Ensure that your setup uses a dynamic DNS or hostname for proper functionality. Alternatively, you can configure an ingress resource to use a static IP address if required.

## Troubleshooting

### Common Issues

- **Database Connection Issues**: Ensure the database connection URL is correct and the database is accessible from the cluster
- **TLS Certificate Issues**: Check cert-manager logs for problems with certificate issuance
- **Blockchain Connection Issues**: Verify RPC endpoint URLs are correct and accessible

For more detailed debugging, check the application logs:

```bash
kubectl logs -l app=clearnode
```
