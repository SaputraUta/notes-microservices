# notes-microservices

A small 3-tier **microservices** app — a Notes CRUD — built to run on a self-managed Kubernetes cluster.

This is the *application*. The *infrastructure* it runs on is [**k8s-baremetal**](https://github.com/SaputraUta/k8s-baremetal): a Kubernetes cluster built from scratch with no managed service.

> Learning project. The point is the full path: write the services → containerize → deploy to a hand-built cluster → GitOps.

---

## Architecture

```
        user
          │
          ▼
      Ingress            "/"  → frontend
   (ingress-nginx)       "/api" → backend
          │
   ┌──────┴───────┐
   ▼              ▼
frontend        backend  ──SQL──▶  Postgres
(static/nginx)  (Go API)          (PVC-backed, persistent)
```

Three services, each **independently built, imaged, and deployed** — microservices living in one monorepo (monorepo ≠ monolith).

## Services

| Service | Tech | Role |
|---------|------|------|
| `frontend` | static HTML/JS on nginx | UI; calls `/api` on the same host |
| `backend` | Go (`net/http` + `pgx`) | REST CRUD API |
| `database` | Postgres | persistence, backed by a PersistentVolumeClaim |

## Structure

```
backend/    Go API + Dockerfile (distroless, static binary)
frontend/   static UI + Dockerfile (nginx)
k8s/        Kubernetes manifests — Deployments, Services, Ingress, PVC
              (applied with kubectl now; managed by ArgoCD / GitOps next)
```

## Deploy

- Images built for `linux/arm64` and pushed to GHCR (private).
- Manifests in `k8s/` deployed to the cluster; the database's data lives on a PVC provisioned by `local-path`.
- Secrets (DB password, registry pull creds) are created out-of-band — never committed.

## Related

- [k8s-baremetal](https://github.com/SaputraUta/k8s-baremetal) — the from-scratch cluster this app is deployed onto.
