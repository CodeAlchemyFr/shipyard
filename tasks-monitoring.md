# Tasks - Surveillance CLI pour Shipyard

## Objectif
Ajouter des capacités de surveillance et monitoring directement dans le CLI Shipyard, permettant aux développeurs de surveiller leurs applications déployées sans quitter leur terminal.

## Architecture de Surveillance

### 1. Stack Technique
- **Métriques** : Intégration avec Prometheus (metrics-server K8s)
- **Logs** : Agrégation via kubectl + parsing intelligent
- **Health Checks** : Monitoring des endpoints de santé
- **Events** : Surveillance des événements Kubernetes
- **Storage** : Extension de la base SQLite pour stocker métriques historiques

### 2. Nouvelles Commandes CLI

#### `shipyard monitor`
**Surveillance temps réel multi-applications**
```bash
shipyard monitor                    # Toutes les apps
shipyard monitor my-app             # App spécifique
shipyard monitor --interval 5s     # Rafraîchissement custom
shipyard monitor --alerts-only     # Seulement les alertes
```

#### `shipyard metrics`
**Consultation des métriques détaillées**
```bash
shipyard metrics my-app                          # Métriques actuelles
shipyard metrics my-app --history 24h           # Historique 24h
shipyard metrics my-app --type cpu,memory       # Métriques spécifiques
shipyard metrics my-app --export csv            # Export données
shipyard metrics --dashboard                    # Vue d'ensemble toutes apps
```

#### `shipyard health`
**Vérification santé applications**
```bash
shipyard health                     # Santé toutes apps
shipyard health my-app              # App spécifique
shipyard health --deep             # Health checks + dependencies
shipyard health --fix              # Auto-diagnostic + suggestions
```

#### `shipyard alerts`
**Gestion des alertes et notifications**
```bash
shipyard alerts list               # Alertes actives
shipyard alerts setup my-app       # Configuration alertes
shipyard alerts history           # Historique alertes
shipyard alerts test              # Test configuration
```

#### `shipyard events`
**Surveillance événements Kubernetes**
```bash
shipyard events                   # Événements récents
shipyard events my-app            # Events app spécifique
shipyard events --follow          # Stream temps réel
shipyard events --errors-only     # Seulement les erreurs
```

## Tâches de Développement

### Phase 1 : Fondations (2-3 semaines)

#### T1.1 - Extension Base de Données
- [ ] Ajouter tables `metrics`, `health_checks`, `alerts`, `events`
- [ ] Système de rétention données (7j/30j configurable)
- [ ] Migration automatique schéma existant
- [ ] Index optimisés pour requêtes temporelles

#### T1.2 - Client Kubernetes Étendu
- [ ] Extension `pkg/k8s/client.go` pour metrics
- [ ] Intégration metrics-server Kubernetes
- [ ] Parser d'événements Kubernetes
- [ ] Collecteur de logs structurés

#### T1.3 - Moteur de Métriques
- [ ] `pkg/monitoring/collector.go` - Collecte métriques
- [ ] `pkg/monitoring/aggregator.go` - Agrégation données
- [ ] `pkg/monitoring/storage.go` - Stockage SQLite
- [ ] Support métriques : CPU, Memory, Network, Disk I/O

### Phase 2 : Commandes Core (2-3 semaines)

#### T2.1 - Commande `shipyard monitor`
- [ ] `cmd/monitor.go` - Interface temps réel
- [ ] Affichage multi-applications en tableau
- [ ] Rafraîchissement automatique configurable
- [ ] Indicateurs visuels (couleurs, symboles)
- [ ] Mode compact vs détaillé

#### T2.2 - Commande `shipyard metrics`
- [ ] `cmd/metrics.go` - Consultation historique
- [ ] Graphiques ASCII pour trends
- [ ] Filtrage par type métrique et période
- [ ] Export formats : CSV, JSON, Prometheus
- [ ] Calculs automatiques : moyennes, percentiles

#### T2.3 - Commande `shipyard health`
- [ ] `cmd/health.go` - Diagnostic santé
- [ ] Tests health checks HTTP/TCP
- [ ] Vérification dépendances (DB, Redis, etc.)
- [ ] Suggestions automatiques de fixes
- [ ] Score de santé global

### Phase 3 : Alerting & Events (1-2 semaines)

#### T3.1 - Système d'Alertes
- [ ] `pkg/alerting/` - Moteur d'alertes
- [ ] Configuration seuils dynamiques
- [ ] Règles d'alertes par application
- [ ] Notifications : email, webhook, Slack
- [ ] Suppression des alertes résolues

#### T3.2 - Surveillance Événements
- [ ] `cmd/events.go` - Stream événements K8s
- [ ] Parsing et catégorisation événements
- [ ] Corrélation events <-> métriques
- [ ] Filtrage intelligent (erreurs, warnings)
- [ ] Export événements critiques

### Phase 4 : Interface & UX (1-2 semaines)

#### T4.1 - Dashboard CLI
- [ ] Interface TUI (Terminal UI) avec `bubbletea`
- [ ] Navigation clavier entre applications
- [ ] Graphiques temps réel en ASCII
- [ ] Panneau split : metrics + logs + events
- [ ] Mode plein écran vs intégré

