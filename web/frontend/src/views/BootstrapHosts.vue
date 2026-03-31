<template>
  <div class="container-fluid px-0">
    <div
      class="d-flex flex-wrap justify-content-between align-items-center gap-3 mb-4"
    >
      <div>
        <h4 class="mb-1">{{ t("bootstrap_page.title") }}</h4>
        <div class="text-muted">{{ t("bootstrap_page.subtitle") }}</div>
      </div>
      <div class="d-flex gap-2">
        <button
          class="btn btn-outline-primary"
          @click="openHostModal()"
          :disabled="!canCreate"
        >
          <i class="bi bi-plus-lg me-1"></i>{{ t("bootstrap_page.add_host") }}
        </button>
        <button
          class="btn btn-outline-primary"
          @click="openBulkModal()"
          :disabled="!canCreate"
        >
          <i class="bi bi-file-earmark-spreadsheet me-1"></i
          >{{ t("bootstrap_page.bulk_import") }}
        </button>
        <button
          class="btn btn-outline-secondary"
          @click="refreshAll"
          :disabled="loadingAny"
        >
          <i class="bi bi-arrow-repeat me-1"></i>{{ t("common.refresh") }}
        </button>
      </div>
    </div>

    <div class="row g-4">
      <div class="col-12 col-xl-4">
        <div class="card border-0 shadow-sm h-100">
          <div class="card-body">
            <div
              class="d-flex justify-content-between align-items-start gap-3 mb-3"
            >
              <div>
                <h5 class="card-title mb-1">
                  {{ t("bootstrap_page.capability_title") }}
                </h5>
                <div class="text-muted small">
                  {{ t("bootstrap_page.capability_hint") }}
                </div>
              </div>
              <span
                class="badge fs-6"
                :class="
                  capability.supported
                    ? 'bg-success-subtle text-success-emphasis'
                    : 'bg-danger-subtle text-danger-emphasis'
                "
              >
                {{
                  capability.supported
                    ? t("bootstrap_page.supported")
                    : t("bootstrap_page.unsupported")
                }}
              </span>
            </div>

            <div class="fm-capability-list small">
              <div class="fm-capability-item">
                <span>{{ t("bootstrap_page.ansible_path") }}</span>
                <code>{{ capability.ansible_playbook_path || "-" }}</code>
              </div>
              <div class="fm-capability-item">
                <span>{{ t("bootstrap_page.sshpass_path") }}</span>
                <code>{{ capability.sshpass_path || "-" }}</code>
              </div>
              <div class="fm-capability-item">
                <span>{{ t("bootstrap_page.role_path") }}</span>
                <code>{{ capability.role_path || "-" }}</code>
              </div>
              <div class="fm-capability-item">
                <span>{{ t("bootstrap_page.default_binary") }}</span>
                <code>{{ capability.default_agent_binary_path || "-" }}</code>
              </div>
            </div>

            <div
              v-if="capability.reasons?.length"
              class="alert alert-warning mt-3 mb-0"
            >
              <div class="fw-semibold mb-2">
                {{ t("bootstrap_page.capability_issues") }}
              </div>
              <ul class="mb-0 ps-3">
                <li v-for="reason in capability.reasons" :key="reason">
                  {{ reason }}
                </li>
              </ul>
            </div>

            <hr class="my-4" />

            <div class="mb-3">
              <h5 class="card-title mb-1">
                {{ t("bootstrap_page.create_title") }}
              </h5>
              <div class="text-muted small">
                {{ t("bootstrap_page.create_hint") }}
              </div>
            </div>

            <div v-if="!canCreate" class="alert alert-secondary">
              {{ t("bootstrap_page.permission_hint") }}
            </div>

            <form @submit.prevent="submitTask">
              <div class="row g-3">
                <div class="col-12">
                  <label class="form-label">{{
                    t("bootstrap_page.task_name")
                  }}</label>
                  <input
                    v-model.trim="taskForm.name"
                    class="form-control"
                    :placeholder="t('bootstrap_page.task_name_placeholder')"
                  />
                </div>
                <div class="col-12">
                  <label class="form-label">{{
                    t("bootstrap_page.server_url")
                  }}</label>
                  <input
                    v-model.trim="taskForm.server_url"
                    class="form-control"
                    placeholder="https://fm.example.com"
                  />
                </div>
                <div class="col-md-6">
                  <label class="form-label">{{
                    t("bootstrap_page.task_cluster_override")
                  }}</label>
                  <select v-model="taskForm.cluster_id" class="form-select">
                    <option :value="null">
                      {{ t("bootstrap_page.keep_host_cluster") }}
                    </option>
                    <option
                      v-for="cluster in clusters"
                      :key="cluster.id"
                      :value="cluster.id"
                    >
                      {{ cluster.alias || cluster.name }}
                    </option>
                  </select>
                </div>
                <div class="col-md-3">
                  <label class="form-label">{{
                    t("bootstrap_page.fluent_type")
                  }}</label>
                  <select v-model="taskForm.fluent_type" class="form-select">
                    <option value="fluentbit">Fluent Bit</option>
                    <option value="fluentd">Fluentd</option>
                    <option value="auto">Auto</option>
                  </select>
                </div>
                <div class="col-md-3 d-flex align-items-end">
                  <div class="form-check form-switch mb-2">
                    <input
                      id="installRuntime"
                      v-model="taskForm.install_runtime"
                      class="form-check-input"
                      type="checkbox"
                    />
                    <label class="form-check-label" for="installRuntime">{{
                      t("bootstrap_page.install_runtime")
                    }}</label>
                  </div>
                </div>
                <div class="col-md-6">
                  <label class="form-label">{{
                    t("bootstrap_page.agent_binary_path")
                  }}</label>
                  <input
                    v-model.trim="taskForm.agent_binary_path"
                    class="form-control"
                    :placeholder="
                      capability.default_agent_binary_path ||
                      'scripts/ansible/files/fluent-manager-agent'
                    "
                  />
                </div>
                <div class="col-md-6">
                  <label class="form-label">{{
                    t("bootstrap_page.agent_download_url")
                  }}</label>
                  <input
                    v-model.trim="taskForm.agent_download_url"
                    class="form-control"
                    placeholder="https://artifacts.example.com/fluent-manager-agent-linux-amd64"
                  />
                </div>
                <div class="col-12">
                  <label class="form-label">{{
                    t("bootstrap_page.agent_api_key")
                  }}</label>
                  <input
                    v-model.trim="taskForm.agent_api_key"
                    type="password"
                    class="form-control"
                    :placeholder="t('bootstrap_page.agent_api_key_placeholder')"
                  />
                  <div class="form-text">
                    {{ t("bootstrap_page.agent_api_key_hint") }}
                  </div>
                </div>
                <div class="col-12">
                  <label class="form-label">{{
                    t("bootstrap_page.target_mode")
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
                        t("bootstrap_page.all_matching")
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
                        t("bootstrap_page.selected_hosts_mode")
                      }}</span>
                    </label>
                  </div>
                </div>
              </div>

              <div class="alert alert-info mt-3 mb-0 small">
                {{
                  useAllMatching
                    ? t("bootstrap_page.matching_count").replace(
                        "{count}",
                        String(hostTotal),
                      )
                    : t("bootstrap_page.selection_summary").replace(
                        "{count}",
                        String(selectedHostIDs.length),
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
                    !capability.supported ||
                    (useAllMatching ? hostTotal === 0 : !selectedHostIDs.length)
                  "
                >
                  <i class="bi bi-rocket-takeoff me-1"></i
                  >{{
                    creatingTask ? t("loading") : t("bootstrap_page.submit")
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
              <h5 class="card-title mb-1">{{ t("bootstrap_page.hosts_title") }}</h5>
              <div class="text-muted small">
                {{ t("bootstrap_page.hosts_hint") }}
              </div>
            </div>

            <div class="row g-3">
              <div class="col-md-4">
                <label class="form-label">{{
                  t("bootstrap_page.host_search")
                }}</label>
                <input
                  v-model.trim="hostFilters.search"
                  class="form-control"
                  :placeholder="t('bootstrap_page.host_search_placeholder')"
                />
              </div>
              <div class="col-md-4">
                <label class="form-label">{{
                  t("bootstrap_page.datacenter")
                }}</label>
                <select
                  v-model="hostFilters.datacenter_id"
                  class="form-select"
                  @change="handleDatacenterChange"
                >
                  <option value="">
                    {{ t("bootstrap_page.all_datacenters") }}
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
                  t("bootstrap_page.region")
                }}</label>
                <select
                  v-model="hostFilters.region_id"
                  class="form-select"
                  @change="handleRegionChange"
                >
                  <option value="">
                    {{ t("bootstrap_page.all_regions") }}
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
                  t("bootstrap_page.cluster")
                }}</label>
                <select v-model="hostFilters.cluster_id" class="form-select">
                  <option value="">
                    {{ t("bootstrap_page.all_clusters") }}
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
                  t("bootstrap_page.environment")
                }}</label>
                <select
                  v-model="hostFilters.environment_id"
                  class="form-select"
                >
                  <option value="">
                    {{ t("bootstrap_page.all_environments") }}
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
                  t("bootstrap_page.auth_type")
                }}</label>
                <select v-model="hostFilters.auth_type" class="form-select">
                  <option value="">
                    {{ t("bootstrap_page.all_auth_types") }}
                  </option>
                  <option value="private_key">
                    {{ t("bootstrap_page.private_key") }}
                  </option>
                  <option value="password">
                    {{ t("bootstrap_page.password") }}
                  </option>
                </select>
              </div>
              <div class="col-12 d-flex justify-content-end">
                <button
                  class="btn btn-outline-secondary"
                  @click="refreshHosts"
                  :disabled="loadingHosts"
                  type="button"
                >
                  <i class="bi bi-funnel me-1"></i>
                  {{
                    loadingHosts
                      ? t("loading")
                      : t("bootstrap_page.refresh_preview")
                  }}
                </button>
              </div>
            </div>

            <div class="d-flex flex-wrap gap-2 justify-content-between mt-3">
              <div class="small text-muted">
                {{
                  t("bootstrap_page.previewing_count")
                    .replace("{shown}", String(hosts.length))
                    .replace("{total}", String(hostTotal))
                }}
              </div>
              <div class="d-flex gap-2">
                <button
                  class="btn btn-sm btn-outline-secondary"
                  @click="toggleSelectAllHosts"
                  :disabled="!hosts.length || useAllMatching"
                  type="button"
                >
                  {{
                    allHostsSelected
                      ? t("bootstrap_page.clear_selection")
                      : t("bootstrap_page.select_preview")
                  }}
                </button>
                <button
                  class="btn btn-sm btn-outline-secondary"
                  @click="selectedHostIDs = []"
                  :disabled="!selectedHostIDs.length || useAllMatching"
                  type="button"
                >
                  {{ t("bootstrap_page.clear_selection") }}
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
                        :checked="allHostsSelected"
                        :disabled="useAllMatching"
                        @change="toggleSelectAllHosts"
                      />
                    </th>
                    <th>{{ t("bootstrap_page.hostname") }}</th>
                    <th>SSH</th>
                    <th>{{ t("common.cluster") }}</th>
                    <th>{{ t("bootstrap_page.auth_type") }}</th>
                    <th>{{ t("bootstrap_page.credentials") }}</th>
                    <th>{{ t("actions") }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="host in hosts" :key="host.id">
                    <td>
                      <input
                        class="form-check-input"
                        type="checkbox"
                        :disabled="useAllMatching"
                        :checked="selectedHostIDs.includes(host.id)"
                        @change="toggleHostSelection(host.id)"
                      />
                    </td>
                    <td>
                      <div class="fw-semibold">{{ host.hostname }}</div>
                      <div class="small text-muted">
                        {{ host.ip_address || "-" }}
                      </div>
                      <div v-if="host.description" class="small text-muted">
                        {{ host.description }}
                      </div>
                    </td>
                    <td>{{ host.ssh_user }}:{{ host.ssh_port }}</td>
                    <td>{{ host.cluster?.name || "-" }}</td>
                    <td>{{ host.auth_type }}</td>
                    <td class="small">
                      <span
                        v-if="host.has_password"
                        class="badge bg-secondary-subtle text-secondary-emphasis me-1"
                        >{{ t("bootstrap_page.password") }}</span
                      >
                      <span
                        v-if="host.has_private_key"
                        class="badge bg-secondary-subtle text-secondary-emphasis me-1"
                        >{{ t("bootstrap_page.private_key") }}</span
                      >
                      <span
                        v-if="host.has_become_password"
                        class="badge bg-secondary-subtle text-secondary-emphasis"
                        >{{ t("bootstrap_page.become_password") }}</span
                      >
                    </td>
                    <td>
                      <div class="btn-group btn-group-sm">
                        <button
                          class="btn btn-outline-primary"
                          @click="openHostModal(host)"
                          :disabled="!canCreate"
                        >
                          <i class="bi bi-pencil"></i>
                        </button>
                        <button
                          class="btn btn-outline-danger"
                          @click="deleteHostRow(host)"
                          :disabled="!canCreate"
                        >
                          <i class="bi bi-trash"></i>
                        </button>
                      </div>
                    </td>
                  </tr>
                  <tr v-if="!hosts.length">
                    <td colspan="7" class="text-center text-muted py-4">
                      {{ t("bootstrap_page.no_matching_hosts") }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <nav v-if="hostTotal > hostPageSize" class="mt-3">
              <ul class="pagination justify-content-center mb-0">
                <li class="page-item" :class="{ disabled: hostPage <= 1 }">
                  <a
                    class="page-link"
                    href="#"
                    @click.prevent="changeHostPage(hostPage - 1)"
                    >{{ t("common.previous") }}</a
                  >
                </li>
                <li class="page-item disabled">
                  <span class="page-link">{{
                    `${hostPage} / ${Math.max(1, Math.ceil(hostTotal / hostPageSize))}`
                  }}</span>
                </li>
                <li
                  class="page-item"
                  :class="{
                    disabled: hostPage >= Math.ceil(hostTotal / hostPageSize),
                  }"
                >
                  <a
                    class="page-link"
                    href="#"
                    @click.prevent="changeHostPage(hostPage + 1)"
                    >{{ t("common.next") }}</a
                  >
                </li>
              </ul>
            </nav>
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
            <h5 class="mb-1">{{ t("bootstrap_page.tasks_title") }}</h5>
            <div class="text-muted small">
              {{ t("bootstrap_page.tasks_hint") }}
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
                  <div class="small text-muted">
                    {{ task.fluent_type }} · {{ task.total_hosts }} hosts
                  </div>
                </td>
                <td>
                  <span class="badge" :class="statusClass(task.status)">{{
                    statusText(task.status)
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
                      task.total_hosts
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
                <td colspan="7" class="text-center text-muted py-4">
                  {{ t("bootstrap_page.no_tasks") }}
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

    <div class="modal fade" id="bootstrapHostModal" tabindex="-1">
      <div class="modal-dialog modal-lg modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{
                editingHostID
                  ? t("bootstrap_page.edit_host")
                  : t("bootstrap_page.add_host")
              }}
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
            ></button>
          </div>
          <div class="modal-body">
            <div class="row g-3">
              <div class="col-md-6">
                <label class="form-label">{{
                  t("bootstrap_page.hostname")
                }}</label>
                <input
                  v-model.trim="hostForm.hostname"
                  class="form-control"
                  placeholder="node-01"
                />
              </div>
              <div class="col-md-6">
                <label class="form-label">IP</label>
                <input
                  v-model.trim="hostForm.ip_address"
                  class="form-control"
                  placeholder="10.0.1.10"
                />
              </div>
              <div class="col-md-4">
                <label class="form-label">{{
                  t("bootstrap_page.ssh_user")
                }}</label>
                <input
                  v-model.trim="hostForm.ssh_user"
                  class="form-control"
                  placeholder="root"
                />
              </div>
              <div class="col-md-4">
                <label class="form-label">{{
                  t("bootstrap_page.ssh_port")
                }}</label>
                <input
                  v-model.number="hostForm.ssh_port"
                  type="number"
                  min="1"
                  max="65535"
                  class="form-control"
                />
              </div>
              <div class="col-md-4">
                <label class="form-label">{{
                  t("bootstrap_page.auth_type")
                }}</label>
                <select v-model="hostForm.auth_type" class="form-select">
                  <option value="private_key">
                    {{ t("bootstrap_page.private_key") }}
                  </option>
                  <option value="password">
                    {{ t("bootstrap_page.password") }}
                  </option>
                </select>
              </div>
              <div class="col-md-6">
                <label class="form-label">{{
                  t("bootstrap_page.cluster")
                }}</label>
                <select v-model="hostForm.cluster_id" class="form-select">
                  <option :value="null">
                    {{ t("bootstrap_page.no_cluster") }}
                  </option>
                  <option
                    v-for="cluster in clusters"
                    :key="cluster.id"
                    :value="cluster.id"
                  >
                    {{ cluster.alias || cluster.name }}
                  </option>
                </select>
              </div>
              <div class="col-md-6">
                <label class="form-label">{{
                  t("bootstrap_page.node_uid")
                }}</label>
                <input
                  v-model.trim="hostForm.node_uid"
                  class="form-control"
                  placeholder="optional-stable-node-uid"
                />
              </div>
              <div class="col-12">
                <label class="form-label">{{
                  t("bootstrap_page.description")
                }}</label>
                <input
                  v-model.trim="hostForm.description"
                  class="form-control"
                  :placeholder="t('bootstrap_page.description_placeholder')"
                />
              </div>
              <div v-if="hostForm.auth_type === 'password'" class="col-12">
                <label class="form-label">{{
                  t("bootstrap_page.password")
                }}</label>
                <input
                  v-model="hostForm.password"
                  type="password"
                  class="form-control"
                  :placeholder="
                    editingHostID ? t('bootstrap_page.keep_secret_hint') : ''
                  "
                />
              </div>
              <div v-else class="col-12">
                <label class="form-label">{{
                  t("bootstrap_page.private_key")
                }}</label>
                <textarea
                  v-model="hostForm.private_key"
                  class="form-control"
                  rows="6"
                  :placeholder="
                    editingHostID
                      ? t('bootstrap_page.keep_secret_hint')
                      : t('bootstrap_page.private_key_placeholder')
                  "
                ></textarea>
              </div>
              <div class="col-md-6">
                <label class="form-label">{{
                  t("bootstrap_page.become_password")
                }}</label>
                <input
                  v-model="hostForm.become_password"
                  type="password"
                  class="form-control"
                  :placeholder="
                    editingHostID ? t('bootstrap_page.keep_secret_hint') : ''
                  "
                />
              </div>
              <div class="col-12">
                <label class="form-label">{{
                  t("bootstrap_page.labels")
                }}</label>
                <input
                  v-model.trim="hostForm.labels"
                  class="form-control"
                  placeholder='{"env":"prod","team":"api"}'
                />
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-outline-secondary" data-bs-dismiss="modal">
              {{ t("cancel") }}
            </button>
            <button
              class="btn btn-primary"
              @click="submitHostForm"
              :disabled="savingHost || !canCreate"
            >
              {{ savingHost ? t("loading") : t("save") }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="modal fade" id="bootstrapBulkModal" tabindex="-1">
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ t("bootstrap_page.bulk_import") }}</h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
            ></button>
          </div>
          <div class="modal-body">
            <div class="alert alert-info">
              <div class="fw-semibold mb-2">
                {{ t("bootstrap_page.bulk_import_hint") }}
              </div>
              <div class="small mb-2">
                {{ t("bootstrap_page.bulk_import_mode_hint") }}
              </div>
              <div class="small mb-2">
                {{ t("bootstrap_page.bulk_import_columns") }}
              </div>
              <code class="small d-block">{{
                t("bootstrap_page.bulk_import_example")
              }}</code>
            </div>
            <div class="row g-3 mb-3">
              <div class="col-md-4">
                <label class="form-label">{{
                  t("bootstrap_page.ssh_user")
                }}</label>
                <input
                  v-model.trim="bulkForm.ssh_user"
                  class="form-control"
                  placeholder="root"
                />
              </div>
              <div class="col-md-4">
                <label class="form-label">{{
                  t("bootstrap_page.ssh_port")
                }}</label>
                <input
                  v-model.number="bulkForm.ssh_port"
                  type="number"
                  min="1"
                  max="65535"
                  class="form-control"
                />
              </div>
              <div class="col-md-4">
                <label class="form-label">{{
                  t("bootstrap_page.auth_type")
                }}</label>
                <select v-model="bulkForm.auth_type" class="form-select">
                  <option value="private_key">
                    {{ t("bootstrap_page.private_key") }}
                  </option>
                  <option value="password">
                    {{ t("bootstrap_page.password") }}
                  </option>
                </select>
              </div>
              <div class="col-md-6">
                <label class="form-label">{{
                  t("bootstrap_page.cluster")
                }}</label>
                <select v-model="bulkForm.cluster_id" class="form-select">
                  <option :value="null">
                    {{ t("bootstrap_page.no_cluster") }}
                  </option>
                  <option
                    v-for="cluster in clusters"
                    :key="cluster.id"
                    :value="cluster.id"
                  >
                    {{ cluster.alias || cluster.name }}
                  </option>
                </select>
              </div>
              <div class="col-md-6">
                <label class="form-label">{{
                  t("bootstrap_page.become_password")
                }}</label>
                <input
                  v-model="bulkForm.become_password"
                  type="password"
                  class="form-control"
                />
              </div>
              <div v-if="bulkForm.auth_type === 'password'" class="col-12">
                <label class="form-label">{{
                  t("bootstrap_page.password")
                }}</label>
                <input
                  v-model="bulkForm.password"
                  type="password"
                  class="form-control"
                />
              </div>
              <div v-else class="col-12">
                <label class="form-label">{{
                  t("bootstrap_page.private_key")
                }}</label>
                <textarea
                  v-model="bulkForm.private_key"
                  class="form-control"
                  rows="5"
                  :placeholder="t('bootstrap_page.private_key_placeholder')"
                ></textarea>
              </div>
              <div class="col-12">
                <label class="form-label">{{
                  t("bootstrap_page.labels")
                }}</label>
                <input
                  v-model.trim="bulkForm.labels"
                  class="form-control"
                  placeholder='{"env":"prod","team":"api"}'
                />
              </div>
            </div>
            <textarea
              v-model="bulkImportText"
              class="form-control"
              rows="12"
              :placeholder="t('bootstrap_page.bulk_import_placeholder')"
            ></textarea>
          </div>
          <div class="modal-footer">
            <button class="btn btn-outline-secondary" data-bs-dismiss="modal">
              {{ t("cancel") }}
            </button>
            <button
              class="btn btn-primary"
              @click="submitBulkImport"
              :disabled="importingHosts || !canCreate"
            >
              {{
                importingHosts
                  ? t("loading")
                  : t("bootstrap_page.bulk_import_submit")
              }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="modal fade" id="bootstrapDetailModal" tabindex="-1">
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{
                t("bootstrap_page.detail_title").replace(
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
              <div class="text-muted small">
                {{ detail.task.message || "-" }}
              </div>
            </div>
            <div class="table-responsive">
              <table class="table table-sm align-middle">
                <thead>
                  <tr>
                    <th>{{ t("bootstrap_page.hostname") }}</th>
                    <th>IP</th>
                    <th>{{ t("bootstrap_page.ssh_user") }}</th>
                    <th>{{ t("common.cluster") }}</th>
                    <th>{{ t("status") }}</th>
                    <th>{{ t("deploys_page.message") }}</th>
                    <th>{{ t("bootstrap_page.output_excerpt") }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="record in detail.records" :key="record.id">
                    <td>
                      <div class="fw-semibold">{{ record.hostname }}</div>
                      <div v-if="record.node?.id" class="small text-muted">
                        Node #{{ record.node.id }}
                      </div>
                    </td>
                    <td>{{ record.ip_address }}</td>
                    <td>{{ record.ssh_user }}:{{ record.ssh_port }}</td>
                    <td>
                      {{
                        record.cluster?.name ||
                        record.bootstrap_host?.cluster?.name ||
                        "-"
                      }}
                    </td>
                    <td>
                      <span class="badge" :class="statusClass(record.status)">{{
                        statusText(record.status)
                      }}</span>
                    </td>
                    <td class="small">{{ record.message || "-" }}</td>
                    <td>
                      <code class="small">{{
                        record.output_excerpt || "-"
                      }}</code>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from "vue";
import {
  createBootstrapHost,
  createBootstrapHostsBulk,
  createBootstrapTask,
  deleteBootstrapHost,
  getBootstrapCapability,
  getBootstrapHostsFiltered,
  getBootstrapTask,
  getBootstrapTasks,
  getClusters,
  getDataCenters,
  getEnvironments,
  getRegions,
  updateBootstrapHost,
} from "../api";
import { useI18n } from "../i18n";
import { useAuthStore } from "../store/auth";

const auth = useAuthStore();
const { t, dateLocale } = useI18n();

const capability = ref({
  supported: false,
  ansible_playbook_path: "",
  sshpass_path: "",
  role_path: "",
  default_agent_binary_path: "",
  default_agent_binary_found: false,
  reasons: [],
});
const hosts = ref([]);
const hostTotal = ref(0);
const hostPage = ref(1);
const hostPageSize = 50;
const tasks = ref([]);
const taskTotal = ref(0);
const taskPage = ref(1);
const taskPageSize = 10;
const clusters = ref([]);
const datacenters = ref([]);
const regions = ref([]);
const environments = ref([]);
const detail = ref(null);
const selectedHostIDs = ref([]);
const loadingAny = ref(false);
const loadingHosts = ref(false);
const creatingTask = ref(false);
const savingHost = ref(false);
const canCreate = computed(() => auth.hasPermission("nodes", "create"));
const useAllMatching = ref(true);
const allHostsSelected = computed(
  () =>
    hosts.value.length > 0 &&
    hosts.value.every((host) => selectedHostIDs.value.includes(host.id)),
);

const filteredRegions = computed(() =>
  hostFilters.datacenter_id
    ? regions.value.filter(
        (region) =>
          String(region.datacenter_id) === String(hostFilters.datacenter_id),
      )
    : regions.value,
);

const filteredClusters = computed(() =>
  clusters.value.filter((cluster) => {
    if (
      hostFilters.region_id &&
      String(cluster.region_id) !== String(hostFilters.region_id)
    ) {
      return false;
    }
    return true;
  }),
);

let pollTimer = null;
let hostModal = null;
let bulkModal = null;
let detailModal = null;
const bulkImportText = ref("");
const importingHosts = ref(false);

const taskForm = reactive({
  name: "",
  server_url: typeof window !== "undefined" ? window.location.origin : "",
  agent_api_key: "",
  cluster_id: null,
  fluent_type: "fluentbit",
  install_runtime: true,
  agent_binary_path: "",
  agent_download_url: "",
});

const hostFilters = reactive({
  search: "",
  datacenter_id: "",
  region_id: "",
  cluster_id: "",
  environment_id: "",
  auth_type: "",
});

const hostForm = reactive({
  hostname: "",
  ip_address: "",
  ssh_port: 22,
  ssh_user: "root",
  auth_type: "private_key",
  password: "",
  private_key: "",
  become_password: "",
  node_uid: "",
  labels: "",
  cluster_id: null,
  description: "",
});
const editingHostID = ref(null);
const bulkForm = reactive({
  ssh_user: "root",
  ssh_port: 22,
  auth_type: "private_key",
  password: "",
  private_key: "",
  become_password: "",
  cluster_id: null,
  labels: "",
});

function statusClass(status) {
  return {
    "bg-success": status === "completed" || status === "success",
    "bg-warning text-dark": status === "running" || status === "pending",
    "bg-danger": status === "failed",
  };
}

function statusText(status) {
  return (
    {
      pending: t("deploys_page.pending"),
      running: t("deploys_page.running"),
      completed: t("deploys_page.completed"),
      failed: t("deploys_page.failed"),
      success: t("deploys_page.completed"),
    }[status] || status
  );
}

function successPct(task) {
  return task.total_hosts ? (task.success_count / task.total_hosts) * 100 : 0;
}

function failPct(task) {
  return task.total_hosts ? (task.fail_count / task.total_hosts) * 100 : 0;
}

function formatTime(value) {
  return value ? new Date(value).toLocaleString(dateLocale.value) : "-";
}

function getErrorMessage(error) {
  return (
    error?.response?.data?.error || error?.message || t("common.request_failed")
  );
}

function buildHostParams() {
  const params = {
    page: hostPage.value,
    page_size: hostPageSize,
  };
  if (hostFilters.search) params.search = hostFilters.search;
  if (hostFilters.datacenter_id) params.datacenter_id = hostFilters.datacenter_id;
  if (hostFilters.region_id) params.region_id = hostFilters.region_id;
  if (hostFilters.cluster_id) params.cluster_id = hostFilters.cluster_id;
  if (hostFilters.environment_id)
    params.environment_id = hostFilters.environment_id;
  if (hostFilters.auth_type) params.auth_type = hostFilters.auth_type;
  return params;
}

function resetTaskForm() {
  taskForm.name = "";
  taskForm.server_url =
    typeof window !== "undefined"
      ? window.location.origin
      : taskForm.server_url;
  taskForm.agent_api_key = "";
  taskForm.cluster_id = null;
  taskForm.fluent_type = "fluentbit";
  taskForm.install_runtime = true;
  taskForm.agent_binary_path = capability.value.default_agent_binary_found
    ? capability.value.default_agent_binary_path
    : "";
  taskForm.agent_download_url = "";
  useAllMatching.value = true;
}

function resetHostForm() {
  editingHostID.value = null;
  hostForm.hostname = "";
  hostForm.ip_address = "";
  hostForm.ssh_port = 22;
  hostForm.ssh_user = "root";
  hostForm.auth_type = "private_key";
  hostForm.password = "";
  hostForm.private_key = "";
  hostForm.become_password = "";
  hostForm.node_uid = "";
  hostForm.labels = "";
  hostForm.cluster_id = null;
  hostForm.description = "";
}

function handleDatacenterChange() {
  hostFilters.region_id = "";
  hostFilters.cluster_id = "";
}

function handleRegionChange() {
  hostFilters.cluster_id = "";
}

function changeHostPage(nextPage) {
  if (
    nextPage < 1 ||
    nextPage > Math.max(1, Math.ceil(hostTotal.value / hostPageSize)) ||
    nextPage === hostPage.value
  ) {
    return;
  }
  hostPage.value = nextPage;
  loadHosts();
}

function taskTotalPages() {
  return Math.max(1, Math.ceil(taskTotal.value / taskPageSize));
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

function openHostModal(host = null) {
  resetHostForm();
  if (host) {
    editingHostID.value = host.id;
    hostForm.hostname = host.hostname;
    hostForm.ip_address = host.ip_address;
    hostForm.ssh_port = host.ssh_port;
    hostForm.ssh_user = host.ssh_user;
    hostForm.auth_type = host.auth_type;
    hostForm.node_uid = host.node_uid || "";
    hostForm.labels = host.labels || "";
    hostForm.cluster_id = host.cluster_id;
    hostForm.description = host.description || "";
  }
  if (!hostModal) {
    hostModal = new window.bootstrap.Modal(
      document.getElementById("bootstrapHostModal"),
    );
  }
  hostModal.show();
}

function openBulkModal() {
  bulkImportText.value = "";
  bulkForm.ssh_user = "root";
  bulkForm.ssh_port = 22;
  bulkForm.auth_type = "private_key";
  bulkForm.password = "";
  bulkForm.private_key = "";
  bulkForm.become_password = "";
  bulkForm.cluster_id = null;
  bulkForm.labels = "";
  if (!bulkModal) {
    bulkModal = new window.bootstrap.Modal(
      document.getElementById("bootstrapBulkModal"),
    );
  }
  bulkModal.show();
}

function toggleHostSelection(id) {
  if (selectedHostIDs.value.includes(id)) {
    selectedHostIDs.value = selectedHostIDs.value.filter((item) => item !== id);
    return;
  }
  selectedHostIDs.value = [...selectedHostIDs.value, id];
}

function toggleSelectAllHosts() {
  if (useAllMatching.value) return;
  const pageIDs = hosts.value.map((host) => host.id);
  if (allHostsSelected.value) {
    selectedHostIDs.value = selectedHostIDs.value.filter(
      (id) => !pageIDs.includes(id),
    );
    return;
  }
  const next = new Set(selectedHostIDs.value);
  pageIDs.forEach((id) => next.add(id));
  selectedHostIDs.value = Array.from(next);
}

async function loadCapability() {
  capability.value = await getBootstrapCapability();
  if (
    capability.value.default_agent_binary_found &&
    !taskForm.agent_binary_path
  ) {
    taskForm.agent_binary_path = capability.value.default_agent_binary_path;
  }
}

async function loadHosts() {
  loadingHosts.value = true;
  try {
    const data = await getBootstrapHostsFiltered(buildHostParams());
    hosts.value = data.data || [];
    hostTotal.value = data.total || 0;
  } finally {
    loadingHosts.value = false;
  }
}

async function refreshHosts() {
  hostPage.value = 1;
  await loadHosts();
}

async function loadTasks() {
  const data = await getBootstrapTasks({
    page: taskPage.value,
    page_size: taskPageSize,
  });
  tasks.value = data.data || [];
  taskTotal.value = data.total || 0;
  taskPage.value = data.page || taskPage.value;
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
    await Promise.all([
      loadCapability(),
      loadHosts(),
      loadTasks(),
      loadLookups(),
    ]);
  } finally {
    loadingAny.value = false;
  }
}

async function submitHostForm() {
  savingHost.value = true;
  try {
    const payload = {
      hostname: hostForm.hostname,
      ip_address: hostForm.ip_address,
      ssh_port: Number(hostForm.ssh_port) || 22,
      ssh_user: hostForm.ssh_user,
      auth_type: hostForm.auth_type,
      password:
        hostForm.auth_type === "password"
          ? hostForm.password || undefined
          : undefined,
      private_key:
        hostForm.auth_type === "private_key"
          ? hostForm.private_key || undefined
          : undefined,
      become_password: hostForm.become_password || undefined,
      node_uid: hostForm.node_uid || undefined,
      labels: hostForm.labels || undefined,
      cluster_id: hostForm.cluster_id,
      description: hostForm.description || undefined,
    };
    if (editingHostID.value) {
      await updateBootstrapHost(editingHostID.value, payload);
    } else {
      await createBootstrapHost(payload);
    }
    hostModal?.hide();
    await loadHosts();
  } catch (error) {
    alert(getErrorMessage(error));
  } finally {
    savingHost.value = false;
  }
}

function parseBulkImportText() {
  const lines = bulkImportText.value
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean);

  if (!lines.length) {
    throw new Error(t("bootstrap_page.bulk_import_empty"));
  }

  if (!bulkForm.ssh_user.trim()) {
    throw new Error(t("bootstrap_page.bulk_import_missing_ssh_user"));
  }
  if (bulkForm.auth_type === "password" && !bulkForm.password) {
    throw new Error(t("bootstrap_page.bulk_import_missing_password"));
  }
  if (bulkForm.auth_type === "private_key" && !bulkForm.private_key.trim()) {
    throw new Error(t("bootstrap_page.bulk_import_missing_private_key"));
  }

  return lines.map((line, index) => {
    const fields = line.includes("\t")
      ? line.split("\t")
      : line.split(/[,\s]+/);
    if (fields.length < 1) {
      throw new Error(
        t("bootstrap_page.bulk_import_invalid").replace(
          "{line}",
          String(index + 1),
        ),
      );
    }

    const [hostname, ip_address = "", node_uid = "", description = ""] =
      fields.map((item) => item.trim());

    if (!hostname) {
      throw new Error(
        t("bootstrap_page.bulk_import_invalid").replace(
          "{line}",
          String(index + 1),
        ),
      );
    }

    return {
      hostname,
      ip_address,
      ssh_user: bulkForm.ssh_user,
      auth_type: bulkForm.auth_type,
      ssh_port: Number(bulkForm.ssh_port) || 22,
      cluster_id: bulkForm.cluster_id,
      node_uid: node_uid || undefined,
      labels: bulkForm.labels || undefined,
      description: description || undefined,
      become_password: bulkForm.become_password || undefined,
      password:
        bulkForm.auth_type === "password"
          ? bulkForm.password || undefined
          : undefined,
      private_key:
        bulkForm.auth_type === "private_key"
          ? bulkForm.private_key || undefined
          : undefined,
    };
  });
}

async function submitBulkImport() {
  importingHosts.value = true;
  try {
    const hostsPayload = parseBulkImportText();
    await createBootstrapHostsBulk({ hosts: hostsPayload });
    bulkModal?.hide();
    await refreshHosts();
  } catch (error) {
    alert(getErrorMessage(error));
  } finally {
    importingHosts.value = false;
  }
}

async function deleteHostRow(host) {
  if (
    !window.confirm(
      t("bootstrap_page.confirm_delete_host").replace("{name}", host.hostname),
    )
  )
    return;
  try {
    await deleteBootstrapHost(host.id);
    selectedHostIDs.value = selectedHostIDs.value.filter((id) => id !== host.id);
    await loadHosts();
  } catch (error) {
    alert(getErrorMessage(error));
  }
}

async function submitTask() {
  creatingTask.value = true;
  try {
    await createBootstrapTask({
      name: taskForm.name || undefined,
      server_url: taskForm.server_url,
      agent_api_key: taskForm.agent_api_key || undefined,
      cluster_id: taskForm.cluster_id,
      fluent_type: taskForm.fluent_type,
      install_runtime: taskForm.install_runtime,
      agent_binary_path: taskForm.agent_binary_path || undefined,
      agent_download_url: taskForm.agent_download_url || undefined,
      all_matching: useAllMatching.value,
      filters: { ...hostFilters },
      host_ids: useAllMatching.value ? [] : selectedHostIDs.value,
    });
    resetTaskForm();
    selectedHostIDs.value = [];
    taskPage.value = 1;
    await loadTasks();
  } catch (error) {
    alert(getErrorMessage(error));
  } finally {
    creatingTask.value = false;
  }
}

async function openDetail(task) {
  detail.value = await getBootstrapTask(task.id);
  if (!detailModal) {
    detailModal = new window.bootstrap.Modal(
      document.getElementById("bootstrapDetailModal"),
    );
  }
  detailModal.show();
}

function startPolling() {
  stopPolling();
  pollTimer = window.setInterval(async () => {
    try {
      await loadTasks();
      if (detail.value?.task?.id) {
        detail.value = await getBootstrapTask(detail.value.task.id);
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
