import api from "./client";
import type {
  BootstrapCapability,
  BootstrapHost,
  BootstrapHostInput,
  BootstrapRecord,
  BootstrapTask,
  BootstrapTaskInput,
} from "../types";

export const getBootstrapCapability = () =>
  api
    .get<BootstrapCapability>("/bootstrap/capability")
    .then(({ data }) => data);

export const getBootstrapTasks = (params?: Record<string, any>) =>
  api
    .get<{
      data: BootstrapTask[];
      total: number;
      page: number;
      page_size: number;
    }>("/bootstrap/tasks", { params })
    .then(({ data }) => data);

export const getBootstrapTask = (id: number) =>
  api
    .get<{
      task: BootstrapTask;
      records: BootstrapRecord[];
    }>(`/bootstrap/tasks/${id}`)
    .then(({ data }) => data);

export const createBootstrapTask = (payload: BootstrapTaskInput) =>
  api.post<BootstrapTask>("/bootstrap/tasks", payload).then(({ data }) => data);

export const getBootstrapHosts = () =>
  api
    .get<{ data: BootstrapHost[] }>("/bootstrap/hosts")
    .then(({ data }) => data);

export const createBootstrapHost = (payload: BootstrapHostInput) =>
  api.post<BootstrapHost>("/bootstrap/hosts", payload).then(({ data }) => data);

export const createBootstrapHostsBulk = (payload: {
  hosts: BootstrapHostInput[];
}) =>
  api
    .post<{
      data: BootstrapHost[];
      count: number;
    }>("/bootstrap/hosts/bulk", payload)
    .then(({ data }) => data);

export const updateBootstrapHost = (id: number, payload: BootstrapHostInput) =>
  api
    .put<BootstrapHost>(`/bootstrap/hosts/${id}`, payload)
    .then(({ data }) => data);

export const deleteBootstrapHost = (id: number) =>
  api.delete(`/bootstrap/hosts/${id}`);
