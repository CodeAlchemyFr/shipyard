# Dockerfile pour admin-escape-nation avec Bun
FROM oven/bun:1-alpine AS builder

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers de dépendances
COPY package.json bun.lock ./

# Installer les dépendances
RUN bun install --frozen-lockfile

# Copier le reste du code source
COPY . .

# Variables d'environnement pour le build (Next.js)
ARG NEXT_PUBLIC_API_URL=https://api.shunt-app.com/v1
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

# Construire l'application
RUN bun run build

# Stage 2: Production stage avec Node.js (Next.js nécessite Node pour le runtime)
FROM node:22-alpine AS runner

# Créer un utilisateur non-root pour la sécurité
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Définir le répertoire de travail
WORKDIR /app

# Variables d'environnement
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000

# Copier les fichiers nécessaires depuis le stage de build
COPY --from=builder /app/public ./public
COPY --from=builder /app/package.json ./
COPY --from=builder /app/next.config.ts ./

# Copier les fichiers de build Next.js
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# Passer à l'utilisateur non-root
USER nextjs

# Exposer le port
EXPOSE 3000

# Commande de démarrage
CMD ["node", "server.js"]