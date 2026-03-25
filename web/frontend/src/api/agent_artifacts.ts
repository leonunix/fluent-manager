import api from "./client";

export const getAgentArtifacts = () =>
  api.get("/agent-artifacts").then(({ data }) => data);

export const uploadAgentArtifact = (payload: {
  file: File;
  name?: string;
  version?: string;
  description?: string;
}) => {
  const formData = new FormData();
  formData.append("file", payload.file);
  if (payload.name) formData.append("name", payload.name);
  if (payload.version) formData.append("version", payload.version);
  if (payload.description) formData.append("description", payload.description);
  return api
    .post("/agent-artifacts", formData, {
      headers: { "Content-Type": "multipart/form-data" },
    })
    .then(({ data }) => data);
};

export const deleteAgentArtifact = (id: number) =>
  api.delete(`/agent-artifacts/${id}`);