#### T4.2 - Alertes Visuelles
- [ ] Notifications desktop (cross-platform)
- [ ] Couleurs et icônes selon criticité
- [ ] Sons d'alerte configurables
- [ ] Badges dans prompt terminal
- [ ] Intégration barre de statut macOS

### Phase 5 : Optimizations (1 semaine)

#### T5.1 - Performance & Efficacité
- [ ] Cache intelligent métriques
- [ ] Requêtes batch vers Kubernetes
- [ ] Compression données historiques
- [ ] Nettoyage automatique anciennes données
- [ ] Parallélisation collecte multi-apps

#### T5.2 - Configuration Avancée
- [ ] `~/.shipyard/monitoring.yaml` - Config globale
- [ ] Profils de monitoring (dev, staging, prod)
- [ ] Seuils adaptatifs par environnement
- [ ] Templates d'alertes réutilisables
- [ ] Import/export configuration

## Spécifications Techniques

### Interface `shipyard monitor`
```
┌─ Shipyard Monitor ─────────────────────────────────────────────────────┐
│ Refreshing every 5s | Press 'q' to quit | Press 'h' for help          │
├────────────────────────────────────────────────────────────────────────┤
│ APP NAME      STATUS   CPU    MEMORY   REPLICAS   LAST DEPLOY   ALERTS │
│ web-app       🟢 OK    45%    67%      3/5        2h ago        0      │
│ api-service   🟡 WARN  78%    45%      2/10       1d ago        2      │
│ worker        🔴 CRIT  12%    89%      1/3        3h ago        1      │
│ database      🟢 OK    23%    56%      1/1        5d ago        0      │
├────────────────────────────────────────────────────────────────────────┤
│ Cluster: ✅ Healthy | Nodes: 3/3 | Total Pods: 47 | Alerts: 3 active  │
└────────────────────────────────────────────────────────────────────────┘
```

### Configuration Monitoring
```yaml
# paas.yaml - extension monitoring
app:
  name: my-app
  image: myapp:latest
  port: 3000

monitoring:
  enabled: true
  health_check:
    path: /health
    interval: 30s
    timeout: 5s
  metrics:
    enabled: true
    path: /metrics
    port: 9090
  alerts:
    cpu_threshold: 80%
    memory_threshold: 85%
    error_rate_threshold: 5%
    response_time_threshold: 1000ms
  retention: 7d
```

### Structure Base de Données
```sql
-- Nouvelles tables monitoring
CREATE TABLE metrics (
    id INTEGER PRIMARY KEY,
    app_id INTEGER,
    metric_type TEXT, -- cpu, memory, network, disk
    value REAL,
    timestamp DATETIME,
    FOREIGN KEY (app_id) REFERENCES apps(id)
);

CREATE TABLE health_checks (
    id INTEGER PRIMARY KEY,
    app_id INTEGER,
    endpoint TEXT,
    status TEXT, -- healthy, unhealthy, timeout
    response_time INTEGER, -- ms
    error_message TEXT,
    checked_at DATETIME,
    FOREIGN KEY (app_id) REFERENCES apps(id)
);

CREATE TABLE alerts (
    id INTEGER PRIMARY KEY,
    app_id INTEGER,
    type TEXT, -- cpu_high, memory_high, error_rate
    threshold REAL,
    current_value REAL,
    status TEXT, -- active, resolved
    created_at DATETIME,
    resolved_at DATETIME,
    FOREIGN KEY (app_id) REFERENCES apps(id)
);
```

## Critères de Succès

### Fonctionnels
- [ ] Monitoring temps réel de 10+ applications simultanément
- [ ] Collecte métriques toutes les 15 secondes
- [ ] Alertes en moins de 1 minute après dépassement seuil
- [ ] Historique consultable sur 7 jours minimum
- [ ] Export de données pour analyse externe

### Techniques
- [ ] CLI reste responsive (< 100ms pour commandes courantes)
- [ ] Base SQLite < 50MB pour 7j d'historique
- [ ] Compatible Kubernetes 1.20+
- [ ] Fonctionne offline (données locales)
- [ ] Cross-platform (Linux, macOS, Windows)

### UX/UI
- [ ] Interface intuitive sans formation
- [ ] Couleurs et indicateurs visuels clairs
- [ ] Aide contextuelle (`shipyard monitor --help`)
- [ ] Configuration en moins de 5 minutes
- [ ] Détection automatique problèmes courants

## Planning Estimé

**Total : 8-10 semaines**
- Phase 1 (Fondations) : 3 semaines
- Phase 2 (Commandes Core) : 3 semaines  
- Phase 3 (Alerting) : 2 semaines
- Phase 4 (Interface) : 2 semaines
- Phase 5 (Optimizations) : 1 semaine

## Risques & Mitigation

### Risques Techniques
- **Performance** : Collecte trop fréquente → cache + batch
- **Stockage** : Base trop lourde → compression + rotation
- **K8s compatibility** : Versions différentes → abstraction API

### Risques UX
- **Complexité** : Trop d'options → defaults intelligents
- **Performance UI** : Lag interface → optimisation refresh
- **Learning curve** : Trop complexe → wizard + templates

## Extensions Futures
- Intégration Grafana/Prometheus externe
- Métriques custom applicatives
- Machine learning pour prédiction
- API REST pour monitoring externe
- Plugin ecosystem pour alertes