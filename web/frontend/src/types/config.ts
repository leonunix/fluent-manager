export interface ConfigTemplate {
  id: number
  name: string
  description: string
  fluent_type: string
  content: string
  variables: string
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
  created_by: number
  creator?: import('./auth').User
  created_at: string
}
