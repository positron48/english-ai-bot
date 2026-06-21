# Agent notes for this repo

## How to query the production DB on k3s

There is no postgres pod in the `linglow` namespace, and the `linglow` app
image does not bundle a `psql` client. The actual Postgres instance backing
`linglow_unified` lives in the `english` namespace, as `deploy/english-postgres`
(role `english`, default db `english`); `linglow_unified` is a sibling database
on that same instance.

Working command (read-only queries):

```bash
kubectl -n english exec -it deploy/english-postgres -- psql -U english -d linglow_unified -c "<SQL>"
```

Notes:
- `-U postgres` does NOT work — that role doesn't exist on this instance.
- Do not port-forward, spin up an ad-hoc `pg-client` pod, or look for
  Vault/ExternalSecret-based access — none of that is set up for this project;
  the `kubectl exec` into `english-postgres` above is the working path.
