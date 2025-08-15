import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Shipyard",
  description: "Open Source PaaS Platform for Kubernetes - Deploy applications with ease",
  base: '/shipyard/',
  head: [
    ['link', { rel: 'icon', href: '/favicon.ico' }]
  ],
  appearance: false, // Disable dark mode toggle
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    logo: { src: '/logo.png', width: 32, height: 32 },
    
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Get Started', link: '/getting-started' },
      { text: 'CLI Reference', link: '/cli/overview' },
      { text: 'API', link: '/api/overview' },
      { text: 'Guides', link: '/guides/deployment' }
    ],

    sidebar: {
      '/getting-started': [
        {
          text: 'Getting Started',
          items: [
            { text: 'Installation', link: '/getting-started' },
            { text: 'Quick Start', link: '/getting-started/quick-start' },
            { text: 'Configuration', link: '/getting-started/configuration' }
          ]
        }
      ],
      '/cli/': [
        {
          text: 'CLI Reference',
          items: [
            { text: 'Overview', link: '/cli/overview' },
            { text: 'shipyard deploy', link: '/cli/deploy' },
            { text: 'shipyard status', link: '/cli/status' },
            { text: 'shipyard logs', link: '/cli/logs' },
            { text: 'shipyard rollback', link: '/cli/rollback' },
            { text: 'shipyard registry', link: '/cli/registry' },
            { text: 'shipyard domain', link: '/cli/domain' }
          ]
        }
      ],
      '/api/': [
        {
          text: 'API Reference',
          items: [
            { text: 'Overview', link: '/api/overview' },
            { text: 'Authentication', link: '/api/auth' },
            { text: 'Applications', link: '/api/applications' },
            { text: 'Deployments', link: '/api/deployments' },
            { text: 'Domains', link: '/api/domains' }
          ]
        }
      ],
      '/guides/': [
        {
          text: 'Guides',
          items: [
            { text: 'Deploying Applications', link: '/guides/deployment' },
            { text: 'Managing Domains', link: '/guides/domains' },
            { text: 'Private Registries', link: '/guides/registries' },
            { text: 'Environment Variables', link: '/guides/environment' },
            { text: 'Scaling & Resources', link: '/guides/scaling' },
            { text: 'Monitoring & Logs', link: '/guides/monitoring' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/CodeAlchemyFr/shipyard' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2025 Shipyard Contributors'
    },

    editLink: {
      pattern: 'https://github.com/CodeAlchemyFr/shipyard/edit/main/docs/:path'
    },

    search: {
      provider: 'local'
    }
  }
})
