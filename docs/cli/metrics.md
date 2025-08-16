# shipyard metrics

Affichez des métriques détaillées pour vos applications déployées.

## Utilisation

```bash
shipyard metrics [app-name] [flags]
```

## Description

La commande `metrics` affiche des métriques détaillées incluant :

- **Utilisation CPU et mémoire** dans le temps
- **Répartition des métriques par pod**
- **Taux de requêtes** et temps de réponse
- **Tendances historiques** et statistiques

## Options

| Flag | Description | Défaut |
|------|-------------|---------|
| `--period, -p` | Période de temps pour les métriques | `1h` |
| `--format, -f` | Format de sortie (table, json, csv) | `table` |

## Exemples

### Métriques pour toutes les applications
```bash
shipyard metrics
```

### Métriques pour une application spécifique
```bash
shipyard metrics my-app
```

### Métriques sur une période personnalisée
```bash
shipyard metrics --period 6h
```

### Export au format JSON
```bash
shipyard metrics --format json > metrics.json
```

### Export au format CSV
```bash
shipyard metrics --format csv > metrics.csv
```

## Formats de sortie

### Format Table (défaut)

```
📊 Application Metrics (Last 1h0m0s)
=====================================

┌───────────────┬──────────┬──────────┬────────────┬────────────┬────────┬──────────┬────────────┐
│ APP NAME      │ AVG CPU% │ MAX CPU% │ AVG MEM(MB)│ MAX MEM(MB)│ PODS   │ REQ/SEC  │ AVG RESP(ms)│
├───────────────┼──────────┼──────────┼────────────┼────────────┼────────┼──────────┼────────────┤
│ web-app       │ 42.5     │ 78.2     │ 256.0      │ 412.8      │ 3      │ 125.5    │ 89.3       │
│ api-service   │ 65.1     │ 89.7     │ 512.0      │ 724.1      │ 2      │ 89.2     │ 156.8      │
└───────────────┴──────────┴──────────┴────────────┴────────────┴────────┴──────────┴────────────┘

💡 Tip: Use 'shipyard monitor' for real-time monitoring
```

### Format JSON

```json
{
  "metrics": [
    {
      "app_name": "web-app",
      "avg_cpu_percent": 42.5,
      "max_cpu_percent": 78.2,
      "avg_memory_mb": 256.0,
      "max_memory_mb": 412.8,
      "pod_count": 3,
      "requests_per_second": 125.5,
      "avg_response_time_ms": 89.3,
      "data_points": 120,
      "period_seconds": 3600
    }
  ]
}
```

### Format CSV

```csv
app_name,avg_cpu_percent,max_cpu_percent,avg_memory_mb,max_memory_mb,pod_count,requests_per_second,avg_response_time_ms,data_points,period_seconds
web-app,42.5,78.2,256.0,412.8,3,125.5,89.3,120,3600
api-service,65.1,89.7,512.0,724.1,2,89.2,156.8,118,3600
```

## Métriques collectées

### CPU
- **Utilisation moyenne** : Pourcentage moyen sur la période
- **Utilisation maximale** : Pic d'utilisation observé
- **Unité** : Pourcentage (%)

### Mémoire
- **Utilisation moyenne** : Mémoire moyenne consommée
- **Utilisation maximale** : Pic de mémoire observé
- **Unité** : Mégaoctets (MB)

### Pods
- **Nombre de pods** : Nombre total de pods déployés
- **Pods prêts** : Nombre de pods en état "Ready"

### Performance (si disponible)
- **Requêtes par seconde** : Taux de requêtes HTTP
- **Temps de réponse moyen** : Latence moyenne en millisecondes

## Périodes supportées

| Format | Description | Exemple |
|--------|-------------|---------|
| `s` | Secondes | `30s`, `300s` |
| `m` | Minutes | `5m`, `30m` |
| `h` | Heures | `1h`, `6h`, `24h` |
| `d` | Jours | `1d`, `7d` |

## Analyse des données

### Identifier les problèmes de performance

```bash
# Métriques sur 24h pour analyser les tendances
shipyard metrics --period 24h

# Focus sur une application problématique
shipyard metrics api-service --period 6h
```

### Export pour analyse externe

```bash
# Export JSON pour outils de visualisation
shipyard metrics --format json --period 1d > daily-metrics.json

# Export CSV pour tableurs
shipyard metrics --format csv --period 1w > weekly-metrics.csv
```

## Automatisation et intégration

### Scripts de monitoring

```bash
#!/bin/bash
# Collecte quotidienne des métriques
DATE=$(date +%Y%m%d)
shipyard metrics --format json --period 24h > "metrics-${DATE}.json"
```

### Intégration CI/CD

```yaml
# GitHub Actions exemple
- name: Collect Metrics
  run: |
    shipyard metrics --format json --period 1h > deployment-metrics.json
    # Analyser les métriques pour validation
```

## Stockage des données

- **Base de données** : SQLite locale
- **Rétention** : Configurable (défaut : 7 jours)
- **Granularité** : Collecte toutes les 30 secondes

## Voir aussi

- [shipyard monitor](./monitor.md) - Monitoring temps réel
- [shipyard health](./health.md) - Vérifications de santé
- [shipyard alerts](./alerts.md) - Gestion des alertes