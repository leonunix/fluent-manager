export interface ConfigTemplate {
  id: number
  name: string
  description: string
  fluent_type: string
  content: string
  variables: string
  source_type: string
  source_modules: string
  flow_layout: string
  versions?: ConfigVersion[]
  created_by: number
  creator?: import('./auth').User
  created_at: string
  updated_at: string
}

export interface ConfigVersion {
  id: number
  template_id: number
  template?: ConfigTemplate
  version: number
  content: string
  hash: string
  comment: string
  source_type: string
  source_modules: string
  flow_layout: string
  created_by: number
  creator?: import('./auth').User
  created_at: string
}

export interface ConfigModule {
  id: number
  name: string
  description: string
  module_type: string
  fluent_type: string
  content: string
  content_fluentd: string
  variables: string
  is_builtin: boolean
  preset_kind: string
  preset_key: string
  versions?: ConfigModuleVersion[]
  created_by: number
  creator?: import('./auth').User
  created_at: string
  updated_at: string
}

export interface ConfigModuleVersion {
  id: number
  module_id: number
  module?: ConfigModule
  version: number
  content: string
  variables: string
  hash: string
  comment: string
  created_by: number
  creator?: import('./auth').User
  created_at: string
}

export interface RenderModuleRef {
  module_id: number
  version_id?: number
  variables?: string
}

export interface RenderedConfig {
  id: number
  name: string
  fluent_type: string
  runtime_version: string
  source_modules: string
  variables: string
  content: string
  hash: string
  created_by: number
  creator?: import('./auth').User
  created_at: string
}
