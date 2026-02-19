# Kubernetes Secrets Management Guide

This document describes how to manage secrets for DEparrow deployments across different environments.

## Overview

DEparrow supports multiple secrets management strategies:

| Environment | Strategy | Description |
|-------------|----------|-------------|
| **Dev** | Kustomize secretGenerator | Embedded secrets (development only) |
| **Staging** | External Secrets Operator | Synced from cloud secret stores |
| **Production** | External Secrets Operator | Synced from cloud secret stores with HA |

## Required Secrets

The following secrets are required for a complete DEparrow deployment:

### Application Secrets (`deparrow-secrets`)

| Secret Key | Description | Required |
|------------|-------------|----------|
| `DEPARROW_JWT_SECRET` | JWT token signing key | Yes |
| `DEPARROW_SECRET_KEY` | Application secret key | Yes |
| `POSTGRES_PASSWORD` | PostgreSQL password | Yes |
| `POSTGRES_REPLICATION_PASSWORD` | PostgreSQL replication password | Yes (HA) |
| `REDIS_PASSWORD` | Redis password | Yes |
| `GRAFANA_ADMIN_USER` | Grafana admin username | Yes |
| `GRAFANA_ADMIN_PASSWORD` | Grafana admin password | Yes |
| `ANTHROPIC_API_KEY` | Anthropic API key | No |
| `OPENAI_API_KEY` | OpenAI API key | No |

### Database Secrets (`postgres-credentials`)

| Secret Key | Description |
|------------|-------------|
| `POSTGRES_USER` | PostgreSQL username |
| `POSTGRES_PASSWORD` | PostgreSQL password |
| `POSTGRES_DB` | Database name |

### Cache Secrets (`redis-credentials`)

| Secret Key | Description |
|------------|-------------|
| `REDIS_PASSWORD` | Redis password |

## External Secrets Operator Setup

### Prerequisites

1. Install External Secrets Operator:

```bash
# Using Helm
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets -n external-secrets --create-namespace

# Or using kubectl
kubectl apply -f https://github.com/external-secrets/external-secrets/releases/latest/download/bundle.yaml
```

2. Verify installation:

```bash
kubectl get pods -n external-secrets-system
```

### AWS Secrets Manager Setup

1. Create IAM policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue",
        "secretsmanager:DescribeSecret"
      ],
      "Resource": "arn:aws:secretsmanager:REGION:ACCOUNT_ID:secret:deparrow/*"
    }
  ]
}
```

2. Create IAM role for service account (IRSA):

```bash
eksctl create iamserviceaccount \
  --cluster your-cluster \
  --name deparrow-external-secrets-sa \
  --namespace deparrow \
  --attach-policy-arn arn:aws:iam::ACCOUNT_ID:policy/DeparrowSecretsPolicy \
  --approve
```

3. Create secrets in AWS Secrets Manager:

```bash
# Application secrets
aws secretsmanager create-secret \
  --name deparrow/production/secrets \
  --secret-string '{
    "DEPARROW_JWT_SECRET": "your-secure-jwt-secret",
    "DEPARROW_SECRET_KEY": "your-secure-secret-key",
    "POSTGRES_PASSWORD": "your-db-password",
    "REDIS_PASSWORD": "your-redis-password",
    "GRAFANA_ADMIN_PASSWORD": "your-grafana-password"
  }'

# Database secrets
aws secretsmanager create-secret \
  --name deparrow/production/database \
  --secret-string '{
    "username": "deparrow",
    "password": "your-db-password",
    "database": "deparrow"
  }'

# Redis secrets
aws secretsmanager create-secret \
  --name deparrow/production/redis \
  --secret-string '{"password": "your-redis-password"}'
```

4. Create SecretStore:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: deparrow-secretstore
  namespace: deparrow
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
```

### Azure Key Vault Setup

1. Create Azure Key Vault:

```bash
az keyvault create --name deparrow-kv --resource-group deparrow-rg
```

2. Create secrets:

```bash
az keyvault secret set --vault-name deparrow-kv --name DEPARROW-JWT-SECRET --value "your-jwt-secret"
az keyvault secret set --vault-name deparrow-kv --name DEPARROW-SECRET-KEY --value "your-secret-key"
# ... repeat for other secrets
```

3. Configure Workload Identity:

```bash
# Create managed identity
az identity create --name deparrow-secrets-identity --resource-group deparrow-rg

# Grant access to Key Vault
az keyvault set-policy --name deparrow-kv \
  --object-id <identity-object-id> \
  --secret-permissions get
```

4. Create SecretStore:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: deparrow-secretstore
  namespace: deparrow
spec:
  provider:
    azurekv:
      tenantId: "<azure-tenant-id>"
      vaultUrl: "https://deparrow-kv.vault.azure.net"
      authManagedIdentity:
        identityId: "<managed-identity-client-id>"
