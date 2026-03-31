import api from "./client";

export const getAgentUpgradeTasks = (params?: Record<string, any>) =>
  api.get("/agent-upgrades", { params });
export const getAgentUpgradeTask = (
  id: number,
  params?: { page?: number; page_size?: number },
) => api.get(`/agent-upgrades/${id}`, { params });
export const createAgentUpgradeTask = (data: Record<string, any>) =>
  api.post("/agent-upgrades", data);
