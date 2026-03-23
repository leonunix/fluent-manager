import api from './client'

export const getSetupStatus = () => api.get('/setup/status')

export const testDBConnection = (data: {
  driver: string
  host?: string
  port?: number
  user?: string
  password?: string
  db_name?: string
  path?: string
}) => api.post('/setup/test-db', data)

export const initializeSystem = (data: {
  db_driver?: string
  db_host?: string
  db_port?: number
  db_user?: string
  db_password?: string
  db_name?: string
  db_path?: string
  username: string
  password: string
  email?: string
  display_name?: string
}) => api.post('/setup/initialize', data)
