# shipyard monitor

Surveillez vos applications en temps réel avec un tableau de bord interactif.

## Utilisation

```bash
shipyard monitor [app-name] [flags]
```

## Description

La commande `monitor` fournit un tableau de bord en temps réel montrant :

- **État des applications** et santé générale
- **Utilisation CPU et mémoire** des pods
- **Nombre de pods** et disponibilité
- **Alertes actives** et leur sévérité
- **Événements récents** du cluster

## Options

| Flag | Description | Défaut |
|------|-------------|---------|
| `--interval, -i` | Intervalle de rafraîchissement | `5s` |
| `--alerts-only` | Afficher seulement les applications avec des alertes | `false` |
| `--compact` | Mode d'affichage compact | `false` |

## Exemples

### Surveiller toutes les applications
```bash
shipyard monitor
```

### Surveiller une application spécifique
```bash
shipyard monitor my-app
```

### Personnaliser l'intervalle de rafraîchissement
```bash
shipyard monitor --interval 10s
```

### Afficher seulement les applications avec des alertes
```bash
shipyard monitor --alerts-only
```

### Mode compact pour les petits terminaux
```bash
shipyard monitor --compact
```

## Interface du tableau de bord

Le tableau de bord affiche :

```
┌─ Shipyard Monitor ─────────────────────────────────────────────────────┐
│ Last updated: 14:32:15 | Press 'q' to quit | Press 'h' for help        │
├────────────────────────────────────────────────────────────────────────┤
│ APP NAME        │ STATUS   │ CPU    │ MEMORY   │ REPLICAS  │ ALERTS │
├─────────────────┼──────────┼────────┼──────────┼───────────┼────────┤
│ web-app         │ 🟢 healthy │ 45.2%  │ 256MB    │ 3/3       │ 0      │
│ api-service     │ 🟡 warning │ 78.9%  │ 512MB    │ 2/3       │ ⚠️ 2   │
├────────────────────────────────────────────────────────────────────────┤
│ Cluster: ✅ Healthy | Nodes: 3/3 | Total Pods: 25                      │
└────────────────────────────────────────────────────────────────────────┘
```

### Légende des statuts

- **🟢 healthy** : Application en bon état
- **🟡 warning** : Alertes ou dégradation partielle
- **🔴 failed** : Application en échec
- **⚫ not-deployed** : Application non déployée
- **⚪ unknown** : État inconnu

## Navigation

- **Ctrl+C** : Quitter le monitoring
- **q** : Quitter (mode interactif)
- **h** : Afficher l'aide (mode interactif)

## Données affichées

### Métriques d'application
- **CPU** : Pourcentage d'utilisation moyenne
- **Mémoire** : Utilisation en MB
- **Replicas** : Pods prêts/pods désirés
- **Alertes** : Nombre d'alertes actives

### Informations cluster
- **Santé globale** : État du cluster
- **Nœuds** : Nœuds prêts/total
- **Pods** : Nombre total de pods

## Intégration avec d'autres commandes

Le monitoring s'intègre avec :

- [`shipyard metrics`](./metrics.md) : Métriques détaillées
- [`shipyard health`](./health.md) : Vérifications de santé
- [`shipyard alerts`](./alerts.md) : Gestion des alertes
- [`shipyard events`](./events.md) : Événements cluster

## Prérequis

- **Cluster Kubernetes** accessible
- **metrics-server** installé pour les métriques CPU/mémoire
- **Applications déployées** avec Shipyard

## Résolution de problèmes

### Pas de métriques CPU/mémoire
```bash
# Vérifier que metrics-server est installé
kubectl get pods -n kube-system | grep metrics-server

# Installer metrics-server si nécessaire
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

### Erreur de connexion au cluster
```bash
# Vérifier la configuration kubectl
kubectl cluster-info

# Vérifier les permissions
kubectl auth can-i get pods
```

### Applications non trouvées
Les applications doivent être déployées avec le label `managed-by=shipyard` pour être détectées.

## Voir aussi

- [shipyard metrics](./metrics.md) - Métriques détaillées
- [shipyard status](./status.md) - État des déploiements
- [Configuration](../getting-started/configuration.md) - Configuration Shipyard