# SSL/TLS Management

Shipyard fournit une gestion SSL/TLS automatique via cert-manager et Let's Encrypt.

## Installation SSL

### Installation automatique

```bash
shipyard ssl install
```

Cette commande va :
1. **Vérifier** si cert-manager est déjà installé
2. **Installer cert-manager** si nécessaire
3. **Demander votre email** pour Let's Encrypt
4. **Créer un ClusterIssuer** Let's Encrypt
5. **Configurer** les certificats automatiques

### Vérification de l'installation

```bash
# Vérifier cert-manager
kubectl get pods -n cert-manager

# Vérifier le ClusterIssuer
kubectl get clusterissuer

# Vérifier les certificats
kubectl get certificates -A
```

## Configuration SSL dans paas.yaml

Les certificats SSL sont automatiquement générés pour tous les domaines listés dans votre `paas.yaml` :

```yaml
app:
  name: my-app
  image: nginx:latest
  port: 80

domains:
  - api.example.com     # Certificat SSL automatique
  - app.example.com     # Certificat SSL automatique
```

## Ingress Controller Support

Shipyard supporte plusieurs ingress controllers :

### Traefik (k3s par défaut)
```yaml
# Annotations automatiques pour Traefik
traefik.ingress.kubernetes.io/router.entrypoints: web,websecure
traefik.ingress.kubernetes.io/router.tls: "true"
```

### nginx-ingress
```yaml
# Annotations automatiques pour nginx
nginx.ingress.kubernetes.io/ssl-redirect: "true"
```

## Processus de validation

1. **Déploiement** : `shipyard deploy` génère l'ingress avec TLS
2. **Demande** : cert-manager demande un certificat à Let's Encrypt
3. **Challenge** : Let's Encrypt valide votre domaine via HTTP-01
4. **Installation** : Le certificat est installé automatiquement
5. **Renouvellement** : cert-manager renouvelle automatiquement avant expiration

## Debugging SSL

### Vérifier l'état des certificats

```bash
# État général
kubectl get certificates -A

# Détails d'un certificat
kubectl describe certificate YOUR-DOMAIN-tls

# Vérifier les challenges
kubectl get challenges -A

# Logs cert-manager
kubectl logs -n cert-manager deployment/cert-manager
```

### Problèmes courants

#### Certificat en état "False"
```bash
kubectl describe certificate YOUR-DOMAIN-tls
```
Vérifiez le message d'erreur dans les conditions.

#### Challenge HTTP-01 échoue
- Vérifiez que votre domaine pointe vers votre cluster
- Vérifiez que le port 80 est accessible depuis internet
- Vérifiez les logs de l'ingress controller

#### Email invalide
```bash
kubectl delete clusterissuer letsencrypt-prod
shipyard ssl install  # Réinstaller avec le bon email
```

## Limites Let's Encrypt

- **Rate limiting** : 50 certificats par domaine/semaine
- **Duplicate certificates** : 5 par semaine
- **Validation** : Le domaine doit être accessible depuis internet

## SSL pour développement

Pour le développement local, vous pouvez :

1. **Désactiver SSL** en enlevant la section `domains:`
2. **Utiliser mkcert** pour des certificats locaux
3. **Utiliser un tunnel** (ngrok, cloudflare tunnel)

## Renouvellement automatique

cert-manager renouvelle automatiquement les certificats :
- **30 jours** avant expiration
- **Backup** : conserve l'ancien certificat
- **Zero-downtime** : pas d'interruption de service

## Configuration avancée

### Staging Let's Encrypt (pour tests)

Pour éviter les limites de rate limiting pendant les tests :

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-staging
    solvers:
    - http01:
        ingress:
          class: traefik
```

### Certificats wildcard (DNS-01)

Pour des certificats `*.example.com`, vous devez configurer DNS-01 avec votre provider DNS.

## Exemples complets

### Application avec SSL
```yaml
app:
  name: webapp
  image: nginx:latest
  port: 80

service:
  type: ClusterIP

domains:
  - webapp.example.com
  - www.example.com

# SSL sera automatiquement configuré pour ces deux domaines
```

### Vérification post-déploiement
```bash
# Déployer
shipyard deploy

# Vérifier le certificat (peut prendre 2-5 minutes)
kubectl get certificate

# Tester HTTPS
curl -I https://webapp.example.com
```