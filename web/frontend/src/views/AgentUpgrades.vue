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
            taskTotal
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
              <tr
                v-for="task in tasks"
                :key="task.id"
                :class="{ 'fm-task-row-active': detailTaskID === task.id }"
              >
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
        <nav v-if="taskTotal > taskPageSize" class="p-3 pt-0">
          <ul class="pagination justify-content-center mb-0">
            <li class="page-item" :class="{ disabled: taskPage <= 1 }">
              <a
                class="page-link"
                href="#"
                @click.prevent="changeTaskPage(taskPage - 1)"
                >{{ t("common.previous") }}</a
              >
            </li>
            <li class="page-item disabled">
              <span class="page-link">{{
                `${taskPage} / ${taskTotalPages()}`
              }}</span>
            </li>
            <li
              class="page-item"
              :class="{ disabled: taskPage >= taskTotalPages() }"
            >
              <a
                class="page-link"
                href="#"
                @click.prevent="changeTaskPage(taskPage + 1)"
                >{{ t("common.next") }}</a
              >
            </li>
          </ul>
        </nav>
      </div>
    </div>

    <div class="modal fade" id="agentUpgradeDetailModal" tabindex="-1">
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header border-0 pb-0">
            <div>
              <h5 class="modal-title">
                {{
                  t("agent_upgrades_page.detail_title").replace(
                    "{id}",
                    detailTaskID || "",
                  )
                }}
              </h5>
              <div class="small text-muted">
                {{ t("agent_upgrades_page.detail_live_hint") }}
              </div>
            </div>
            <div class="d-flex align-items-center gap-2">
              <button
                type="button"
                class="btn btn-sm btn-outline-secondary"
                @click="refreshDetail()"
                :disabled="detailLoading || detailRefreshing || !detailTaskID"
              >
                <i
                  class="bi me-1"
                  :class="
                    detailRefreshing
                      ? 'bi-arrow-repeat fm-spin'
                      : 'bi-arrow-clockwise'
                  "
                ></i>
                {{ t("common.refresh") }}
              </button>
              <button
                type="button"
                class="btn-close"
                data-bs-dismiss="modal"
              ></button>
            </div>
          </div>
          <div class="modal-body">
            <div
              v-if="detailLoading && !detail"
              class="d-flex flex-column align-items-center justify-content-center text-center text-muted py-5"
            >
              <div class="spinner-border text-primary mb-3" role="status"></div>
              <div>{{ t("agent_upgrades_page.detail_loading") }}</div>
            </div>

            <template v-else-if="detail">
              <section class="fm-upgrade-live-shell mb-4">
                <div
                  class="d-flex flex-wrap justify-content-between align-items-start gap-3"
                >
                  <div>
                    <div class="fm-live-kicker">
                      {{ t("agent_upgrades_page.detail_live_title") }}
                    </div>
                    <div class="h5 mb-1">{{ detail.task.name }}</div>
                    <div class="text-muted small">
                      {{ detailStageText(detail.task) }}
                    </div>
                  </div>
                  <div class="text-start text-md-end">
                    <span
                      class="badge rounded-pill"
                      :class="statusClass(detail.task.status)"
                    >
                      {{ upgradeStatusText(detail.task.status) }}
                    </span>
                    <div class="small text-muted mt-2 text-break">
                      {{ detailTaskSummary(detail.task) }}
                    </div>
                  </div>
                </div>

                <div
                  v-if="detailJustCreated && detailAutoRefreshEnabled"
                  class="alert alert-primary border-0 mt-3 mb-0 py-2"
                >
                  <i class="bi bi-broadcast-pin me-2"></i>
                  {{ t("agent_upgrades_page.detail_created_hint") }}
                </div>

                <div
                  v-if="detailError"
                  class="alert alert-warning border-0 mt-3 mb-0 py-2"
                >
                  <i class="bi bi-exclamation-triangle me-2"></i>
                  {{ detailError }}
                </div>

                <div class="fm-upgrade-flow">
                  <div class="fm-upgrade-flow__rail">
                    <template
                      v-for="(stage, index) in detailFlowStages"
                      :key="stage.key"
                    >
                      <article
                        class="fm-upgrade-flow__stage"
                        :class="`is-${stage.state}`"
                      >
                        <div class="fm-upgrade-flow__icon">
                          <i
                            class="bi"
                            :class="[stage.icon, stage.spin ? 'fm-spin' : '']"
                          ></i>
                        </div>
                        <div class="fm-upgrade-flow__content">
                          <div class="fm-upgrade-flow__label">
                            {{ stage.label }}
                          </div>
                          <div class="fm-upgrade-flow__meta">
                            {{ stage.meta }}
                          </div>
                        </div>
                      </article>
                      <div
                        v-if="index < detailFlowStages.length - 1"
                        class="fm-upgrade-flow__connector"
                        :class="
                          flowConnectorClass(
                            stage,
                            detailFlowStages[index + 1],
                          )
                        "
                      >
                        <span class="fm-upgrade-flow__connector-line"></span>
                      </div>
                    </template>
                  </div>
                </div>

                <div class="progress fm-detail-progress mt-3">
                  <div
                    class="progress-bar bg-success"
                    :style="{ width: detailSuccessPct + '%' }"
                  ></div>
                  <div
                    class="progress-bar bg-danger"
                    :style="{ width: detailFailPct + '%' }"
                  ></div>
                  <div
                    class="progress-bar fm-progress-running"
                    :style="{ width: detailRunningPct + '%' }"
                  ></div>
                  <div
                    class="progress-bar fm-progress-pending"
                    :style="{ width: detailPendingPct + '%' }"
                  ></div>
                </div>

                <div class="row g-3 mt-1">
                  <div class="col-6 col-lg-3">
                    <div class="fm-detail-stat">
                      <div class="fm-detail-stat-label">
                        {{ t("agent_upgrades_page.total_nodes_label") }}
                      </div>
                      <div class="fm-detail-stat-value">
                        {{ detailStats.total }}
                      </div>
                    </div>
                  </div>
                  <div class="col-6 col-lg-3">
                    <div class="fm-detail-stat">
                      <div class="fm-detail-stat-label text-success">
                        {{ t("agent_upgrades_page.success_nodes_label") }}
                      </div>
                      <div class="fm-detail-stat-value">
                        {{ detailStats.success }}
                      </div>
                    </div>
                  </div>
                  <div class="col-6 col-lg-3">
                    <div class="fm-detail-stat">
                      <div class="fm-detail-stat-label text-danger">
                        {{ t("agent_upgrades_page.failed_nodes_label") }}
                      </div>
                      <div class="fm-detail-stat-value">
                        {{ detailStats.failed }}
                      </div>
                    </div>
                  </div>
                  <div class="col-6 col-lg-3">
                    <div class="fm-detail-stat">
                      <div class="fm-detail-stat-label text-info-emphasis">
                        {{
                          detailStats.running
                            ? t("agent_upgrades_page.running_nodes_label")
                            : t("agent_upgrades_page.pending_nodes_label")
                        }}
                      </div>
                      <div class="fm-detail-stat-value">
                        {{ detailStats.running || detailStats.pending }}
                      </div>
                    </div>
                  </div>
                </div>

                <div
                  class="d-flex flex-wrap justify-content-between align-items-center gap-2 mt-3 pt-3 border-top"
                >
                  <div class="d-flex flex-wrap gap-2">
                    <span class="badge text-bg-light border">
                      {{
                        detailAutoRefreshEnabled
                          ? t("agent_upgrades_page.auto_refresh_on").replace(
                              "{seconds}",
                              String(taskPollIntervalSeconds),
                            )
                          : t("agent_upgrades_page.auto_refresh_off")
                      }}
                    </span>
                    <span class="badge text-bg-light border">
                      {{
                        t("agent_upgrades_page.last_refresh").replace(
                          "{time}",
                          formatTime(detailLastRefreshedAt),
                        )
                      }}
                    </span>
                  </div>
                  <div class="small text-muted d-flex flex-wrap gap-3">
                    <span
                      >{{ t("deploys_page.created_at") }}
                      {{ formatTime(detail.task.created_at) }}</span
                    >
                    <span
                      >{{ t("agent_upgrades_page.started_at") }}
                      {{ formatTime(detail.task.started_at) }}</span
                    >
                    <span
                      >{{ t("agent_upgrades_page.finished_at") }}
                      {{ formatTime(detail.task.finished_at) }}</span
                    >
                  </div>
                </div>
              </section>

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
                        <code class="small fm-output-excerpt">{{
                          record.output_excerpt || "-"
                        }}</code>
                      </td>
                    </tr>
                    <tr v-if="!detail.records?.length">
                      <td colspan="6" class="text-center text-muted py-4">
                        {{ t("agent_upgrades_page.no_records") }}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <nav v-if="detailRecordTotal > detailRecordPageSize" class="mt-3">
                <ul class="pagination justify-content-center mb-0">
                  <li
                    class="page-item"
                    :class="{ disabled: detailRecordPage <= 1 }"
                  >
                    <a
                      class="page-link"
                      href="#"
                      @click.prevent="changeDetailRecordPage(detailRecordPage - 1)"
                      >{{ t("common.previous") }}</a
                    >
                  </li>
                  <li class="page-item disabled">
                    <span class="page-link">{{
                      `${detailRecordPage} / ${detailRecordTotalPages()}`
                    }}</span>
                  </li>
                  <li
                    class="page-item"
                    :class="{
                      disabled: detailRecordPage >= detailRecordTotalPages(),
                    }"
                  >
                    <a
                      class="page-link"
                      href="#"
                      @click.prevent="changeDetailRecordPage(detailRecordPage + 1)"
                      >{{ t("common.next") }}</a
                    >
                  </li>
                </ul>
              </nav>
            </template>
          </div>
          <div class="modal-footer border-0 pt-0">
            <button class="btn btn-outline-secondary" data-bs-dismiss="modal">
              {{ t("common.close") }}
            </button>
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
const taskTotal = ref(0);
const taskPage = ref(1);
const taskPageSize = 10;
const detail = ref(null);
const detailTaskID = ref(null);
const detailLoading = ref(false);
const detailRefreshing = ref(false);
const detailError = ref("");
const detailLastRefreshedAt = ref(null);
const detailJustCreated = ref(false);
const detailRecordTotal = ref(0);
const detailRecordPage = ref(1);
const detailRecordPageSize = 20;
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
let detailModalEl = null;
let detailModalHiddenHandler = null;
let artifactModal = null;
const taskPollIntervalMs = 5000;
const taskPollIntervalSeconds = taskPollIntervalMs / 1000;

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

