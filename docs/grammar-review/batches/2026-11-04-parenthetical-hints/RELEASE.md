# Release 0.12.239 — verified candidate

Release authorized by user on 2026-11-04 after completed independent audit.

- English course: `6d68d185c18f4e6ef70b249838c15cd3030378c7` (`master`).
- Spanish course: `288c22193010fd53c8f2a703b65c59610a3a7d93` (`grammar-review-batches`).
- Scope: 578 bracket-selected corrections, complaint #71 follow-up, previously prepared complaint #73, embedded bundles and audit records.
- Full `make check P=3 COUNT=1`: PASS. 5294 Go tests passed, 0 failed, 51 skipped; coverage 81.4%. Webapp lint/type-check/tests/build, Go lint/dependency verification, reading validators, and Postgres integration tests passed.
- All 121 independently accepted grammar source hashes still match after the full check. Original tracked webapp build artifact restored byte-for-byte.
- Local release commit/tag prepared; publication awaits GitHub authentication. Push both course commits first, then the application commit and tag `0.12.239`. No tag pushed or production import performed.

After image rollout the server operator must run:

```bash
kubectl -n linglow rollout status deploy/linglow --timeout=300s
kubectl -n linglow exec deploy/linglow -- /app/import_learning_content --course-code=all --commit
```

Run [complaint verification SQL](../../../complaints/verify-2026-11-04.sql) on `english-postgres` / `linglow_unified` before resolving complaints #71/#73. Local production kubectl is prohibited by repository AGENTS.md. No production state is inferred from a tag or CI success.
