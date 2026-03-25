<template>
  <div class="container-fluid px-0">
    <div
      class="d-flex flex-wrap justify-content-between align-items-center gap-3 mb-4"
    >
      <div>
        <h4 class="mb-1">{{ t("agent_upgrades_page.title") }}</h4>
        <div class="text-muted">{{ t("agent_upgrades_page.subtitle") }}</div>
      </div>
      <button
        class="btn btn-outline-secondary"
        @click="refreshAll"
        :disabled="loadingAny"
      >
        <i class="bi bi-arrow-repeat me-1"></i>{{ t("common.refresh") }}
      </button>
    </div>

    <div class="row g-4">
      <div class="col-12 col-xl-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div class="mb-3">
              <h5 class="card-title mb-1">
                {{ t("agent_upgrades_page.create_title") }}
              </h5>
              <div class="text-muted small">
                {{ t("agent_upgrades_page.create_hint") }}
              </div>
            </div>

            <div v-if="!canCreate" class="alert alert-secondary">
              {{ t("agent_upgrades_page.permission_hint") }}
            </div>

            <form @submit.prevent="submitTask">
              <div class="row g-3">
                <div class="col-12">
                  <label class="form-label">{{
                    t("agent_upgrades_page.task_name")
                  }}</label>
                  <input
                    v-model.trim="taskForm.name"
                    class="form-control"
                    :placeholder="
                      t('agent_upgrades_page.task_name_placeholder')
                    "
                  />
                </div>
                <div class="col-12">
                  <label class="form-label">{{
                    t("agent_upgrades_page.source_mode")
                  }}</label>
                  <div class="d-flex flex-column gap-2">
                    <label class="form-check">
                      <input
                        v-model="sourceMode"
                        class="form-check-input"
                        type="radio"
                        value="artifact"
                      />
                      <span class="form-check-label">{{
                        t("agent_upgrades_page.source_artifact")
                      }}</span>
                    </label>
                    <label class="form-check">
                      <input
                        v-model="sourceMode"
                        class="form-check-input"
                        type="radio"
                        value="url"
                      />
                      <span class="form-check-label">{{
                        t("agent_upgrades_page.source_url")
                      }}</span>
                    </label>
                  </div>
                </div>
                <div v-if="sourceMode === 'artifact'" class="col-12">
                  <label class="form-label">{{
                    t("agent_upgrades_page.artifact")
                  }}</label>
                  <div class="input-group">
                    <select
                      v-model="taskForm.artifact_id"
                      class="form-select"
                      @change="applySelectedArtifactMetadata"
                    >
                      <option value="">
                        {{ t("agent_upgrades_page.select_artifact") }}
                      </option>
                      <option
                        v-for="artifact in artifacts"
                        :key="artifact.id"
                        :value="String(artifact.id)"
                      >
                        {{
                          artifact.version
                            ? `${artifact.name} (${artifact.version})`
                            : artifact.name
                        }}
                      </option>
                    </select>
                    <button
                      type="button"
                      class="btn btn-outline-secondary"
                      @click="openArtifactModal"
                      :disabled="!canCreate"
                    >
                      {{ t("agent_upgrades_page.upload_artifact") }}
                    </button>
                  </div>
                  <div class="form-text">
                    {{ t("agent_upgrades_page.artifact_hint") }}
                  </div>
                </div>
                <div v-else class="col-12">
                  <label class="form-label">{{
                    t("agent_upgrades_page.package_url")
                  }}</label>
                  <input
                    v-model.trim="taskForm.package_url"
                    class="form-control"
                    :placeholder="
                      t('agent_upgrades_page.package_url_placeholder')
                    "
                  />
                </div>
                <div class="col-md-6">
                  <label class="form-label">{{
                    t("agent_upgrades_page.target_version")
                  }}</label>
                  <input
                    v-model.trim="taskForm.target_version"
                    class="form-control"
                    :placeholder="
                      t('agent_upgrades_page.target_version_placeholder')
                    "
                  />
                </div>
                <div class="col-md-6">
                  <label class="form-label">{{
                    t("agent_upgrades_page.checksum")
                  }}</label>
                  <input
                    v-model.trim="taskForm.checksum"
                    class="form-control"
                    :placeholder="
                      t('agent_upgrades_page.checksum_placeholder')
                    "
                  />
                </div>
                <div class="col-md-6">
                  <label class="form-label">{{
                    t("agent_upgrades_page.service_unit")
                  }}</label>
                  <input
                    v-model.trim="taskForm.service_unit"
                    class="form-control"
                    :placeholder="
                      t('agent_upgrades_page.service_unit_placeholder')
                    "
                  />
                </div>
                <div class="col-md-6">
                  <label class="form-label">{{
                    t("agent_upgrades_page.binary_path")
                  }}</label>
                  <input
                    v-model.trim="taskForm.binary_path"
                    class="form-control"
                    :placeholder="
                      t('agent_upgrades_page.binary_path_placeholder')
                    "
                  />
                </div>
                <div class="col-12">
                  <label class="form-label">{{
                    t("agent_upgrades_page.target_mode")
                  }}</label>
                  <div class="d-flex flex-column gap-2">
                    <label class="form-check">
                      <input
                        v-model="useAllMatching"
                        class="form-check-input"
                        type="radio"
                        :value="true"
                      />
                      <span class="form-check-label">{{
                        t("agent_upgrades_page.all_matching")
                      }}</span>
                    </label>
                    <label class="form-check">
                      <input
                        v-model="useAllMatching"
                        class="form-check-input"
                        type="radio"
                        :value="false"
                      />
                      <span class="form-check-label">{{
                        t("agent_upgrades_page.selected_nodes_mode")
                      }}</span>
                    </label>
                  </div>
                </div>
              </div>

              <div class="alert alert-info mt-3 mb-0 small">
                {{
                  useAllMatching
                    ? t("agent_upgrades_page.matching_count").replace(
                        "{count}",
                        String(previewTotal),
                      )
                    : t("agent_upgrades_page.selection_summary").replace(
                        "{count}",
                        String(selectedNodeIDs.length),
                      )
                }}
              </div>

              <div class="d-flex flex-wrap gap-2 mt-3">
                <button
                  type="submit"
                  class="btn btn-primary"
                  :disabled="
                    creatingTask ||
                    !canCreate ||
                    (sourceMode === 'artifact'
                      ? !taskForm.artifact_id
                      : !taskForm.package_url) ||
                    (useAllMatching ? previewTotal === 0 : !selectedNodeIDs.length)
                  "
                >
                  <i class="bi bi-arrow-up-circle me-1"></i>
                  {{
                    creatingTask
                      ? t("loading")
                      : t("agent_upgrades_page.create_task")
                  }}
                </button>
                <button
                  type="button"
                  class="btn btn-outline-secondary"
                  @click="resetTaskForm"
                  :disabled="creatingTask"
                >
                  {{ t("bootstrap_page.reset") }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>

      <div class="col-12 col-xl-8">
        <div class="card border-0 shadow-sm">
          <div class="card-body">
            <div class="mb-3">
              <h5 class="card-title mb-1">
                {{ t("agent_upgrades_page.preview_title") }}
              </h5>
              <div class="text-muted small">
                {{ t("agent_upgrades_page.preview_hint") }}
              </div>
            </div>

            <div class="row g-3">
              <div class="col-md-4">
                <label class="form-label">{{
                  t("agent_upgrades_page.search")
                }}</label>
                <input
                  v-model.trim="filters.search"
                  class="form-control"
                  :placeholder="t('agent_upgrades_page.search_placeholder')"
                />
              </div>
              <div class="col-md-4">
                <label class="form-label">{{
                  t("agent_upgrades_page.datacenter")
                }}</label>
                <select
                  v-model="filters.datacenter_id"
                  class="form-select"
                  @change="handleDatacenterChange"
                >
                  <option value="">
                    {{ t("agent_upgrades_page.all_datacenters") }}
                  </option>
                  <option
                    v-for="dc in datacenters"
                    :key="dc.id"
                    :value="String(dc.id)"
                  >
                    {{ dc.alias || dc.name }}
                  </option>
                </select>
              </div>
              <div class="col-md-4">
                <label class="form-label">{{
                  t("agent_upgrades_page.region")
                }}</label>
                <select
                  v-model="filters.region_id"
                  class="form-select"
                  @change="handleRegionChange"
                >
                  <option value="">
                    {{ t("agent_upgrades_page.all_regions") }}
                  </option>
                  <option
                    v-for="region in filteredRegions"
                    :key="region.id"
                    :value="String(region.id)"
                  >
                    {{ region.alias || region.name }}
                  </option>
                </select>
              </div>
              <div class="col-md-4">
                <label class="form-label">{{
                  t("agent_upgrades_page.cluster")
                }}</label>
                <select v-model="filters.cluster_id" class="form-select">
                  <option value="">
                    {{ t("agent_upgrades_page.all_clusters") }}
                  </option>
                  <option
                    v-for="cluster in filteredClusters"
                    :key="cluster.id"
                    :value="String(cluster.id)"
                  >
                    {{ cluster.alias || cluster.name }}
                  </option>
                </select>
              </div>
              <div class="col-md-4">
                <label class="form-label">{{
                  t("agent_upgrades_page.environment")
                }}</label>
                <select v-model="filters.environment_id" class="form-select">
                  <option value="">
                    {{ t("agent_upgrades_page.all_environments") }}
                  </option>
                  <option
                    v-for="env in environments"
                    :key="env.id"
                    :value="String(env.id)"
                  >
                    {{ env.alias || env.name }}
                  </option>
                </select>
              </div>
              <div class="col-md-4">
                <label class="form-label">{{
                  t("agent_upgrades_page.status")
                }}</label>
                <select v-model="filters.status" class="form-select">
                  <option value="">
                    {{ t("agent_upgrades_page.all_status") }}
                  </option>
                  <option value="online">
                    {{ t("nodes_page.online") }}
                  </option>
                  <option value="offline">
                    {{ t("nodes_page.offline") }}
                  </option>
                  <option value="error">{{ t("nodes_page.error") }}</option>
                </select>
              </div>
              <div class="col-md-4">
                <label class="form-label">{{
                  t("agent_upgrades_page.fluent_type")
                }}</label>
                <select v-model="filters.fluent_type" class="form-select">
                  <option value="">
                    {{ t("agent_upgrades_page.all_types") }}
                  </option>
                  <option value="fluentbit">Fluent Bit</option>
                  <option value="fluentd">Fluentd</option>
                </select>
              </div>
              <div class="col-md-4">
                <label class="form-label">{{
                  t("agent_upgrades_page.agent_version")
                }}</label>
                <input
                  v-model.trim="filters.agent_version"
                  class="form-control"
                  :placeholder="
                    t('agent_upgrades_page.agent_version_placeholder')
                  "
                />
              </div>
              <div class="col-md-4 d-flex align-items-end">
                <button
                  class="btn btn-outline-secondary w-100"
                  @click="loadPreview"
                  :disabled="loadingPreview"
                  type="button"
                >
                  <i class="bi bi-funnel me-1"></i>
                  {{
                    loadingPreview
                      ? t("loading")
                      : t("agent_upgrades_page.refresh_preview")
                  }}
                </button>
              </div>
            </div>

            <div class="d-flex flex-wrap gap-2 justify-content-between mt-3">
              <div class="small text-muted">
                {{
                  t("agent_upgrades_page.previewing_count")
                    .replace("{shown}", String(previewNodes.length))
                    .replace("{total}", String(previewTotal))
                }}
              </div>
              <div class="d-flex gap-2">
                <button
                  type="button"
                  class="btn btn-sm btn-outline-secondary"
                  @click="toggleSelectPreview"
                  :disabled="!previewNodes.length || useAllMatching"
                >
                  {{
                    allPreviewSelected
                      ? t("bootstrap_page.clear_selection")
                      : t("agent_upgrades_page.select_preview")
                  }}
                </button>
                <button
                  type="button"
                  class="btn btn-sm btn-outline-secondary"
                  @click="selectedNodeIDs = []"
                  :disabled="!selectedNodeIDs.length || useAllMatching"
                >
                  {{ t("agent_upgrades_page.clear_selection") }}
                </button>
              </div>
            </div>

            <div class="table-responsive mt-3">
              <table class="table table-hover align-middle mb-0">
                <thead>
                  <tr>
                    <th style="width: 48px">
                      <input
                        class="form-check-input"
                        type="checkbox"
                        :checked="allPreviewSelected"
                        :disabled="useAllMatching"
                        @change="toggleSelectPreview"
                      />
                    </th>
                    <th>{{ t("agent_upgrades_page.hostname") }}</th>
                    <th>IP</th>
                    <th>{{ t("agent_upgrades_page.current_version") }}</th>
                    <th>{{ t("agent_upgrades_page.status") }}</th>
                    <th>{{ t("agent_upgrades_page.cluster") }}</th>
                    <th>{{ t("agent_upgrades_page.last_heartbeat") }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="node in previewNodes" :key="node.id">
                    <td>
                      <input
                        class="form-check-input"
                        type="checkbox"
                        :disabled="useAllMatching"
                        :checked="selectedNodeIDs.includes(node.id)"
                        @change="toggleNodeSelection(node.id)"
                      />
                    </td>
                    <td>
                      <div class="fw-semibold">{{ node.hostname }}</div>
                      <div class="small text-muted">{{ node.node_uid }}</div>
                    </td>
                    <td>{{ node.ip_address || "-" }}</td>
                    <td>{{ node.agent_version || "-" }}</td>
                    <td>
                      <span class="badge" :class="statusClass(node.status)">{{
                        statusText(node.status)
                      }}</span>
                    </td>
                    <td>
                      {{
                        node.cluster?.alias ||
                        node.cluster?.name ||
                        t("nodes_page.unassigned")
                      }}
                    </td>
                    <td>{{ formatTime(node.last_heartbeat) }}</td>
                  </tr>
                  <tr v-if="!previewNodes.length">
                    <td colspan="7" class="text-center text-muted py-4">
                      {{ t("agent_upgrades_page.no_candidates") }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm mt-4">
      <div class="card-body p-0">
        <div
          class="d-flex justify-content-between align-items-center p-3 border-bottom"
        >
          <div>
            <h5 class="mb-1">{{ t("agent_upgrades_page.tasks_title") }}</h5>
            <div class="text-muted small">
              {{ t("agent_upgrades_page.tasks_hint") }}
            </div>
          </div>
          <span class="badge bg-secondary-subtle text-secondary-emphasis">{{
            tasks.length
          }}</span>
        </div>

        <div class="table-responsive">
          <table class="table table-hover align-middle mb-0">
            <thead>
              <tr>
                <th>ID</th>
                <th>{{ t("common.name") }}</th>
                <th>{{ t("agent_upgrades_page.target_version") }}</th>
                <th>{{ t("status") }}</th>
                <th>{{ t("deploys_page.progress") }}</th>
                <th>{{ t("deploys_page.creator") }}</th>
                <th>{{ t("deploys_page.created_at") }}</th>
                <th>{{ t("actions") }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="task in tasks" :key="task.id">
                <td>#{{ task.id }}</td>
                <td>
                  <div class="fw-semibold">{{ task.name }}</div>
                  <div class="small text-muted text-break">
                    {{
                      task.artifact?.name ||
                      task.package_url
                    }}
                  </div>
                </td>
                <td>{{ task.target_version || "-" }}</td>
                <td>
                  <span class="badge" :class="statusClass(task.status)">{{
                    upgradeStatusText(task.status)
                  }}</span>
                </td>
                <td>
                  <div class="progress" style="width: 140px">
                    <div
                      class="progress-bar bg-success"
                      :style="{ width: successPct(task) + '%' }"
                    ></div>
                    <div
                      class="progress-bar bg-danger"
                      :style="{ width: failPct(task) + '%' }"
                    ></div>
                  </div>
                  <small
                    >{{ task.success_count }}/{{ task.fail_count }}/{{
                      task.total_nodes
                    }}</small
                  >
                </td>
                <td>{{ task.creator?.username || "-" }}</td>
                <td>{{ formatTime(task.created_at) }}</td>
                <td>
                  <button
                    class="btn btn-sm btn-outline-primary"
                    @click="openDetail(task)"
                  >
                    <i class="bi bi-eye me-1"></i>{{ t("common.details") }}
                  </button>
                </td>
              </tr>
              <tr v-if="!tasks.length">
                <td colspan="8" class="text-center text-muted py-4">
                  {{ t("agent_upgrades_page.no_tasks") }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="modal fade" id="agentUpgradeDetailModal" tabindex="-1">
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{
                t("agent_upgrades_page.detail_title").replace(
                  "{id}",
                  detail?.task?.id || "",
                )
              }}
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
            ></button>
          </div>
          <div v-if="detail" class="modal-body">
            <div class="mb-3">
              <div class="fw-semibold">{{ detail.task.name }}</div>
              <div class="small text-muted text-break">
                {{ detail.task.artifact?.name || detail.task.package_url }}
              </div>
            </div>
            <div class="table-responsive">
              <table class="table table-sm align-middle">
                <thead>
                  <tr>
                    <th>{{ t("agent_upgrades_page.hostname") }}</th>
                    <th>{{ t("agent_upgrades_page.current_version") }}</th>
                    <th>{{ t("agent_upgrades_page.cluster") }}</th>
                    <th>{{ t("status") }}</th>
                    <th>{{ t("deploys_page.message") }}</th>
                    <th>{{ t("agent_upgrades_page.output_excerpt") }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="record in detail.records" :key="record.id">
                    <td>
                      <div class="fw-semibold">{{ record.node?.hostname }}</div>
                      <div class="small text-muted">
                        {{ record.node?.ip_address || "-" }}
                      </div>
                    </td>
                    <td>{{ record.node?.agent_version || "-" }}</td>
                    <td>
                      {{
                        record.node?.cluster?.alias ||
                        record.node?.cluster?.name ||
                        t("nodes_page.unassigned")
                      }}
                    </td>
                    <td>
                      <span class="badge" :class="statusClass(record.status)">{{
                        upgradeStatusText(record.status)
                      }}</span>
                    </td>
                    <td class="small">{{ record.message || "-" }}</td>
                    <td>
                      <code class="small">{{
                        record.output_excerpt || "-"
                      }}</code>
                    </td>
                  </tr>
                  <tr v-if="!detail.records?.length">
                    <td colspan="6" class="text-center text-muted py-4">
                      {{ t("no_data") }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="modal fade" id="agentArtifactModal" tabindex="-1">
      <div class="modal-dialog modal-lg modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{ t("agent_upgrades_page.upload_artifact") }}
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
            ></button>
          </div>
          <div class="modal-body">
            <div class="row g-3">
              <div class="col-12">
                <label class="form-label">{{
                  t("agent_upgrades_page.artifact_file")
                }}</label>
                <input
                  class="form-control"
                  type="file"
                  @change="artifactForm.file = $event.target.files?.[0] || null"
                />
              </div>
              <div class="col-md-6">
                <label class="form-label">{{
                  t("agent_upgrades_page.artifact_name")
                }}</label>
                <input
                  v-model.trim="artifactForm.name"
                  class="form-control"
                  :placeholder="
                    t('agent_upgrades_page.artifact_name_placeholder')
                  "
                />
              </div>
              <div class="col-md-6">
                <label class="form-label">{{
                  t("agent_upgrades_page.artifact_version")
                }}</label>
                <input
                  v-model.trim="artifactForm.version"
                  class="form-control"
                  :placeholder="
                    t('agent_upgrades_page.artifact_version_placeholder')
                  "
                />
              </div>
              <div class="col-12">
                <label class="form-label">{{
                  t("agent_upgrades_page.artifact_description")
                }}</label>
                <textarea
                  v-model.trim="artifactForm.description"
                  class="form-control"
                  rows="3"
                  :placeholder="
                    t('agent_upgrades_page.artifact_description_placeholder')
                  "
                ></textarea>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-outline-secondary" data-bs-dismiss="modal">
              {{ t("cancel") }}
            </button>
            <button
              class="btn btn-primary"
              @click="submitArtifact"
              :disabled="uploadingArtifact"
            >
              {{
                uploadingArtifact
                  ? t("loading")
                  : t("agent_upgrades_page.upload_artifact")
              }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from "vue";
import {
  createAgentUpgradeTask,
  getAgentArtifacts,
  getAgentUpgradeTask,
  getAgentUpgradeTasks,
  getClusters,
  getDataCenters,
  getEnvironments,
  getNodes,
  getRegions,
  uploadAgentArtifact,
} from "../api";
import { useI18n } from "../i18n";
import { useAuthStore } from "../store/auth";

const auth = useAuthStore();
const { t, dateLocale } = useI18n();

const artifacts = ref([]);
const tasks = ref([]);
const detail = ref(null);
const previewNodes = ref([]);
const previewTotal = ref(0);
const selectedNodeIDs = ref([]);
const clusters = ref([]);
const environments = ref([]);
const datacenters = ref([]);
const regions = ref([]);
const loadingAny = ref(false);
const loadingPreview = ref(false);
const creatingTask = ref(false);
const uploadingArtifact = ref(false);
const useAllMatching = ref(true);
const sourceMode = ref("artifact");
const canCreate = computed(() => auth.hasPermission("nodes", "update"));

let pollTimer = null;
let detailModal = null;
let artifactModal = null;

const taskForm = reactive({
  name: "",
  artifact_id: "",
  package_url: "",
  target_version: "",
  checksum: "",
  service_unit: "fluent-manager-agent",
  binary_path: "",
});

const artifactForm = reactive({
  file: null,
  name: "",
  version: "",
  description: "",
});

const filters = reactive({
  search: "",
  datacenter_id: "",
  region_id: "",
  cluster_id: "",
  environment_id: "",
  status: "online",
  fluent_type: "",
  agent_version: "",
});

const filteredRegions = computed(() =>
  filters.datacenter_id
    ? regions.value.filter(
        (region) => String(region.datacenter_id) === String(filters.datacenter_id),
      )
    : regions.value,
);

const filteredClusters = computed(() =>
  clusters.value.filter((cluster) => {
    if (filters.region_id && String(cluster.region_id) !== String(filters.region_id)) {
      return false;
    }
    return true;
  }),
);

const allPreviewSelected = computed(
  () =>
    previewNodes.value.length > 0 &&
    previewNodes.value.every((node) => selectedNodeIDs.value.includes(node.id)),
);

function buildPreviewParams() {
  const params = {
    page: 1,
    page_size: 100,
  };
  if (filters.search) params.search = filters.search;
  if (filters.datacenter_id) params.datacenter_id = filters.datacenter_id;
  if (filters.region_id) params.region_id = filters.region_id;
  if (filters.cluster_id) params.cluster_id = filters.cluster_id;
  if (filters.environment_id) params.environment_id = filters.environment_id;
  if (filters.status) params.status = filters.status;
  if (filters.fluent_type) params.fluent_type = filters.fluent_type;
  if (filters.agent_version) params.agent_version = filters.agent_version;
  return params;
}

function resetTaskForm() {
  taskForm.name = "";
  taskForm.artifact_id = "";
  taskForm.package_url = "";
  taskForm.target_version = "";
  taskForm.checksum = "";
  taskForm.service_unit = "fluent-manager-agent";
  taskForm.binary_path = "";
  sourceMode.value = "artifact";
}

function resetArtifactForm() {
  artifactForm.file = null;
  artifactForm.name = "";
  artifactForm.version = "";
  artifactForm.description = "";
}

function statusClass(status) {
  return {
    "bg-success": status === "completed" || status === "success",
    "bg-warning text-dark": status === "pending" || status === "running",
    "bg-danger": status === "failed",
    "bg-info text-dark": status === "partial",
  };
}

function statusText(status) {
  return (
    {
      online: t("nodes_page.online"),
      offline: t("nodes_page.offline"),
      error: t("nodes_page.error"),
    }[status] || status
  );
}

function upgradeStatusText(status) {
  return (
    {
      pending: t("deploys_page.pending"),
      running: t("deploys_page.running"),
      completed: t("deploys_page.completed"),
      failed: t("deploys_page.failed"),
      partial: t("agent_upgrades_page.partial"),
    }[status] || status
  );
}

function formatTime(value) {
  return value ? new Date(value).toLocaleString(dateLocale.value) : "-";
}

function successPct(task) {
  return task.total_nodes ? (task.success_count / task.total_nodes) * 100 : 0;
}

function failPct(task) {
  return task.total_nodes ? (task.fail_count / task.total_nodes) * 100 : 0;
}

function getErrorMessage(error) {
  return (
    error?.response?.data?.error || error?.message || t("common.request_failed")
  );
}

function handleDatacenterChange() {
  filters.region_id = "";
  filters.cluster_id = "";
}

function handleRegionChange() {
  filters.cluster_id = "";
}

function toggleNodeSelection(id) {
  if (selectedNodeIDs.value.includes(id)) {
    selectedNodeIDs.value = selectedNodeIDs.value.filter((item) => item !== id);
    return;
  }
  selectedNodeIDs.value = [...selectedNodeIDs.value, id];
}

function toggleSelectPreview() {
  if (useAllMatching.value) return;
  if (allPreviewSelected.value) {
    selectedNodeIDs.value = selectedNodeIDs.value.filter(
      (id) => !previewNodes.value.some((node) => node.id === id),
    );
    return;
  }
  const next = new Set(selectedNodeIDs.value);
  previewNodes.value.forEach((node) => next.add(node.id));
  selectedNodeIDs.value = Array.from(next);
}

async function loadPreview() {
  loadingPreview.value = true;
  try {
    const { data } = await getNodes(buildPreviewParams());
    previewNodes.value = data.data || [];
    previewTotal.value = data.total || 0;
    if (!useAllMatching.value) {
      selectedNodeIDs.value = selectedNodeIDs.value.filter((id) =>
        previewNodes.value.some((node) => node.id === id),
      );
    }
  } catch (error) {
    alert(getErrorMessage(error));
  } finally {
    loadingPreview.value = false;
  }
}

async function loadTasks() {
  const { data } = await getAgentUpgradeTasks({ page_size: 50 });
  tasks.value = data.data || [];
}

async function loadArtifacts() {
  const data = await getAgentArtifacts();
  artifacts.value = data.data || [];
  if (
    sourceMode.value === "artifact" &&
    !taskForm.artifact_id &&
    artifacts.value.length
  ) {
    taskForm.artifact_id = String(artifacts.value[0].id);
    applySelectedArtifactMetadata();
  }
}

async function loadLookups() {
  const [clusterRes, envRes, dcRes, regionRes] = await Promise.all([
    getClusters(),
    getEnvironments(),
    getDataCenters(),
    getRegions(),
  ]);
  clusters.value = clusterRes.data.data || clusterRes.data || [];
  environments.value = envRes.data.data || envRes.data || [];
  datacenters.value = dcRes.data.data || dcRes.data || [];
  regions.value = regionRes.data.data || regionRes.data || [];
}

async function refreshAll() {
  loadingAny.value = true;
  try {
    await Promise.all([loadLookups(), loadPreview(), loadTasks(), loadArtifacts()]);
  } finally {
    loadingAny.value = false;
  }
}

function applySelectedArtifactMetadata() {
  if (sourceMode.value !== "artifact") return;
  const artifact = artifacts.value.find(
    (item) => String(item.id) === String(taskForm.artifact_id),
  );
  if (!artifact) return;
  if (artifact.version) {
    taskForm.target_version = artifact.version;
  }
  if (artifact.sha256) {
    taskForm.checksum = artifact.sha256;
  }
}

function openArtifactModal() {
  resetArtifactForm();
  if (!artifactModal) {
    artifactModal = new window.bootstrap.Modal(
      document.getElementById("agentArtifactModal"),
    );
  }
  artifactModal.show();
}

async function submitArtifact() {
  if (!artifactForm.file) {
    alert(t("agent_upgrades_page.artifact_file_required"));
    return;
  }

  uploadingArtifact.value = true;
  try {
    const created = await uploadAgentArtifact({
      file: artifactForm.file,
      name: artifactForm.name || undefined,
      version: artifactForm.version || undefined,
      description: artifactForm.description || undefined,
    });
    await loadArtifacts();
    sourceMode.value = "artifact";
    taskForm.artifact_id = String(created.id);
    applySelectedArtifactMetadata();
    artifactModal?.hide();
  } catch (error) {
    alert(getErrorMessage(error));
  } finally {
    uploadingArtifact.value = false;
  }
}

async function submitTask() {
  creatingTask.value = true;
  try {
    await createAgentUpgradeTask({
      name: taskForm.name || undefined,
      artifact_id:
        sourceMode.value === "artifact" && taskForm.artifact_id
          ? Number(taskForm.artifact_id)
          : undefined,
      package_url:
        sourceMode.value === "url"
          ? taskForm.package_url || undefined
          : undefined,
      target_version: taskForm.target_version || undefined,
      checksum: taskForm.checksum || undefined,
      service_unit: taskForm.service_unit || undefined,
      binary_path: taskForm.binary_path || undefined,
      all_matching: useAllMatching.value,
      node_ids: useAllMatching.value ? [] : selectedNodeIDs.value,
      filters: { ...filters },
    });
    resetTaskForm();
    if (!useAllMatching.value) {
      selectedNodeIDs.value = [];
    }
    await loadTasks();
  } catch (error) {
    alert(getErrorMessage(error));
  } finally {
    creatingTask.value = false;
  }
}

async function openDetail(task) {
  const { data } = await getAgentUpgradeTask(task.id);
  detail.value = data;
  if (!detailModal) {
    detailModal = new window.bootstrap.Modal(
      document.getElementById("agentUpgradeDetailModal"),
    );
  }
  detailModal.show();
}

function startPolling() {
  stopPolling();
  pollTimer = window.setInterval(async () => {
    if (
      !tasks.value.some((task) => task.status === "pending" || task.status === "running")
    ) {
      return;
    }
    try {
      await loadTasks();
      if (detail.value?.task?.id) {
        const { data } = await getAgentUpgradeTask(detail.value.task.id);
        detail.value = data;
      }
    } catch {}
  }, 8000);
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

onMounted(async () => {
  await refreshAll();
  startPolling();
});

onUnmounted(() => {
  stopPolling();
});
</script>