const detailAutoRefreshEnabled = computed(() =>
  isActiveTaskStatus(detail.value?.task?.status),
);

const detailStats = computed(() => {
  const task = detail.value?.task;
  const records = detail.value?.records || [];
  const total = Number(task?.total_nodes || records.length || 0);

  if (!records.length) {
    const success = Number(task?.success_count || 0);
    const failed = Number(task?.fail_count || 0);
    const remaining = Math.max(total - success - failed, 0);
    return {
      total,
      success,
      failed,
      running: task?.status === "running" ? remaining : 0,
      pending: task?.status === "pending" ? remaining : 0,
    };
  }

  return {
    total,
    success: records.filter((record) => isSuccessStatus(record.status)).length,
    failed: records.filter((record) => isFailedStatus(record.status)).length,
    running: records.filter((record) => record.status === "running").length,
    pending: records.filter((record) => record.status === "pending").length,
  };
});

const detailSuccessPct = computed(() =>
  percentage(detailStats.value.success, detailStats.value.total),
);
const detailFailPct = computed(() =>
  percentage(detailStats.value.failed, detailStats.value.total),
);
const detailRunningPct = computed(() =>
  percentage(detailStats.value.running, detailStats.value.total),
);
const detailPendingPct = computed(() =>
  percentage(detailStats.value.pending, detailStats.value.total),
);
const detailResolvedCount = computed(
  () => detailStats.value.success + detailStats.value.failed,
);
const detailOpenCount = computed(() =>
  Math.max(detailStats.value.total - detailResolvedCount.value, 0),
);
const detailDeliveredCount = computed(
  () => detailStats.value.running + detailResolvedCount.value,
);
const detailFlowStages = computed(() => {
  const task = detail.value?.task;
  if (!task) return [];

  const finished =
    task.status === "completed" ||
    task.status === "failed" ||
    task.status === "partial";
  const started =
    Boolean(task.started_at) || isActiveTaskStatus(task.status) || finished;
  const deliveryStarted = detailDeliveredCount.value > 0 || finished;
  const resultCollectionStarted = detailResolvedCount.value > 0 || finished;
  const deliverActive =
    task.status === "running" &&
    deliveryStarted &&
    detailStats.value.running > 0 &&
    !resultCollectionStarted;
  const executeActive = task.status === "running" && detailStats.value.running > 0;
  const verifyActive = task.status === "running" && resultCollectionStarted;

  return [
    {
      key: "queued",
      label: t("agent_upgrades_page.flow_queued"),
      meta:
        task.status === "pending"
          ? t("agent_upgrades_page.flow_meta_waiting")
          : t("agent_upgrades_page.flow_meta_dispatched"),
      icon: "bi-hourglass-split",
      state: task.status === "pending" ? "active" : "done",
      spin: false,
    },
    {
      key: "deliver",
      label: t("agent_upgrades_page.flow_deliver"),
      meta: !started
        ? t("agent_upgrades_page.flow_meta_deliver_pending")
        : deliverActive
          ? t("agent_upgrades_page.flow_meta_delivering")
          : deliveryStarted
            ? t("agent_upgrades_page.flow_meta_delivered")
            : t("agent_upgrades_page.flow_meta_deliver_pending"),
      icon: "bi-send-check",
      state: !started ? "pending" : deliverActive ? "active" : deliveryStarted ? "done" : "pending",
      spin: deliverActive,
    },
    {
      key: "execute",
      label: t("agent_upgrades_page.flow_execute"),
      meta: !started
        ? t("agent_upgrades_page.flow_meta_execute_pending")
        : interpolate(t("agent_upgrades_page.flow_meta_executing"), {
            count: detailStats.value.running,
            total: detailStats.value.total,
          }),
      icon: "bi-cpu",
      state: !started ? "pending" : executeActive ? "active" : resultCollectionStarted ? "done" : "pending",
      spin: executeActive,
    },
    {
      key: "verify",
      label: t("agent_upgrades_page.flow_verify"),
      meta: !resultCollectionStarted
        ? t("agent_upgrades_page.flow_meta_verify_pending")
        : interpolate(t("agent_upgrades_page.flow_meta_verifying"), {
            done: detailResolvedCount.value,
            total: detailStats.value.total,
            remaining: detailOpenCount.value,
          }),
      icon: "bi-clipboard-data",
      state: !started ? "pending" : verifyActive ? "active" : resultCollectionStarted ? "done" : "pending",
      spin: verifyActive,
    },
    {
      key: "finish",
      label: t("agent_upgrades_page.flow_finish"),
      meta: finished
        ? interpolate(t("agent_upgrades_page.flow_meta_finished"), {
            success: detailStats.value.success,
            failed: detailStats.value.failed,
          })
        : t("agent_upgrades_page.flow_meta_finish_pending"),
      icon:
        task.status === "failed"
          ? "bi-x-octagon"
          : task.status === "partial"
            ? "bi-exclamation-circle"
            : "bi-check2-circle",
      state:
        task.status === "completed"
          ? "success"
          : task.status === "failed"
            ? "failed"
            : task.status === "partial"
              ? "warning"
              : "pending",
      spin: false,
    },
  ];
});

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

