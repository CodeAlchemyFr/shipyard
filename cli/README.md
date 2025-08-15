# Shipyard CLI

CLI pour déployer des applications sur Kubernetes avec simplicité.

## Installation

```bash
cd cli
go build -o shipyard main.go
```

## Configuration

Créez un fichier `paas.yaml` dans le répertoire de votre application :

```yaml
app:
  name: my-api
  image: ghcr.io/myuser/my-api:latest
  port: 3000

resources:
  cpu: 200m
  memory: 256Mi

scaling:
  min: 1
  max: 5
  target_cpu: 70

env:
  NODE_ENV: production

secrets:
  DATABASE_URL: postgresql://user:password@postgres:5432/mydb

domains:
  - api.mycompany.com
```

## Utilisation

### Déployer une application

```bash
./shipyard deploy
```

Cette commande va :
1. Lire la configuration `paas.yaml`
2. Générer les manifests Kubernetes dans `manifests/`
3. Appliquer les manifests sur le cluster
4. Attendre que le déploiement soit prêt

### Voir le statut

```bash
./shipyard status
```

### Voir les logs

```bash
./shipyard logs my-api --tail
./shipyard logs my-api --since=1h
```

### Voir l'historique des déploiements

```bash
./shipyard releases
./shipyard releases --limit=20
```

### Rollback en cas de problème

```bash
./shipyard rollback                    # Rollback automatique vers dernière version stable
./shipyard rollback v1634567890        # Rollback vers version spécifique
./shipyard rollback v1.2.3            # Rollback vers tag d'image spécifique
```

### Gestion de la base de données

```bash
./shipyard db status                   # Statistiques de la base SQLite
./shipyard db cleanup                  # Nettoyage des anciens déploiements
```

### Gestion des domaines

```bash
./shipyard domain add api.mycompany.com   # Ajouter un domaine
./shipyard domain list                    # Domaines de l'app courante
./shipyard domain list-all                # Tous les domaines (toutes apps)
./shipyard domain remove api.old.com      # Supprimer un domaine
```

## Structure générée

```
manifests/
├── apps/
│   └── my-api/
│       ├── deployment.yaml (avec labels de version)
│       ├── secrets.yaml
│       └── service.yaml
├── shared/
│   └── mycompany.com.yaml (ingress par domaine de base)
└── shipyard.db (base SQLite : versions + domaines)
```

## Fonctionnalités

- ✅ Génération automatique des manifests K8s
- ✅ Gestion des secrets (base64 encodés)
- ✅ Ingress partagés par domaine
- ✅ Auto-scaling HPA
- ✅ SSL automatique avec cert-manager
- ✅ Application directe sur le cluster
- ✅ Logs en temps réel
- ✅ Statut des applications
- ✅ **Versioning des déploiements**
- ✅ **Historique complet des images déployées**
- ✅ **Rollback automatique vers version stable**
- ✅ **Labels de traçabilité sur tous les manifests**
- ✅ **Base de données SQLite (versions + domaines)**
- ✅ **Gestion centralisée des domaines**
- ✅ **Ingress intelligents par base domain**
- ✅ **Commandes de maintenance DB**