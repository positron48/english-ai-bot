# Spanish k3s rollout checklist

## A. GitOps files
1. Create `apps/spanish/base/*` from `apps/english/base/*` with full rename:
   - namespace/resources `english` -> `spanish`
   - postgres service `spanish-postgres`
   - secrets/configmap names `spanish-*`
2. Create `apps/spanish/prod/kustomization.yaml`.
3. Add `../../apps/spanish/prod` to `clusters/prod/kustomization.yaml`.
4. Add `clusters/prod/infra/image-automation/spanish-image.yaml` and include it in `.../kustomization.yaml`.

## B. Server secrets
```bash
kubectl create namespace spanish --dry-run=client -o yaml | kubectl apply -f -

kubectl -n spanish create secret generic spanish-postgres \
  --from-literal=POSTGRES_DB='spanish' \
  --from-literal=POSTGRES_USER='spanish' \
  --from-literal=POSTGRES_PASSWORD='<POSTGRES_PASSWORD>' \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n spanish create secret generic spanish-secrets \
  --from-literal=DATABASE_URL='postgres://spanish:<POSTGRES_PASSWORD>@spanish-postgres:5432/spanish?sslmode=disable' \
  --from-literal=AI_API_KEY='<OPENROUTER_API_KEY>' \
  --from-literal=WEBAPP_JWT_SECRET='<WEBAPP_JWT_SECRET>' \
  --from-literal=TELEGRAM_TOKEN='<TELEGRAM_TOKEN_OPTIONAL>' \
  --from-literal=TTS_API_KEY='<TTS_API_KEY_OPTIONAL>' \
  --dry-run=client -o yaml | kubectl apply -f -
```

## C. Optional pull secret
```bash
kubectl -n spanish create secret docker-registry ghcr-creds \
  --docker-server=ghcr.io \
  --docker-username='<GITHUB_USERNAME>' \
  --docker-password='<PAT_WITH_read:packages>' \
  --docker-email='<EMAIL>'
```

## D. Flux reconcile
```bash
flux reconcile image repository spanish -n flux-system
flux reconcile image policy spanish -n flux-system
flux reconcile image update flux-system -n flux-system
flux reconcile kustomization flux-system -n flux-system --with-source
```

## E. Runtime checks
```bash
kubectl get deploy -n spanish
kubectl get pods -n spanish
kubectl get ingress -n spanish
kubectl logs -n spanish deploy/spanish --tail=200
kubectl logs -n spanish deploy/spanish-postgres --tail=200
```

## F. Backup/observability additions
1. `apps/k3s-backup/base/configmap.yaml`: add spanish postgres + spanish tts dump blocks.
2. `apps/k3s-backup/base/cronjob.yaml`: add `SPANISH_NAMESPACE=spanish`.
3. `clusters/prod/infra/observability/alloy-config.yaml`: include `spanish` in namespace regex.
4. `clusters/prod/infra/observability/grafana.yaml`: add `Postgres Spanish` datasource via `grafana-db-datasources` secret.

## G. Rollback
```bash
kubectl rollout history deploy/spanish -n spanish
kubectl rollout undo deploy/spanish -n spanish
kubectl rollout status deploy/spanish -n spanish
```

If issue is data-related, restore only Spanish DB/PVC from backup; do not touch English namespace.