function taskTotalPages() {
  return Math.max(1, Math.ceil(taskTotal.value / taskPageSize));
}

function detailRecordTotalPages() {
  return Math.max(1, Math.ceil(detailRecordTotal.value / detailRecordPageSize));
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

function percentage(value, total) {
  return total ? (value / total) * 100 : 0;
}

function interpolate(template, replacements = {}) {
  return Object.entries(replacements).reduce(
    (result, [key, value]) =>
      result.replace(new RegExp(`\\{${key}\\}`, "g"), String(value)),
    template,
  );
}

function isSuccessStatus(status) {
  return status === "completed" || status === "success";
}

function isFailedStatus(status) {
  return status === "failed";
}

function isActiveTaskStatus(status) {
  return status === "pending" || status === "running";
}

function detailTaskSummary(task) {
  return task?.artifact?.name || task?.package_url || "-";
}

function detailStageText(task) {
  if (!task) return "-";
  if (task.status === "pending") {
    return t("agent_upgrades_page.stage_pending");
  }
  if (task.status === "running") {
    if (detailResolvedCount.value > 0) {
      return t("agent_upgrades_page.stage_verifying");
    }
    return t("agent_upgrades_page.stage_running");
  }
  if (task.status === "completed") {
    return t("agent_upgrades_page.stage_completed");
  }
  if (task.status === "failed") {
    return t("agent_upgrades_page.stage_failed");
  }
  if (task.status === "partial") {
    return t("agent_upgrades_page.stage_partial");
  }
  return upgradeStatusText(task.status);
}

function flowConnectorClass(currentStage, nextStage) {
  return {
    "is-filled":
      currentStage.state !== "pending" && nextStage.state !== "pending",
    "is-active":
      currentStage.state === "active" || nextStage.state === "active",
    "is-failed":
      nextStage.state === "failed",
    "is-success":
      nextStage.state === "success",
    "is-warning":
      nextStage.state === "warning",
  };
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

async function changeTaskPage(nextPage) {
  if (
    nextPage < 1 ||
    nextPage > taskTotalPages() ||
    nextPage === taskPage.value
  ) {
    return;
  }
  taskPage.value = nextPage;
  await loadTasks();
}

async function changeDetailRecordPage(nextPage) {
  if (
    nextPage < 1 ||
    nextPage > detailRecordTotalPages() ||
    nextPage === detailRecordPage.value
  ) {
    return;
  }
  detailRecordPage.value = nextPage;
  await refreshDetail();
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
  const { data } = await getAgentUpgradeTasks({
    page: taskPage.value,
    page_size: taskPageSize,
  });
  tasks.value = data.data || [];
  taskTotal.value = data.total || 0;
  taskPage.value = data.page || taskPage.value;
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
    const { data: createdTask } = await createAgentUpgradeTask({
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
    taskPage.value = 1;
    await Promise.all([
      loadTasks(),
      openDetail(createdTask, { justCreated: true }),
    ]);
  } catch (error) {
    alert(getErrorMessage(error));
  } finally {
    creatingTask.value = false;
  }
}

function clearDetailState() {
  detail.value = null;
  detailTaskID.value = null;
  detailLoading.value = false;
  detailRefreshing.value = false;
  detailError.value = "";
  detailLastRefreshedAt.value = null;
  detailJustCreated.value = false;
  detailRecordTotal.value = 0;
  detailRecordPage.value = 1;
}

function ensureDetailModal() {
  if (!detailModalEl) {
    detailModalEl = document.getElementById("agentUpgradeDetailModal");
  }
  if (!detailModal && detailModalEl) {
    detailModal = new window.bootstrap.Modal(detailModalEl);
  }
  if (detailModalEl && !detailModalHiddenHandler) {
    detailModalHiddenHandler = () => {
      clearDetailState();
    };
    detailModalEl.addEventListener("hidden.bs.modal", detailModalHiddenHandler);
  }
}

async function refreshDetail(options = {}) {
  const taskID = detailTaskID.value;
  if (!taskID) return null;

  if (options.background) {
    detailRefreshing.value = true;
  } else {
    detailLoading.value = true;
  }

  try {
    const { data } = await getAgentUpgradeTask(taskID, {
      page: detailRecordPage.value,
      page_size: detailRecordPageSize,
    });
    if (detailTaskID.value !== taskID) {
      return null;
    }
    detail.value = data;
    detailError.value = "";
    detailLastRefreshedAt.value = new Date().toISOString();
    detailRecordTotal.value = data.total || 0;
    detailRecordPage.value = data.page || detailRecordPage.value;
    return data;
  } catch (error) {
    if (!options.background && detailTaskID.value === taskID) {
      detailError.value = getErrorMessage(error);
    }
    return null;
  } finally {
    if (options.background) {
      detailRefreshing.value = false;
    } else {
      detailLoading.value = false;
    }
  }
}

async function openDetail(task, options = {}) {
  if (detailTaskID.value !== task.id) {
    detailRecordPage.value = 1;
  }
  detailTaskID.value = task.id;
  detailJustCreated.value = !!options.justCreated;
  detailError.value = "";
  detailRecordTotal.value = Number(task.total_nodes || 0);
  detail.value = {
    task: {
      ...task,
      success_count: Number(task.success_count || 0),
      fail_count: Number(task.fail_count || 0),
      total_nodes: Number(task.total_nodes || 0),
    },
    records: detail.value?.task?.id === task.id ? detail.value.records || [] : [],
  };

  ensureDetailModal();
  detailModal?.show();
  await refreshDetail();
}

function startPolling() {
  stopPolling();
  pollTimer = window.setInterval(async () => {
    if (
      !tasks.value.some((task) => isActiveTaskStatus(task.status)) &&
      !detailAutoRefreshEnabled.value
    ) {
      return;
    }
    try {
      await loadTasks();
      if (detailTaskID.value && detailAutoRefreshEnabled.value) {
        await refreshDetail({ background: true });
      }
    } catch {}
  }, taskPollIntervalMs);
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
  if (detailModalEl && detailModalHiddenHandler) {
    detailModalEl.removeEventListener(
      "hidden.bs.modal",
      detailModalHiddenHandler,
    );
  }
});
</script>

<style scoped>
.fm-task-row-active > td {
  background: linear-gradient(90deg, rgba(13, 110, 253, 0.08), rgba(13, 110, 253, 0.02));
}

.fm-upgrade-live-shell {
  padding: 1.25rem;
  border: 1px solid rgba(13, 110, 253, 0.12);
  border-radius: 1rem;
  background:
    radial-gradient(circle at top right, rgba(13, 202, 240, 0.14), transparent 32%),
    linear-gradient(180deg, rgba(248, 249, 250, 0.96), rgba(255, 255, 255, 1));
}

.fm-live-kicker {
  margin-bottom: 0.35rem;
  font-size: 0.74rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--bs-secondary-color);
}

.fm-upgrade-flow {
  margin-top: 1rem;
  padding-top: 0.15rem;
}

.fm-upgrade-flow__rail {
  display: flex;
  align-items: stretch;
  gap: 0.75rem;
  overflow-x: auto;
  padding-bottom: 0.35rem;
}

.fm-upgrade-flow__stage {
  min-width: 178px;
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
  padding: 0.95rem 1rem;
  border-radius: 18px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.06);
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    border-color 0.2s ease;
}

