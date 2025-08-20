# 🚀 Shipyard CLI

CLI moderne pour déployer des applications sur Kubernetes avec simplicité et interactivité.

## Installation

### Installation automatique avec k3s (Recommandé)

**Linux/macOS:**
```bash
curl -sSL https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.sh | bash
```

**Windows PowerShell:**
```powershell
Invoke-WebRequest -Uri "https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.ps1" -OutFile "install.ps1"; .\install.ps1
```

Cette installation inclut automatiquement :
- ✅ **Shipyard CLI** dernière version
- ✅ **k3s/k3d** (Kubernetes léger)
- ✅ **cert-manager** (certificats SSL automatiques)
- ✅ **Configuration SSL** interactive

### Installation manuelle

**Compilation depuis les sources:**
```bash
cd cli
go build -o shipyard main.go
```

**Téléchargement du binaire:**
```bash
# Remplacez PLATFORM par: linux-amd64, darwin-amd64, windows-amd64.exe, etc.
wget https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/shipyard-PLATFORM
chmod +x shipyard-PLATFORM
mv shipyard-PLATFORM /usr/local/bin/shipyard
```

## Configuration

Créez un fichier `paas.yaml` dans le répertoire de votre application :

```yaml
app:
  name: my-api
  image: ghcr.io/myuser/my-api:latest
  port: 3000

service:
  type: ClusterIP     # ou NodePort pour accès externe
  # externalPort: 30000  # requis pour NodePort

health:
  liveness:
    path: /health
    port: 3000
    initialDelaySeconds: 30
    periodSeconds: 10
  readiness:
    path: /ready
    port: 3000
    initialDelaySeconds: 5
    periodSeconds: 5

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

### Supprimer une application

```bash
./shipyard delete                         # Supprimer l'app courante (paas.yaml)
./shipyard delete my-api                  # Supprimer une app spécifique
./shipyard delete --all                   # Supprimer toutes les applications
./shipyard delete --force                 # Supprimer sans confirmation
```

### Mise à niveau du CLI

```bash
./shipyard upgrade                        # Mettre à jour vers la dernière version
./shipyard upgrade --force               # Forcer la mise à jour
./shipyard upgrade --yes                 # Sans confirmation
```

### Gestion SSL/TLS

```bash
./shipyard ssl install                   # Installer cert-manager pour SSL automatique
```

Cette commande va :
1. Installer cert-manager sur votre cluster Kubernetes
2. Demander votre email pour Let's Encrypt
3. Créer un ClusterIssuer pour les certificats automatiques
4. Configurer HTTPS automatique pour vos domaines

### Commandes interactives

Shipyard offre des modes interactifs pour une gestion simplifiée :

**Gestion des registries :**
```bash
./shipyard registry                       # Mode interactif
```
- Ajouter/supprimer des registries Docker
- Configurer registry par défaut
- Configuration simplifiée (URL, username, token uniquement)

**Gestion des domaines :**
```bash
./shipyard domain                         # Mode interactif
```
- Ajouter/supprimer des domaines avec assistant
- Configuration SSL automatique
- Vérification DNS en temps réel

**Rollback interactif :**
```bash
./shipyard rollback                       # Mode interactif avec historique
```
- Affichage de l'historique des déploiements
- Sélection visuelle de la version cible
- Rollback en un clic vers toute version réussie

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

### Core Features
- ✅ **Génération automatique des manifests K8s**
- ✅ **Gestion des secrets** (base64 encodés)
- ✅ **Ingress partagés par domaine** 
- ✅ **Auto-scaling HPA**
- ✅ **Application directe sur le cluster**
- ✅ **Logs en temps réel**
- ✅ **Statut des applications**

### Versioning & Déploiements
- ✅ **Versioning des déploiements**
- ✅ **Historique complet des images déployées**
- ✅ **Rollback automatique vers version stable**
- ✅ **Labels de traçabilité sur tous les manifests**

### Base de données & Domaines
- ✅ **Base de données SQLite** (versions + domaines)
- ✅ **Gestion centralisée des domaines**
- ✅ **Ingress intelligents par base domain**
- ✅ **Commandes de maintenance DB**

### SSL/TLS & Sécurité
- ✅ **SSL automatique avec cert-manager**
- ✅ **Configuration Let's Encrypt interactive**
- ✅ **Support Traefik (k3s) et nginx-ingress**
- ✅ **Certificats HTTPS automatiques**

### Services & Networking
- ✅ **Configuration de services avancée** (ClusterIP, NodePort)
- ✅ **Health checks configurables** (liveness, readiness)
- ✅ **Support ports externes personnalisés**

### Gestion du cycle de vie
- ✅ **Suppression complète des applications** 
- ✅ **Nettoyage automatique des dossiers vides**
- ✅ **Mise à niveau automatique du CLI**
- ✅ **Installation SSL en un clic**

### Compatibilité
- ✅ **k3s (Traefik) support natif**
- ✅ **nginx-ingress support** 
- ✅ **Multi-platform** (Linux, macOS, Windows)
- ✅ **Installation automatique k3s + cert-manager**