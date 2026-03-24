# Spanish k3s rollout checklist

## A. GitOps files (в дереве `devops-time-host`)
Готово: `apps/spanish/base/*`, `apps/spanish/prod`, подключение в `clusters/prod/kustomization.yaml`, `clusters/prod/infra/image-automation/spanish-image.yaml`, бэкап/Alloy/Grafana — см. `docs/SPANISH_K3S_ROLLOUT_RUNBOOK.md`.

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

## F. Backup/observability
Сделано в `devops-time-host`: k3s-backup (postgres + TTS для `spanish`), `SPANISH_NAMESPACE` в cronjob, Alloy regex, datasource Postgres Spanish в `grafana.yaml` (на сервере добавить `SPANISH_POSTGRES_*` в секрет `grafana-db-datasources`).

## G. Rollback
```bash
kubectl rollout history deploy/spanish -n spanish
kubectl rollout undo deploy/spanish -n spanish
kubectl rollout status deploy/spanish -n spanish
```

If issue is data-related, restore only Spanish DB/PVC from backup; do not touch English namespace.
