# shipyard alerts

Gérez les alertes pour vos applications déployées.

## Utilisation

```bash
shipyard alerts [command] [flags]
```

## Commandes disponibles

| Commande | Description |
|----------|-------------|
| `list` | Lister les alertes actives |
| `history` | Afficher l'historique des alertes |
| `resolve` | Résoudre une alerte |
| `config` | Configurer les seuils d'alerte |

## Options globales

| Flag | Description | Défaut |
|------|-------------|---------|
| `--active, -a` | Afficher seulement les alertes actives | `false` |
| `--period, -p` | Période pour l'historique | `24h` |

## Commandes détaillées

### shipyard alerts list

Affiche les alertes actives ou toutes les alertes.

```bash
# Toutes les alertes actives
shipyard alerts list

# Alertes pour une application spécifique
shipyard alerts list my-app

# Seulement les alertes actives
shipyard alerts list --active
```

**Exemple de sortie :**

```
🚨 Application Alerts
=====================

┌────┬────────────┬───────────────┬──────────┬────────┬────────────┬────────────────────────────────────────┐
│ ID │ APP        │ TYPE          │ SEVERITY │ STATUS │ DURATION   │ MESSAGE                                │
├────┼────────────┼───────────────┼──────────┼────────┼────────────┼────────────────────────────────────────┤
│ 1  │ web-app    │ cpu_high      │ ⚠️ warning │ 🟡 active│ 2h         │ CPU usage 85.2% exceeds threshold 80.0%│
│ 2  │ api-service│ memory_high   │ 🔴 critical│ 🟡 active│ 45m        │ Memory usage 612MB exceeds threshold   │
│ 3  │ api-service│ response_time │ ⚠️ warning │ 🟡 active│ 30m        │ Response time 1250ms exceeds 1000ms    │
└────┴────────────┴───────────────┴──────────┴────────┴────────────┴────────────────────────────────────────┘

📊 Summary: 3 total alerts, 3 active (1 critical)
💡 Tip: Use 'shipyard alerts resolve <id>' to resolve alerts
```

### shipyard alerts history

Affiche l'historique des alertes sur une période donnée.

```bash
# Historique des dernières 24h
shipyard alerts history

# Historique sur une période personnalisée
shipyard alerts history --period 7d

# Historique pour une application
shipyard alerts history my-app --period 1d
```

**Exemple de sortie :**

```
🚨 Alert History (Last 24h0m0s)
===============================

📅 2024-01-15 (5 alerts)
──────────────────────────────────────────────────
  14:30 🟡 [web-app] ⚠️ (2h) - CPU usage 85.2% exceeds threshold
  12:15 ✅ [api-service] 🔴 (1.5h) - Memory usage resolved
  10:30 🟡 [worker] ⚠️ (45m) - Pod restart count exceeded
  09:45 ✅ [web-app] ⚠️ (20m) - Response time back to normal
  08:20 🟡 [database] 🔴 (3h) - Disk usage critical

📊 Summary: 5 total alerts, 3 resolved (60.0%), 2 critical
```

### shipyard alerts resolve

Résout manuellement une alerte spécifique.

```bash
# Résoudre l'alerte avec l'ID 123
shipyard alerts resolve 123
```

**Exemple de sortie :**

```
Resolving alert ID 123...
✅ Alert 123 resolved successfully
```

### shipyard alerts config

Affiche et configure les seuils d'alerte pour une application.

```bash
# Voir la configuration des alertes
shipyard alerts config my-app
```

**Exemple de sortie :**

```
⚙️ Alert Configuration for my-app
==================================

Thresholds:
  CPU Usage:       80.0%
  Memory Usage:    85.0%
  Response Time:   1000ms
  Error Rate:      5.0%

Health Checks:
  Enabled:         true
  Interval:        30s

Notifications:
  Enabled:         false
  Channels:        email, slack

💡 Use 'shipyard config edit my-app' to modify these settings
```

## Types d'alertes

### Alertes de ressources

| Type | Description | Seuil par défaut |
|------|-------------|------------------|
| `cpu_high` | Utilisation CPU élevée | 80% |
| `memory_high` | Utilisation mémoire élevée | 85% |
| `disk_high` | Utilisation disque élevée | 90% |

### Alertes de performance