.fm-upgrade-flow__icon {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  font-size: 1.15rem;
  color: #64748b;
  background: rgba(148, 163, 184, 0.16);
}

.fm-upgrade-flow__content {
  min-width: 0;
}

.fm-upgrade-flow__label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #64748b;
}

.fm-upgrade-flow__meta {
  margin-top: 0.4rem;
  font-size: 0.88rem;
  line-height: 1.4;
  color: #0f172a;
}

.fm-upgrade-flow__stage.is-done {
  border-color: rgba(37, 99, 235, 0.18);
}

.fm-upgrade-flow__stage.is-done .fm-upgrade-flow__icon {
  color: #2563eb;
  background: rgba(59, 130, 246, 0.14);
}

.fm-upgrade-flow__stage.is-done .fm-upgrade-flow__label {
  color: #1d4ed8;
}

.fm-upgrade-flow__stage.is-active {
  transform: translateY(-2px);
  border-color: rgba(14, 165, 233, 0.32);
  background:
    radial-gradient(circle at top left, rgba(125, 211, 252, 0.28), transparent 42%),
    rgba(255, 255, 255, 0.96);
  box-shadow: 0 18px 44px rgba(14, 165, 233, 0.16);
}

.fm-upgrade-flow__stage.is-active .fm-upgrade-flow__icon {
  color: #0369a1;
  background: linear-gradient(135deg, rgba(125, 211, 252, 0.4), rgba(34, 197, 94, 0.18));
  animation: fm-flow-pulse 1.8s ease-in-out infinite;
}

