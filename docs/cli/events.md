# shipyard events

Affichez les événements récents des applications et du cluster.

## Utilisation

```bash
shipyard events [app-name] [flags]
```

## Description

La commande `events` affiche les événements récents incluant :

- **Mises à jour de déploiement** et modifications
- **Changements d'état des pods** (création, destruction, redémarrage)
- **Modifications de services** et configurations
- **Conditions d'erreur** et avertissements
- **Messages du système** Kubernetes

## Options

| Flag | Description | Défaut |
|------|-------------|---------|
| `--follow, -f` | Diffuser les événements en temps réel | `false` |
| `--since, -s` | Afficher les événements depuis une durée | `1h` |
| `--type, -t` | Filtrer par type (normal, warning, error) | - |

## Exemples

### Voir tous les événements récents
```bash
shipyard events
```

### Événements pour une application spécifique
```bash
shipyard events my-app
```

### Diffusion en temps réel
```bash
shipyard events --follow
```

### Événements sur une période personnalisée
```bash
shipyard events --since 6h
```

### Filtrer par type d'événement
```bash
shipyard events --type error
```

### Combinaison de filtres
```bash
shipyard events my-app --type warning --since 2h
```

## Affichage des événements

### Vue tableau standard

```
📅 Recent Events (Last 1h0m0s)
==============================

┌────────────────────┬────────────┬────────────┬──────────┬──────────────────┬────────────────────────────────────────┐
│ TIME               │ APP        │ COMPONENT  │ TYPE     │ REASON           │ MESSAGE                                │
├────────────────────┼────────────┼────────────┼──────────┼──────────────────┼────────────────────────────────────────┤
│ 14:20:35 Jan 15    │ web-app    │ deployment │ ✅ normal  │ ScalingReplicaSet│ Scaled up replica set to 3             │
│ 14:15:20 Jan 15    │ api-service│ pod        │ ⚠️ warning│ BackOff (x3)     │ Back-off restarting failed container   │
│ 14:10:15 Jan 15    │ api-service│ deployment │ ✅ normal  │ DeploymentRollout│ Deployment has minimum availability    │
│ 14:05:10 Jan 15    │ worker     │ pod        │ ❌ error   │ FailedMount      │ Unable to attach volumes: timed out    │
│ 13:50:45 Jan 15    │ web-app    │ service    │ ✅ normal  │ ServiceCreated   │ Service created successfully           │
└────────────────────┴────────────┴────────────┴──────────┴──────────────────┴────────────────────────────────────────┘

📊 Summary: 5 total events, 3 normal, 1 warnings, 1 errors
💡 Tip: Use --follow to stream events in real-time
```

### Mode diffusion temps réel

```bash
$ shipyard events --follow
🔍 Streaming events (press Ctrl+C to stop)
Filtering for app: my-app
Filtering for type: warning

[14:32:15] ⚠️ [api-service/pod] BackOff (x4): Back-off restarting failed container api-service
[14:32:45] ✅ [web-app/deployment] ScalingReplicaSet: Scaled up replica set web-app-7d4b9c8f5c to 4
[14:33:12] ❌ [database/pod] FailedMount: Unable to attach or mount volumes
```

## Types d'événements

### Événements normaux (✅ normal)

| Raison | Composant | Description |
|--------|-----------|-------------|
| `ScalingReplicaSet` | Deployment | Mise à l'échelle des replicas |
| `ServiceCreated` | Service | Création d'un service |
| `DeploymentRollout` | Deployment | Progression du déploiement |
| `PodCreated` | Pod | Création d'un pod |
| `PodStarted` | Pod | Démarrage réussi d'un pod |

### Avertissements (⚠️ warning)

| Raison | Composant | Description |
|--------|-----------|-------------|
| `BackOff` | Pod | Redémarrage en backoff |
| `Unhealthy` | Pod | Échec des health checks |
| `HighMemoryUsage` | Pod | Utilisation mémoire élevée |
| `SlowStart` | Pod | Démarrage lent |
| `ImagePullBackOff` | Pod | Problème de téléchargement d'image |

### Erreurs (❌ error)

| Raison | Composant | Description |
|--------|-----------|-------------|
| `FailedMount` | Pod | Échec de montage de volume |
| `PodFailed` | Pod | Échec critique du pod |
| `DeploymentFailed` | Deployment | Échec de déploiement |
| `ServiceUnavailable` | Service | Service indisponible |
| `ImagePullError` | Pod | Erreur de téléchargement d'image |

## Composants surveillés

### Pods
- **Cycle de vie** : Création, démarrage, arrêt
- **Santé** : Health checks, readiness
- **Ressources** : Utilisation CPU/mémoire
- **Volumes** : Montage, erreurs I/O