```

### HashiCorp Vault Setup

1. Enable Kubernetes auth:

```bash
vault auth enable kubernetes
vault write auth/kubernetes/config \
  kubernetes_host="https://kubernetes.default.svc:443"
```

2. Create policy:

```hcl
path "secret/data/deparrow/*" {
  capabilities = ["read"]
}
```

3. Create role:

```bash
vault write auth/kubernetes/role/deparrow-secrets \
  bound_service_account_names=deparrow-external-secrets-sa \
  bound_service_account_namespaces=deparrow \
  policies=deparrow-secrets-policy
```

4. Create secrets:

```bash
vault kv put secret/deparrow/production \
  DEPARROW_JWT_SECRET="your-jwt-secret" \
  DEPARROW_SECRET_KEY="your-secret-key" \
  POSTGRES_PASSWORD="your-db-password"
```

5. Create SecretStore:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: deparrow-secretstore
  namespace: deparrow
spec:
  provider:
    vault:
      server: "https://vault.example.com"
      path: secret
      version: v2
      auth:
        kubernetes:
          mountPath: kubernetes
          role: deparrow-secrets
```

## Deployment Instructions

### Development

```bash
# Uses embedded secrets - no external secret store required
kubectl apply -k deparrow/k8s/overlays/dev
```

### Staging

1. Configure SecretStore in `deparrow/k8s/overlays/staging/patches/external-secrets.yaml`
2. Create secrets in your external secret store
3. Enable external secrets in kustomization:

```yaml
# Uncomment in overlays/staging/kustomization.yaml
resources:
  - patches/external-secrets.yaml
```

4. Deploy:

```bash
kubectl apply -k deparrow/k8s/overlays/staging
```

### Production

1. Configure SecretStore in `deparrow/k8s/overlays/production/patches/external-secrets.yaml`
2. Create secrets in your external secret store
3. Enable external secrets in kustomization:

```yaml
# Uncomment in overlays/production/kustomization.yaml
resources:
  - patches/external-secrets.yaml
```

4. Deploy:

```bash
kubectl apply -k deparrow/k8s/overlays/production
```

## Secret Rotation

External Secrets Operator automatically refreshes secrets at the configured interval. To force an immediate refresh:

```bash
# Trigger refresh by restarting the ExternalSecret
kubectl annotate externalsecret deparrow-secrets-external \
  force-refresh="$(date +%s)" --overwrite
```

## Security Best Practices

1. **Never commit secrets to version control** - Use external secret stores
2. **Use least privilege** - IAM roles should only access required secrets
3. **Enable audit logging** - Track secret access in your secret store
4. **Rotate secrets regularly** - Set rotation policies in your secret store
5. **Use separate secret stores** - Different stores for dev/staging/production
6. **Enable encryption at rest** - All major secret stores encrypt by default
7. **Network isolation** - Use network policies to restrict pod communication

## Troubleshooting

### ExternalSecret not syncing

```bash
# Check ExternalSecret status
kubectl get externalsecret -n deparrow

# Check events
kubectl describe externalsecret deparrow-secrets-external -n deparrow

# Check External Secrets Operator logs
kubectl logs -n external-secrets-system -l app.kubernetes.io/name=external-secrets
```

### SecretStore authentication failed

```bash
# Verify SecretStore configuration
kubectl describe secretstore deparrow-secretstore -n deparrow

# Check service account annotations
kubectl describe sa deparrow-external-secrets-sa -n deparrow

# For AWS, verify IRSA
kubectl auth can-i get secrets --as=system:serviceaccount:deparrow:deparrow-external-secrets-sa
```

### Secrets not mounted in pods

```bash
# Verify secret exists
kubectl get secret deparrow-secrets -n deparrow

# Check pod events
kubectl describe pod <pod-name> -n deparrow

# Verify secret references in deployment
kubectl get deployment metaos -n deparrow -o jsonpath='{.spec.template.spec.containers[*].envFrom}'
```

## Migration from Embedded Secrets

To migrate from embedded secrets to External Secrets Operator:

1. Create secrets in your external secret store
2. Deploy ExternalSecret resources
3. Verify secrets are synced
4. Remove embedded secrets from kustomization

```bash
# Step 1: Deploy external secrets
kubectl apply -k deparrow/k8s/base -R --dry-run=client | grep -v "kind: Secret" | kubectl apply -f -

# Step 2: Deploy ExternalSecrets
kubectl apply -f deparrow/k8s/base/external-secret.yaml

# Step 3: Verify
kubectl get externalsecret -n deparrow
kubectl get secret deparrow-secrets -n deparrow
```

## File Reference

| File | Purpose |
|------|---------|
| `base/secrets.yaml` | Development placeholder secrets |
| `base/external-secret.yaml` | ExternalSecret definitions and SecretStore examples |
| `overlays/production/patches/external-secrets.yaml` | Production ExternalSecret configuration |
| `overlays/staging/patches/external-secrets.yaml` | Staging ExternalSecret configuration |