.fm-upgrade-flow__stage.is-active .fm-upgrade-flow__label {
  color: #0284c7;
}

.fm-upgrade-flow__stage.is-pending {
  border-style: dashed;
  opacity: 0.82;
}

.fm-upgrade-flow__stage.is-success {
  border-color: rgba(34, 197, 94, 0.25);
  background:
    radial-gradient(circle at top left, rgba(134, 239, 172, 0.22), transparent 42%),
    rgba(255, 255, 255, 0.96);
}

.fm-upgrade-flow__stage.is-success .fm-upgrade-flow__icon {
  color: #15803d;
  background: rgba(74, 222, 128, 0.18);
}

.fm-upgrade-flow__stage.is-success .fm-upgrade-flow__label {
  color: #15803d;
}

.fm-upgrade-flow__stage.is-failed {
  border-color: rgba(239, 68, 68, 0.28);
  background:
    radial-gradient(circle at top left, rgba(252, 165, 165, 0.22), transparent 42%),
    rgba(255, 255, 255, 0.96);
}

.fm-upgrade-flow__stage.is-failed .fm-upgrade-flow__icon {
  color: #b91c1c;
  background: rgba(248, 113, 113, 0.18);
}

.fm-upgrade-flow__stage.is-failed .fm-upgrade-flow__label {
  color: #b91c1c;
}

