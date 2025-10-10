# Installation

https://gist.github.com/bhimsur/b6c575916883ff7712861beacbe1ff0b

```
kubectl -n argocd apply -f argocd/projects/infrastructure.yaml
kubectl -n argocd apply -f argocd/applications/postgres.yaml
kubectl -n argocd apply -f argocd/applications/kafka.yaml
kubectl -n argocd apply -f argocd/applications/iam.yaml
kubectl -n argocd apply -f argocd/applications/mail.yaml
```
