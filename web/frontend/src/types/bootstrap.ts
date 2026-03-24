export interface BootstrapCapability {
  supported: boolean;
  ansible_playbook_path: string;
  sshpass_path: string;
  role_path: string;
  default_agent_binary_path: string;
  default_agent_binary_found: boolean;
  reasons: string[];
}

export interface BootstrapTask {
  id: number;
  name: string;
  status: string;
  message: string;
  cluster_id: number | null;
  cluster?: import("./topology").Cluster;
  fluent_type: string;
  install_runtime: boolean;
  server_url: string;
  agent_binary_source: string;
  agent_binary_path: string;
  agent_download_url: string;
  total_hosts: number;
  success_count: number;
  fail_count: number;
  started_at: string | null;
  finished_at: string | null;
  created_by: number;
  creator?: import("./auth").User;
  created_at: string;
  updated_at: string;
}

export interface BootstrapHost {
  id: number;
  hostname: string;
  ip_address: string;
  ssh_port: number;
  ssh_user: string;
  auth_type: string;
  has_password: boolean;
  has_private_key: boolean;
  has_become_password: boolean;
  node_uid: string;
  labels: string;
  description: string;
  cluster_id: number | null;
  cluster?: import("./topology").Cluster;
  created_by: number;
  creator?: import("./auth").User;
  created_at: string;
  updated_at: string;
}

export interface BootstrapRecord {
  id: number;
  bootstrap_task_id: number;
  bootstrap_host_id: number | null;
  bootstrap_host?: BootstrapHost;
  hostname: string;
  ip_address: string;
  ssh_port: number;
  ssh_user: string;
  auth_type: string;
  node_uid: string;
  labels: string;
  cluster_id: number | null;
  cluster?: import("./topology").Cluster;
  alias: string;
  node_id: number | null;
  node?: import("./node").Node;
  status: string;
  message: string;
  output_excerpt: string;
  created_at: string;
  updated_at: string;
}

export interface BootstrapHostInput {
  hostname: string;
  ip_address: string;
  ssh_port: number;
  ssh_user: string;
  auth_type: string;
  password?: string;
  private_key?: string;
  become_password?: string;
  node_uid?: string;
  labels?: string;
  cluster_id?: number | null;
  description?: string;
}

export interface BootstrapTaskInput {
  name?: string;
  server_url: string;
  agent_api_key?: string;
  cluster_id?: number | null;
  fluent_type: string;
  install_runtime: boolean;
  agent_binary_path?: string;
  agent_download_url?: string;
  host_ids?: number[];
  hosts?: BootstrapHostInput[];
}
