'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Plus, Upload, Trash2, Copy, Download } from 'lucide-react'

interface PaasConfig {
  name: string
  image: string
  port: number
  namespace?: string
  service: {
    exposePublic: boolean
    type: 'ClusterIP' | 'NodePort'
    externalPort?: number
  }
  cicd: {
    enabled: boolean
    imageTag?: string
  }
  resources: {
    cpu: string
    memory: string
  }
  scaling: {
    min: number
    max: number
    targetCPU: number
  }
  env: Array<{key: string, value: string}>
  domains: Array<{hostname: string}>
  health: {
    enabled: boolean
    liveness: {
      path: string
    }
    readiness: {
      path: string
    }
  }
}

export default function Home() {
  const [config, setConfig] = useState<PaasConfig>({
    name: 'my-app',
    image: 'nginx:latest',
    port: 3000,
    service: {
      exposePublic: false,
      type: 'ClusterIP'
    },
    cicd: {
      enabled: false,
      imageTag: '${IMAGE_TAG}'
    },
    resources: {
      cpu: '100m',
      memory: '128Mi'
    },
    scaling: {
      min: 1,
      max: 10,
      targetCPU: 70
    },
    env: [],
    domains: [],
    health: {
      enabled: false,
      liveness: {
        path: '/health'
      },
      readiness: {
        path: '/ready'
      }
    }
  })

  const [selectedRegistry, setSelectedRegistry] = useState('')
  const [imageName, setImageName] = useState('')
  const [generatedYaml, setGeneratedYaml] = useState('')
  const [nameValidation, setNameValidation] = useState<{isValid: boolean, suggestion?: string, message?: string}>({isValid: true})
  const [namespaceValidation, setNamespaceValidation] = useState<{isValid: boolean, suggestion?: string, message?: string}>({isValid: true})

  const registries = {
    'ghcr.io': 'GitHub Container Registry',
    'docker.io': 'Docker Hub', 
    'gcr.io': 'Google Container Registry',
    'registry.gitlab.com': 'GitLab Container Registry',
    'custom': 'Registry personnalisé'
  }

  const generateYaml = () => {
    if (!config.name || !config.image) return ''
    
    let yaml = `# Shipyard Application Configuration
# Generated automatically

app:
  name: ${config.name}
  image: ${config.image}
  port: ${config.port}${config.namespace ? `
  namespace: ${config.namespace}` : ''}

service:
  type: ${config.service.type}`

    if (config.cicd.enabled) {
      yaml += `

cicd:
  enabled: true`
      if (config.cicd.imageTag && config.cicd.imageTag !== '${IMAGE_TAG}') {
        yaml += `
  image_tag: "${config.cicd.imageTag}"`
      }
    }

    if (config.service.type === 'NodePort' && config.service.externalPort) {
      yaml += `
  externalPort: ${config.service.externalPort}`
    }

    yaml += `

resources:
  cpu: ${config.resources.cpu}
  memory: ${config.resources.memory}

scaling:
  min: ${config.scaling.min}
  max: ${config.scaling.max}
  target_cpu: ${config.scaling.targetCPU}`

    if (config.env.length > 0) {
      yaml += `

env:`
      config.env.forEach(envVar => {
        if (envVar.key && envVar.value) {
          yaml += `
  ${envVar.key}: "${envVar.value}"`
        }
      })
    }

    if (config.domains.length > 0) {
      yaml += `

domains:`
      config.domains.forEach(domain => {
        if (domain.hostname) {
          yaml += `
  - ${domain.hostname}`
        }
      })
    }

    if (config.health.enabled) {
      yaml += `

health:
  liveness:
    path: ${config.health.liveness.path}
  readiness:
    path: ${config.health.readiness.path}`
    }

    return yaml
  }

  // Real-time YAML generation
  useEffect(() => {
    const yaml = generateYaml()
    setGeneratedYaml(yaml)
  }, [config])

  // DNS validation function
  const validateDNSName = (name: string) => {
    if (!name) return { isValid: true }
    
    // DNS-1035 rules: lowercase alphanumeric + hyphens, start with letter, end with alphanumeric
    const isValid = /^[a-z]([a-z0-9-]*[a-z0-9])?$/.test(name)
    
    if (isValid) {
      return { isValid: true }
    }
    
    // Generate suggestion
    let suggestion = name.toLowerCase()
      .replace(/_/g, '-')  // Replace underscores with hyphens
      .replace(/[^a-z0-9-]/g, '')  // Remove invalid characters
    
    // Ensure it starts with a letter
    if (suggestion && suggestion[0] >= '0' && suggestion[0] <= '9') {
      suggestion = 'app-' + suggestion
    }
    
    // Remove leading/trailing hyphens
    suggestion = suggestion.replace(/^-+|-+$/g, '')
    
    if (!suggestion) {
      suggestion = 'my-app'
    }
    
    return {
      isValid: false,
      suggestion,
      message: `Nom invalide. Règles DNS: lettres minuscules, chiffres et tirets (-), commencer par une lettre.`
    }
  }

  const updateImageFromRegistry = () => {
    if (selectedRegistry && imageName) {
      let fullImage = ''
      if (selectedRegistry === 'custom') {
        fullImage = imageName
      } else if (selectedRegistry === 'docker.io') {
        fullImage = imageName
      } else {
        fullImage = `${selectedRegistry}/${imageName}`
      }
      setConfig({...config, image: fullImage})
    }
  }

  const addEnvVar = () => {
    setConfig({...config, env: [...config.env, {key: '', value: ''}]})
  }

  const updateEnvVar = (index: number, field: 'key' | 'value', value: string) => {
    const newEnv = [...config.env]
    newEnv[index][field] = value
    setConfig({...config, env: newEnv})
  }

  const removeEnvVar = (index: number) => {
    setConfig({...config, env: config.env.filter((_, i) => i !== index)})
  }

  const addDomain = () => {
    setConfig({...config, domains: [...config.domains, {hostname: ''}]})
  }

  const updateDomain = (index: number, hostname: string) => {
    const newDomains = [...config.domains]
    newDomains[index].hostname = hostname
    setConfig({...config, domains: newDomains})
  }

  const removeDomain = (index: number) => {
    setConfig({...config, domains: config.domains.filter((_, i) => i !== index)})
  }

  const handleEnvFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = (e) => {
      const content = e.target?.result as string
      const envVars = content.split('\n')
        .filter(line => line.trim() && !line.startsWith('#') && line.includes('='))
        .map(line => {
          const [key, ...valueParts] = line.split('=')
          return {
            key: key.trim(),
            value: valueParts.join('=').trim().replace(/^["']|["']$/g, '')
          }
        })
      setConfig({...config, env: [...config.env, ...envVars]})
    }
    reader.readAsText(file)
    
    // Reset input
    event.target.value = ''
  }

  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(generatedYaml)
      alert('paas.yaml copié dans le presse-papiers!')
    } catch (err) {
      console.error('Erreur lors de la copie:', err)
    }
  }

  return (
    <div className="container mx-auto p-6 max-w-7xl">
      <h1 className="text-3xl font-bold text-center mb-8">Générateur PaaS YAML</h1>
      
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="space-y-6">
          {/* Configuration App */}
          <Card>
            <CardHeader>
              <CardTitle>Configuration de l'application</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <Label htmlFor="name">Nom de l'application</Label>
                <Input
                  id="name"
                  placeholder="mon-app"
                  value={config.name}
                  onChange={(e) => {
                    const newName = e.target.value
                    setConfig({...config, name: newName})
                    setNameValidation(validateDNSName(newName))
                  }}
                  className={!nameValidation.isValid ? 'border-red-500' : ''}
                />
                {!nameValidation.isValid && (
                  <div className="text-sm space-y-1">
                    <p className="text-red-600">{nameValidation.message}</p>
                    <p className="text-blue-600">
                      💡 Suggestion: <code className="bg-blue-50 px-1 rounded">{nameValidation.suggestion}</code>
                      <button 
                        onClick={() => {
                          if (nameValidation.suggestion) {
                            setConfig({...config, name: nameValidation.suggestion})
                            setNameValidation({isValid: true})
                          }
                        }}
                        className="ml-2 text-blue-600 underline hover:text-blue-800"
                      >
                        Utiliser
                      </button>
                    </p>
                  </div>
                )}
              </div>

              {/* Image Docker avec registry */}
              <div className="space-y-3">
                <Label>Image Docker</Label>
                <div className="grid grid-cols-2 gap-2">
                  <Select value={selectedRegistry} onValueChange={setSelectedRegistry}>
                    <SelectTrigger>
                      <SelectValue placeholder="Sélectionner registry" />
                    </SelectTrigger>
                    <SelectContent>
                      {Object.entries(registries).map(([key, label]) => (
                        <SelectItem key={key} value={key}>{label}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <Input
                    placeholder={selectedRegistry === 'custom' ? 'registry.com/user/image:tag' : 'user/image:tag'}
                    value={imageName}
                    onChange={(e) => {
                      setImageName(e.target.value)
                      // Auto-update on typing
                      const name = e.target.value
                      if (selectedRegistry && name) {
                        let fullImage = ''
                        if (selectedRegistry === 'custom') {
                          fullImage = name
                        } else if (selectedRegistry === 'docker.io') {
                          fullImage = name
                        } else {
                          fullImage = `${selectedRegistry}/${name}`
                        }
                        setConfig({...config, image: fullImage})
                      }
                    }}
                  />
                </div>
                <Input
                  placeholder="Image complète (auto-générée)"
                  value={config.image}
                  onChange={(e) => setConfig({...config, image: e.target.value})}
                />
              </div>

              <div>
                <Label htmlFor="port">Port de l'application</Label>
                <Input
                  id="port"
                  type="number"
                  placeholder="3000"
                  value={config.port}
                  onChange={(e) => setConfig({...config, port: parseInt(e.target.value) || 3000})}
                />
                <p className="text-xs text-gray-500 mt-1">Port exposé par votre image Docker</p>
              </div>

              <div>
                <Label htmlFor="namespace">Namespace (optionnel)</Label>
                <Input
                  id="namespace"
                  placeholder={`Par défaut: ${config.name || 'nom-de-app'}`}
                  value={config.namespace || ''}
                  onChange={(e) => {
                    const newNamespace = e.target.value
                    setConfig({...config, namespace: newNamespace || undefined})
                    if (newNamespace) {
                      setNamespaceValidation(validateDNSName(newNamespace))
                    } else {
                      setNamespaceValidation({isValid: true})
                    }
                  }}
                  className={!namespaceValidation.isValid ? 'border-red-500' : ''}
                />
                {!namespaceValidation.isValid && (
                  <div className="text-sm space-y-1">
                    <p className="text-red-600">{namespaceValidation.message}</p>
                    <p className="text-blue-600">
                      💡 Suggestion: <code className="bg-blue-50 px-1 rounded">{namespaceValidation.suggestion}</code>
                      <button 
                        onClick={() => {
                          if (namespaceValidation.suggestion) {
                            setConfig({...config, namespace: namespaceValidation.suggestion})
                            setNamespaceValidation({isValid: true})
                          }
                        }}
                        className="ml-2 text-blue-600 underline hover:text-blue-800"
                      >
                        Utiliser
                      </button>
                    </p>
                  </div>
                )}
                <p className="text-xs text-gray-500 mt-1">Si non spécifié, le nom de l'application sera utilisé comme namespace</p>
              </div>
            </CardContent>
          </Card>

          {/* CI/CD Configuration */}
          <Card>
            <CardHeader>
              <CardTitle>⚙️ Configuration CI/CD</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  id="cicdEnabled"
                  checked={config.cicd.enabled}
                  onChange={(e) => setConfig({...config, cicd: {...config.cicd, enabled: e.target.checked}})}
                />
                <Label htmlFor="cicdEnabled">Activer le mode CI/CD</Label>
              </div>
              <div className="text-sm text-gray-600 space-y-2">
                <p><strong>🎯 Mode CI/CD activé :</strong></p>
                <ul className="list-disc list-inside space-y-1 ml-2">
                  <li>Premier déploiement avec l'image réelle spécifiée</li>
                  <li>Image automatiquement remplacée par <code className="bg-gray-100 px-1 rounded">${"{IMAGE_TAG}"}</code></li>
                  <li>Permet les mises à jour via <code className="bg-gray-100 px-1 rounded">kubectl set image</code></li>
                </ul>
              </div>
              
              {config.cicd.enabled && (
                <div>
                  <Label htmlFor="imageTag">Placeholder pour l'image (optionnel)</Label>
                  <Input
                    id="imageTag"
                    placeholder="${IMAGE_TAG}"
                    value={config.cicd.imageTag || ''}
                    onChange={(e) => setConfig({...config, cicd: {...config.cicd, imageTag: e.target.value || '${IMAGE_TAG}'}})}
                  />
                  <p className="text-xs text-gray-500 mt-1">Par défaut: ${"IMAGE_TAG"}</p>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Service Configuration */}
          <Card>
            <CardHeader>
              <CardTitle>Configuration du Service</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-3">
                <div className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    id="exposePublic"
                    checked={config.service.exposePublic}
                    onChange={(e) => {
                      const exposePublic = e.target.checked
                      setConfig({
                        ...config, 
                        service: {
                          ...config.service,
                          exposePublic,
                          type: exposePublic ? 'NodePort' : 'ClusterIP',
                          externalPort: exposePublic ? config.service.externalPort || 30000 : undefined
                        }
                      })
                    }}
                  />
                  <Label htmlFor="exposePublic">Souhaitez-vous exposer un port publique?</Label>
                </div>
                <p className="text-sm text-gray-600">
                  {config.service.exposePublic 
                    ? "🌐 Service NodePort - accessible depuis l'extérieur du cluster" 
                    : "🔒 Service ClusterIP - accessible uniquement dans le cluster"}
                </p>
                
                {config.service.exposePublic && (
                  <div>
                    <Label htmlFor="externalPort">Port externe (NodePort)</Label>
                    <Input
                      id="externalPort"
                      type="number"
                      placeholder="30000"
                      min="30000"
                      max="32767"
                      value={config.service.externalPort || ''}
                      onChange={(e) => setConfig({
                        ...config, 
                        service: {
                          ...config.service,
                          externalPort: parseInt(e.target.value) || 30000
                        }
                      })}
                    />
                    <p className="text-xs text-gray-500 mt-1">Port entre 30000-32767 pour accès externe</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Resources */}
          <Card>
            <CardHeader>
              <CardTitle>Ressources</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="cpu">CPU (en millicores)</Label>
                  <Input
                    id="cpu"
                    placeholder="100m, 500m, 1000m, 2000m..."
                    value={config.resources.cpu}
                    onChange={(e) => setConfig({...config, resources: {...config.resources, cpu: e.target.value}})}
                  />
                  <p className="text-xs text-gray-500 mt-1">Ex: 100m, 500m, 1, 2, 4</p>
                </div>

                <div>
                  <Label htmlFor="memory">Mémoire</Label>
                  <Input
                    id="memory"
                    placeholder="128Mi, 1Gi, 16Gi..."
                    value={config.resources.memory}
                    onChange={(e) => setConfig({...config, resources: {...config.resources, memory: e.target.value}})}
                  />
                  <p className="text-xs text-gray-500 mt-1">Ex: 256Mi, 1Gi, 16Gi, 32Gi</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Scaling */}
          <Card>
            <CardHeader>
              <CardTitle>Scaling</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-3 gap-4">
                <div>
                  <Label htmlFor="min">Min replicas</Label>
                  <Input
                    id="min"
                    type="number"
                    min="1"
                    value={config.scaling.min}
                    onChange={(e) => setConfig({...config, scaling: {...config.scaling, min: parseInt(e.target.value) || 1}})}
                  />
                </div>
                <div>
                  <Label htmlFor="max">Max replicas</Label>
                  <Input
                    id="max"
                    type="number"
                    min="1"
                    value={config.scaling.max}
                    onChange={(e) => setConfig({...config, scaling: {...config.scaling, max: parseInt(e.target.value) || 10}})}
                  />
                </div>
                <div>
                  <Label htmlFor="targetCPU">Target CPU %</Label>
                  <Input
                    id="targetCPU"
                    type="number"
                    min="10"
                    max="100"
                    value={config.scaling.targetCPU}
                    onChange={(e) => setConfig({...config, scaling: {...config.scaling, targetCPU: parseInt(e.target.value) || 70}})}
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Health Checks */}
          <Card>
            <CardHeader>
              <CardTitle>Health Checks</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  id="healthEnabled"
                  checked={config.health.enabled}
                  onChange={(e) => setConfig({...config, health: {...config.health, enabled: e.target.checked}})}
                />
                <Label htmlFor="healthEnabled">Activer les health checks</Label>
              </div>
              
              {config.health.enabled && (
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="liveness">Liveness path</Label>
                    <Input
                      id="liveness"
                      placeholder="/health"
                      value={config.health.liveness.path}
                      onChange={(e) => setConfig({...config, health: {...config.health, liveness: {...config.health.liveness, path: e.target.value}}})}
                    />
                  </div>
                  <div>
                    <Label htmlFor="readiness">Readiness path</Label>
                    <Input
                      id="readiness"
                      placeholder="/ready"
                      value={config.health.readiness.path}
                      onChange={(e) => setConfig({...config, health: {...config.health, readiness: {...config.health.readiness, path: e.target.value}}})}
                    />
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Environment Variables */}
          <Card>
            <CardHeader>
              <CardTitle>Variables d'environnement</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex gap-2">
                <Button onClick={addEnvVar} variant="outline" size="sm">
                  <Plus className="w-4 h-4 mr-1" />
                  Ajouter
                </Button>
                <div>
                  <input
                    type="file"
                    onChange={handleEnvFileUpload}
                    className="hidden"
                    id="envFile"
                  />
                  <Button 
                    onClick={() => document.getElementById('envFile')?.click()}
                    variant="outline" 
                    size="sm"
                  >
                    <Upload className="w-4 h-4 mr-1" />
                    Uploader .env
                  </Button>
                </div>
              </div>
              
              {config.env.map((envVar, index) => (
                <div key={index} className="flex gap-2 items-center">
                  <Input
                    placeholder="CLE"
                    value={envVar.key}
                    onChange={(e) => updateEnvVar(index, 'key', e.target.value)}
                    className="flex-1"
                  />
                  <span className="text-gray-500">=</span>
                  <Input
                    placeholder="valeur"
                    value={envVar.value}
                    onChange={(e) => updateEnvVar(index, 'value', e.target.value)}
                    className="flex-1"
                  />
                  <Button
                    onClick={() => removeEnvVar(index)}
                    variant="outline"
                    size="sm"
                    className="text-red-600 hover:text-red-800 hover:bg-red-50"
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              ))}
            </CardContent>
          </Card>

          {/* DNS Validation Info */}
          <Card>
            <CardHeader>
              <CardTitle>⚙️ Règles de nommage DNS</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3 text-sm">
                <div>
                  <p className="font-medium text-gray-800">📋 Règles DNS-1035 (Kubernetes):</p>
                  <ul className="list-disc list-inside space-y-1 ml-2 text-gray-600">
                    <li>Uniquement lettres minuscules, chiffres et tirets (-)</li>
                    <li>Doit commencer par une lettre</li>
                    <li>Doit finir par une lettre ou un chiffre</li>
                    <li>Pas de underscores (_), espaces ou caractères spéciaux</li>
                  </ul>
                </div>
                
                <div className="bg-blue-50 p-3 rounded border-l-4 border-blue-400">
                  <p className="font-medium text-blue-800">🎩 Noms finaux générés:</p>
                  <div className="space-y-1 text-blue-700">
                    <p>• <strong>Service:</strong> <code>{nameValidation.isValid ? config.name : (nameValidation.suggestion || config.name)}</code></p>
                    <p>• <strong>Namespace:</strong> <code>{config.namespace ? (namespaceValidation.isValid ? config.namespace : (namespaceValidation.suggestion || config.namespace)) : (nameValidation.isValid ? config.name : (nameValidation.suggestion || config.name))}</code></p>
                  </div>
                </div>
                
                <div>
                  <p className="font-medium text-gray-800">🎯 Comportement des namespaces:</p>
                  <ul className="list-disc list-inside space-y-1 ml-2 text-gray-600">
                    <li>Si aucun namespace spécifié → utilise le nom de l'app</li>
                    <li>Chaque app dans son propre namespace pour l'isolation</li>
                    <li>Namespace créé automatiquement si inexistant</li>
                  </ul>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Domains */}
          <Card>
            <CardHeader>
              <CardTitle>Domaines</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <Button onClick={addDomain} variant="outline" size="sm">
                <Plus className="w-4 h-4 mr-1" />
                Ajouter domaine
              </Button>
              
              {config.domains.map((domain, index) => (
                <div key={index} className="flex gap-2 items-center">
                  <Input
                    placeholder="mon-app.exemple.com"
                    value={domain.hostname}
                    onChange={(e) => updateDomain(index, e.target.value)}
                    className="flex-1"
                  />
                  <Button
                    onClick={() => removeDomain(index)}
                    variant="outline"
                    size="sm"
                    className="text-red-600 hover:text-red-800 hover:bg-red-50"
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              ))}
            </CardContent>
          </Card>

          <div className="text-center space-y-2">
            <div className="text-sm text-gray-500">
              ⚡ Le YAML se génère automatiquement en temps réel
            </div>
            {(!nameValidation.isValid || !namespaceValidation.isValid) && (
              <div className="text-sm text-orange-600 bg-orange-50 p-2 rounded border">
                ⚠️ Certains noms ne respectent pas les règles DNS. Utilisez les suggestions ci-dessus.
              </div>
            )}
          </div>
        </div>

        {/* Generated YAML */}
        <Card>
          <CardHeader>
            <CardTitle>paas.yaml généré</CardTitle>
          </CardHeader>
          <CardContent>
            {generatedYaml ? (
              <div className="space-y-4">
                <Textarea
                  value={generatedYaml}
                  readOnly
                  rows={25}
                  className="font-mono text-sm"
                />
                <div className="grid grid-cols-2 gap-4">
                  <Button onClick={copyToClipboard}>
                    <Copy className="w-4 h-4 mr-2" />
                    Copier dans le presse-papiers
                  </Button>
                  <Button 
                    onClick={() => {
                      const blob = new Blob([generatedYaml], { type: 'text/yaml' })
                      const url = URL.createObjectURL(blob)
                      const a = document.createElement('a')
                      a.href = url
                      a.download = 'paas.yaml'
                      a.click()
                      URL.revokeObjectURL(url)
                    }}
                    variant="outline"
                  >
                    <Download className="w-4 h-4 mr-2" />
                    Télécharger paas.yaml
                  </Button>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-500 py-8">
                Remplissez le formulaire pour voir le paas.yaml généré automatiquement
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}