.fm-upgrade-flow__stage.is-warning {
  border-color: rgba(245, 158, 11, 0.28);
  background:
    radial-gradient(circle at top left, rgba(253, 224, 71, 0.22), transparent 42%),
    rgba(255, 255, 255, 0.96);
}

.fm-upgrade-flow__stage.is-warning .fm-upgrade-flow__icon {
  color: #b45309;
  background: rgba(251, 191, 36, 0.2);
}

.fm-upgrade-flow__stage.is-warning .fm-upgrade-flow__label {
  color: #b45309;
}

.fm-upgrade-flow__connector {
  position: relative;
  flex: 0 0 54px;
  display: flex;
  align-items: center;
}

.fm-upgrade-flow__connector-line {
  position: relative;
  display: block;
  width: 100%;
  height: 4px;
  border-radius: 999px;
  overflow: hidden;
  background: rgba(148, 163, 184, 0.28);
}

.fm-upgrade-flow__connector-line::after {
  content: "";
  position: absolute;
  inset: 0;
  border-radius: inherit;
  transform: scaleX(0);
  transform-origin: left center;
  background: linear-gradient(90deg, #38bdf8, #2563eb);
  transition: transform 0.25s ease;
}

.fm-upgrade-flow__connector::after {
  content: "";
  position: absolute;
  right: -3px;
  top: 50%;
  transform: translateY(-50%);
  border-left: 8px solid rgba(148, 163, 184, 0.45);
  border-top: 6px solid transparent;
  border-bottom: 6px solid transparent;
}

.fm-upgrade-flow__connector.is-filled .fm-upgrade-flow__connector-line::after,
.fm-upgrade-flow__connector.is-active .fm-upgrade-flow__connector-line::after,
.fm-upgrade-flow__connector.is-failed .fm-upgrade-flow__connector-line::after,
.fm-upgrade-flow__connector.is-success .fm-upgrade-flow__connector-line::after,
.fm-upgrade-flow__connector.is-warning .fm-upgrade-flow__connector-line::after {
  transform: scaleX(1);
}

.fm-upgrade-flow__connector.is-active .fm-upgrade-flow__connector-line::after {
  background: linear-gradient(90deg, #22d3ee, #0ea5e9, #2563eb);
  background-size: 200% 100%;
  animation: fm-flow-glide 1.6s linear infinite;
}

.fm-upgrade-flow__connector.is-active::after {
  border-left-color: #0ea5e9;
}

.fm-upgrade-flow__connector.is-failed .fm-upgrade-flow__connector-line::after {
  background: linear-gradient(90deg, #f87171, #ef4444);
}

.fm-upgrade-flow__connector.is-failed::after {
  border-left-color: #ef4444;
}

.fm-upgrade-flow__connector.is-success .fm-upgrade-flow__connector-line::after {
  background: linear-gradient(90deg, #4ade80, #22c55e);
}

.fm-upgrade-flow__connector.is-success::after {
  border-left-color: #22c55e;
}

.fm-upgrade-flow__connector.is-warning .fm-upgrade-flow__connector-line::after {
  background: linear-gradient(90deg, #facc15, #f59e0b);
}

.fm-upgrade-flow__connector.is-warning::after {
  border-left-color: #f59e0b;
}

.fm-detail-progress {
  height: 0.8rem;
  border-radius: 999px;
  background: rgba(108, 117, 125, 0.12);
}

.fm-progress-running {
  background: #0dcaf0;
}

.fm-progress-pending {
  background: rgba(108, 117, 125, 0.32);
}

.fm-detail-stat {
  height: 100%;
  padding: 0.9rem 1rem;
  border-radius: 0.9rem;
  background: rgba(255, 255, 255, 0.88);
  border: 1px solid rgba(33, 37, 41, 0.08);
}

.fm-detail-stat-label {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--bs-secondary-color);
}

.fm-detail-stat-value {
  margin-top: 0.3rem;
  font-size: 1.6rem;
  font-weight: 700;
  line-height: 1;
  color: var(--bs-body-color);
}

.fm-output-excerpt {
  display: block;
  max-width: 420px;
  white-space: pre-wrap;
  word-break: break-word;
}

.fm-spin {
  animation: fm-spin 1s linear infinite;
}

@keyframes fm-flow-pulse {
  0%,
  100% {
    box-shadow: 0 0 0 0 rgba(14, 165, 233, 0.2);
  }

  50% {
    box-shadow: 0 0 0 10px rgba(14, 165, 233, 0);
  }
}

@keyframes fm-flow-glide {
  from {
    background-position: 0% 0%;
  }

  to {
    background-position: 200% 0%;
  }
}

@keyframes fm-spin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 767.98px) {
  .fm-upgrade-flow__stage {
    min-width: 164px;
  }

  .fm-upgrade-flow__connector {
    flex-basis: 40px;
  }
}
</style>