| Type | Description | Seuil par défaut |
|------|-------------|------------------|
| `response_time_high` | Temps de réponse élevé | 1000ms |
| `error_rate_high` | Taux d'erreur élevé | 5% |
| `requests_low` | Trafic anormalement bas | 10 req/min |

### Alertes de disponibilité

| Type | Description | Condition |
|------|-------------|-----------|
| `pod_restart` | Redémarrages fréquents | >3 en 1h |
| `pod_down` | Pod indisponible | >5 minutes |
| `service_down` | Service inaccessible | >2 minutes |

## Niveaux de sévérité

| Niveau | Icône | Description | Action |
|--------|-------|-------------|--------|
| `info` | 🔵 | Information | Surveillance |
| `warning` | ⚠️ | Attention requise | Investigation |
| `critical` | 🔴 | Action immédiate | Intervention urgente |

## États des alertes

| État | Icône | Description |
|------|-------|-------------|
| `active` | 🟡 | Alerte en cours |
| `resolved` | ✅ | Alerte résolue |
| `suppressed` | 🔇 | Alerte supprimée |

## Configuration des seuils

### Modification des seuils

Les seuils peuvent être configurés dans la base de données :

```sql
-- Configurer les seuils pour une application
UPDATE monitoring_config SET
  cpu_threshold = 70.0,
  memory_threshold = 80.0,
  response_time_threshold = 800
WHERE app_id = 1;
```

### Seuils recommandés

#### Applications web
- **CPU** : 70-80%
- **Mémoire** : 80-85%
- **Temps de réponse** : 500-1000ms

#### APIs
- **CPU** : 60-70%
- **Mémoire** : 75-80%
- **Temps de réponse** : 200-500ms

#### Workers/Jobs
- **CPU** : 80-90%
- **Mémoire** : 85-90%
- **Échecs** : <5%

## Automatisation

### Scripts de gestion des alertes

```bash
#!/bin/bash
# Résoudre automatiquement les alertes anciennes
ALERT_IDS=$(shipyard alerts list --format json | jq -r '.[] | select(.duration > "24h") | .id')

for id in $ALERT_IDS; do
    echo "Auto-resolving old alert $id"
    shipyard alerts resolve $id
done
```

### Intégration CI/CD

```yaml
# GitHub Actions exemple
- name: Check for Critical Alerts
  run: |
    CRITICAL_ALERTS=$(shipyard alerts list --active --format json | jq '[.[] | select(.severity == "critical")] | length')
    if [ $CRITICAL_ALERTS -gt 0 ]; then
      echo "❌ $CRITICAL_ALERTS critical alerts found"
      exit 1
    fi
    echo "✅ No critical alerts"
```

## Notifications (à venir)

La configuration des notifications permettra d'envoyer des alertes vers :

- **Email** : Notifications par email
- **Slack** : Messages dans les canaux Slack
- **Webhooks** : Intégrations personnalisées
- **SMS** : Messages texte urgents

## Export et analyse

### Export JSON

```bash
# Export pour analyse
shipyard alerts history --period 7d --format json > alerts-week.json

# Analyse avec jq
cat alerts-week.json | jq '.[] | select(.severity == "critical")'
```

### Métriques d'alertes

```bash
# Compter les alertes par type
shipyard alerts history --period 1d --format json | jq 'group_by(.type) | map({type: .[0].type, count: length})'

# Temps moyen de résolution
shipyard alerts history --period 7d --format json | jq 'map(select(.resolved_at)) | map(.duration) | add / length'
```

## Résolution de problèmes

### Alertes qui ne se déclenchent pas

```bash
# Vérifier la configuration
shipyard alerts config my-app

# Vérifier les métriques
shipyard metrics my-app --period 30m

# Vérifier les logs du collector
shipyard logs shipyard-monitoring
```

### Trop d'alertes (fatigue d'alerte)

1. **Ajuster les seuils** : Augmenter les valeurs si trop sensibles
2. **Ajouter de l'hystérésis** : Éviter les oscillations
3. **Grouper les alertes** : Regrouper les alertes similaires
4. **Filtrer par sévérité** : Se concentrer sur les critiques

## Voir aussi

- [shipyard monitor](./monitor.md) - Monitoring temps réel
- [shipyard health](./health.md) - Vérifications de santé
- [shipyard metrics](./metrics.md) - Métriques détaillées