### Deployments
- **Rollouts** : Déploiements et mises à jour
- **Scaling** : Mise à l'échelle automatique/manuelle
- **Strategy** : Rolling updates, blue/green

### Services
- **Création/Modification** : Changements de configuration
- **Endpoints** : Mise à jour des backends
- **Load Balancing** : Distribution du trafic

### Volumes
- **Persistent Volumes** : Création, attachement
- **Montage** : Succès/échecs de montage
- **Stockage** : Problèmes d'espace

## Filtrage avancé

### Par période de temps

```bash
# Dernières 30 minutes
shipyard events --since 30m

# Dernières 6 heures
shipyard events --since 6h

# Dernier jour
shipyard events --since 24h

# Dernière semaine
shipyard events --since 168h
```

### Par application et composant

```bash
# Événements d'une application
shipyard events web-app

# Événements de pods seulement
shipyard events --type normal | grep pod

# Événements de déploiement
shipyard events | grep deployment
```

### Par sévérité

```bash
# Seulement les erreurs
shipyard events --type error

# Seulement les avertissements
shipyard events --type warning

# Événements normaux
shipyard events --type normal
```

## Analyse des événements

### Diagnostiquer les problèmes

```bash
# Voir les erreurs récentes
shipyard events --type error --since 1h

# Problèmes spécifiques à une app
shipyard events my-app --type error

# Suivre en temps réel les problèmes
shipyard events --follow --type error
```

### Patterns courants

#### Problèmes de démarrage
```bash
# Rechercher les erreurs de montage et d'image
shipyard events --type error | grep -E "(FailedMount|ImagePull)"
```

#### Problèmes de performance
```bash
# Rechercher les redémarrages fréquents
shipyard events --type warning | grep BackOff
```

#### Problèmes de déploiement
```bash
# Voir les événements de déploiement
shipyard events --since 2h | grep -E "(Deployment|Scaling)"
```

## Intégration et automatisation

### Scripts de monitoring

```bash
#!/bin/bash
# Alerter sur les erreurs critiques
ERROR_COUNT=$(shipyard events --type error --since 5m --format json | jq 'length')

if [ $ERROR_COUNT -gt 0 ]; then
    echo "🚨 $ERROR_COUNT errors in the last 5 minutes"
    shipyard events --type error --since 5m
    # Envoyer notification
fi
```

### Surveillance continue

```bash
# Script de surveillance en arrière-plan
shipyard events --follow --type error | while read event; do
    echo "$(date): $event" >> /var/log/shipyard-errors.log
    # Traitement des erreurs critiques
    if echo "$event" | grep -q "FailedMount\|PodFailed"; then
        # Déclencher une alerte
        alert-manager send "Critical error: $event"
    fi
done
```

### Export pour analyse

```bash
# Export JSON pour outils d'analyse
shipyard events --since 24h --format json > daily-events.json

# Analyse avec jq
cat daily-events.json | jq 'group_by(.type) | map({type: .[0].type, count: length})'

# Export CSV pour tableurs
shipyard events --since 7d --format csv > weekly-events.csv
```

## Corrélation avec d'autres données

### Événements et métriques

```bash
# Voir les événements pendant un pic de CPU
shipyard events --since 1h | grep -E "(cpu|memory|resource)"

# Corréler avec les métriques
shipyard metrics my-app --period 1h
shipyard events my-app --since 1h
```

### Événements et alertes

```bash
# Voir les événements liés aux alertes actives
shipyard alerts list --active
shipyard events --type error --since 2h
```

## Configuration du monitoring d'événements

### Rétention des événements
- **Défaut** : 7 jours
- **Configuration** : Base de données SQLite
- **Nettoyage** : Automatique

### Sources d'événements
- **Kubernetes API** : Événements natifs
- **Shipyard** : Événements applicatifs
- **Monitoring** : Alertes et métriques

## Résolution de problèmes

### Pas d'événements affichés

```bash
# Vérifier la connectivité Kubernetes
kubectl get events

# Vérifier les permissions
kubectl auth can-i get events

# Vérifier la période
shipyard events --since 24h
```

### Trop d'événements

```bash
# Filtrer par type pour réduire le bruit
shipyard events --type error

# Filtrer par application
shipyard events my-critical-app

# Utiliser une période plus courte
shipyard events --since 30m
```

## Voir aussi

- [shipyard monitor](./monitor.md) - Monitoring temps réel
- [shipyard alerts](./alerts.md) - Gestion des alertes
- [shipyard logs](./logs.md) - Logs des applications