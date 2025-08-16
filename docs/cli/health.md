# shipyard health

Vérifiez l'état de santé de vos applications déployées.

## Utilisation

```bash
shipyard health [app-name] [flags]
```

## Description

La commande `health` vérifie l'état de santé des applications via :

- **Vérifications HTTP** sur les endpoints de santé
- **État des pods** et disponibilité
- **Disponibilité des services** Kubernetes
- **Historique des vérifications** récentes

## Options

| Flag | Description | Défaut |
|------|-------------|---------|
| `--watch, -w` | Surveiller continuellement | `false` |
| `--history, -t` | Afficher l'historique sur une durée | - |

## Exemples

### Vérifier toutes les applications
```bash
shipyard health
```

### Vérifier une application spécifique
```bash
shipyard health my-app
```

### Surveillance continue
```bash
shipyard health --watch
```

### Historique des vérifications
```bash
shipyard health --history 1h
```

## Affichage des résultats

### État actuel

```
🏥 Application Health Status
============================

┌───────────────┬────────────┬──────────┬──────┬────────────┬────────┬────────────┐
│ APP NAME      │ ENDPOINT   │ STATUS   │ CODE │ RESP TIME  │ UPTIME │ LAST CHECK │
├───────────────┼────────────┼──────────┼──────┼────────────┼────────┼────────────┤
│ web-app       │ /health    │ 🟢 healthy │ 200  │ 45ms       │ 99.8%  │ 30s ago    │
│ api-service   │ /health    │ 🔴 unhealthy│ 503  │ 2000ms     │ 95.2%  │ 15s ago    │
│               │            │ └─ Service unavailable            │        │            │
│ worker        │ /ping      │ 🟢 healthy │ 200  │ 12ms       │ 100.0% │ 45s ago    │
└───────────────┴────────────┴──────────┴──────┴────────────┴────────┴────────────┘

💡 Tip: Use --watch to monitor health continuously
```

### Historique des vérifications

```
🏥 Health History (Last 1h0m0s)
===============================

┌────────────────────┬──────────┬──────┬────────────┬────────────┐
│ TIMESTAMP          │ STATUS   │ CODE │ RESP TIME  │ ENDPOINT   │
├────────────────────┼──────────┼──────┼────────────┼────────────┤
│ 14:30:15 Jan 15    │ 🟢 healthy │ 200  │ 50ms       │ /health    │
│ 14:25:15 Jan 15    │ 🟢 healthy │ 200  │ 48ms       │ /health    │
│ 14:20:15 Jan 15    │ 🔴 unhealthy│ 503  │ 2000ms     │ /health    │
│ 14:15:15 Jan 15    │ 🟢 healthy │ 200  │ 52ms       │ /health    │
└────────────────────┴──────────┴──────┴────────────┴────────────┘

📊 Success Rate: 75.0% (3/4 checks successful)
```

## Indicateurs de santé

### États des applications

| Icône | État | Description |
|-------|------|-------------|
| 🟢 | healthy | Application en bon état |
| 🟡 | degraded | Performance dégradée |
| 🔴 | unhealthy | Application en échec |
| ⚪ | unknown | État indéterminé |

### Codes de réponse HTTP

| Code | Signification | Action |
|------|---------------|--------|
| 200-299 | Succès | Application saine |
| 300-399 | Redirection | Vérifier la configuration |
| 400-499 | Erreur client | Problème de configuration |
| 500-599 | Erreur serveur | Application défaillante |

## Configuration des endpoints

### Configuration par défaut
- **Endpoint** : `/health`
- **Méthode** : `GET`
- **Timeout** : `5s`
- **Intervalle** : `30s`

### Configuration personnalisée

Les endpoints de santé sont configurés dans la base de données :

```sql
-- Configuration monitoring pour une app
INSERT INTO monitoring_config (
  app_id, 
  health_check_path, 
  health_check_interval,
  health_check_timeout
) VALUES (
  1, 
  '/api/health', 
  60,  -- 60 secondes
  10   -- 10 secondes de timeout
);
```

## Mode surveillance continue

### Démarrer la surveillance

```bash
shipyard health --watch
```

### Interface de surveillance

```
🔍 Health Status - 14:32:45 (auto-refresh: 10s)

🏥 Application Health Status
============================
[Tableau mis à jour automatiquement]
```

- **Rafraîchissement automatique** : Toutes les 10 secondes
- **Indication temporelle** : Horodatage de la dernière mise à jour
- **Arrêt** : Ctrl+C pour quitter

## Métriques de santé

### Temps de réponse
- **Bon** : < 200ms
- **Acceptable** : 200-1000ms
- **Lent** : 1000-5000ms
- **Critique** : > 5000ms

### Taux de disponibilité
- **Excellent** : > 99.9%
- **Bon** : 99.0-99.9%
- **Acceptable** : 95.0-99.0%
- **Problématique** : < 95.0%

## Intégration avec les alertes

Les vérifications de santé peuvent déclencher des alertes :

```bash
# Voir les alertes liées à la santé
shipyard alerts list --type health

# Configurer les seuils d'alerte
shipyard alerts config my-app
```

## Résolution de problèmes

### Endpoint de santé non trouvé (404)

```bash
# Vérifier que l'endpoint existe
curl http://my-app/health

# Configurer le bon endpoint
shipyard config set my-app health.endpoint /api/v1/health
```

### Timeout de connexion

```bash
# Vérifier la connectivité réseau
kubectl port-forward svc/my-app 8080:80
curl http://localhost:8080/health

# Augmenter le timeout si nécessaire
shipyard config set my-app health.timeout 30s
```

### Service indisponible (503)

```bash
# Vérifier les logs de l'application
shipyard logs my-app

# Vérifier l'état des pods
kubectl get pods -l app=my-app
```

## Bonnes pratiques

### Implémentation d'un endpoint de santé

```go
// Exemple en Go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    // Vérifications internes
    if !isDatabaseConnected() {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "unhealthy",
            "reason": "database disconnected"
        })
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy"
    })
}
```

### Monitoring proactif

```bash
# Script de surveillance automatique
#!/bin/bash
while true; do
    if ! shipyard health my-app --format json | jq -e '.[] | select(.status != "healthy")'; then
        echo "All applications healthy"
    else
        echo "Health issues detected!"
        # Envoyer une notification
    fi
    sleep 300  # Vérifier toutes les 5 minutes
done
```

## Voir aussi

- [shipyard monitor](./monitor.md) - Monitoring temps réel
- [shipyard alerts](./alerts.md) - Gestion des alertes
- [shipyard events](./events.md) - Événements système