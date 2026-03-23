<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h4 class="mb-0">{{ t('users_page.title') }}</h4>
    </div>

    <!-- Tabs -->
    <ul class="nav nav-tabs mb-3">
      <li class="nav-item">
        <a class="nav-link" :class="{ active: activeTab === 'users' }" href="#" @click.prevent="activeTab = 'users'">
          <i class="bi bi-people me-1"></i>{{ t('users_page.tab_users') }}
        </a>
      </li>
      <li class="nav-item">
        <a class="nav-link" :class="{ active: activeTab === 'groups' }" href="#" @click.prevent="activeTab = 'groups'">
          <i class="bi bi-collection me-1"></i>{{ t('users_page.tab_groups') }}
        </a>
      </li>
      <li class="nav-item">
        <a class="nav-link" :class="{ active: activeTab === 'auth_settings' }" href="#" @click.prevent="activeTab = 'auth_settings'">
          <i class="bi bi-gear me-1"></i>{{ t('users_page.tab_auth_settings') }}
        </a>
      </li>
    </ul>

    <!-- ==================== USERS TAB ==================== -->
    <div v-show="activeTab === 'users'">
      <div class="d-flex justify-content-end mb-3">
        <button class="btn btn-primary" @click="openCreate">
          <i class="bi bi-plus-lg me-1"></i>{{ t('users_page.create') }}
        </button>
      </div>

      <div class="card border-0 shadow-sm">
        <div class="card-body p-0">
          <table class="table table-hover mb-0">
            <thead>
              <tr>
                <th>{{ t('users_page.username') }}</th>
                <th>{{ t('users_page.display_name') }}</th>
                <th>{{ t('users_page.email') }}</th>
                <th>{{ t('users_page.auth_source') }}</th>
                <th>{{ t('users_page.roles') }}</th>
                <th>{{ t('users_page.tab_groups') }}</th>
                <th>{{ t('users_page.scopes') }}</th>
                <th>{{ t('status') }}</th>
                <th>{{ t('users_page.last_login') }}</th>
                <th>{{ t('actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in users" :key="u.id">
                <td><strong>{{ u.username }}</strong></td>
                <td>{{ u.display_name || '-' }}</td>
                <td>{{ u.email || '-' }}</td>
                <td><span class="badge bg-secondary">{{ u.auth_source }}</span></td>
                <td>
                  <div v-if="u.roles?.length" class="fm-role-chip-list">
                    <span v-for="r in u.roles" :key="r.id" class="fm-role-chip" :style="{ '--role-accent': roleAccent(r) }">
                      <i class="bi bi-shield-check"></i> {{ r.name }}
                    </span>
                  </div>
                  <span v-else class="text-muted small">{{ t('users_page.role_pick_empty') }}</span>
                </td>
                <td>
                  <div v-if="u.groups?.length" class="fm-role-chip-list">
                    <span v-for="g in u.groups" :key="g.id" class="badge bg-purple me-1">{{ g.name }}</span>
                  </div>
                  <span v-else class="text-muted small">-</span>
                </td>
                <td>
                  <span v-if="!userScopesMap[u.id] || userScopesMap[u.id].length === 0" class="badge bg-dark">{{ t('users_page.global') }}</span>
                  <span v-for="s in (userScopesMap[u.id] || [])" :key="s.id" class="badge bg-info me-1">
                    {{ scopeIcon(s.scope_type) }} {{ s.scope_name || s.scope_type + ':' + s.scope_id }}
                  </span>
                </td>
                <td>
                  <span :class="u.is_active ? 'bg-success' : 'bg-danger'" class="badge">
                    {{ u.is_active ? t('users_page.active') : t('users_page.inactive') }}
                  </span>
                </td>
                <td>{{ formatTime(u.last_login_at) }}</td>
                <td>
                  <button class="btn btn-sm btn-outline-primary me-1" @click="openEdit(u)"><i class="bi bi-pencil"></i></button>
                  <button class="btn btn-sm btn-outline-info me-1" @click="openScopeEdit(u)" :title="t('users_page.scope_manage_title')"><i class="bi bi-shield-lock"></i></button>
                  <button v-if="u.username !== 'admin'" class="btn btn-sm btn-outline-danger" @click="handleDelete(u)"><i class="bi bi-trash"></i></button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ==================== GROUPS TAB ==================== -->
    <div v-show="activeTab === 'groups'">
      <div class="d-flex justify-content-end mb-3">
        <button class="btn btn-primary" @click="openGroupCreate">
          <i class="bi bi-plus-lg me-1"></i>{{ t('users_page.group_create') }}
        </button>
      </div>

      <div class="card border-0 shadow-sm">
        <div class="card-body p-0">
          <table class="table table-hover mb-0">
            <thead>
              <tr>
                <th>{{ t('users_page.group_name') }}</th>
                <th>{{ t('users_page.description') }}</th>
                <th>{{ t('users_page.roles') }}</th>
                <th>{{ t('users_page.scopes') }}</th>
                <th>{{ t('users_page.group_members') }}</th>
                <th>{{ t('actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="g in groups" :key="g.id">
                <td><strong>{{ g.name }}</strong></td>
                <td>{{ g.description || '-' }}</td>
                <td>
                  <div v-if="g.roles?.length" class="fm-role-chip-list">
                    <span v-for="r in g.roles" :key="r.id" class="fm-role-chip" :style="{ '--role-accent': roleAccent(r) }">
                      <i class="bi bi-shield-check"></i> {{ r.name }}
                    </span>
                  </div>
                  <span v-else class="text-muted small">-</span>
                </td>
                <td>
                  <span v-if="!g.scopes?.length" class="badge bg-dark">{{ t('users_page.global') }}</span>
                  <span v-for="s in (g.scopes || [])" :key="s.id" class="badge bg-info me-1">
                    {{ scopeIcon(s.scope_type) }} {{ s.scope_name || s.scope_type + ':' + s.scope_id }}
                  </span>
                </td>
                <td><span class="badge bg-secondary">{{ g.member_count || 0 }}</span></td>
                <td>
                  <button class="btn btn-sm btn-outline-primary me-1" @click="openGroupEdit(g)"><i class="bi bi-pencil"></i></button>
                  <button class="btn btn-sm btn-outline-danger" @click="handleGroupDelete(g)"><i class="bi bi-trash"></i></button>
                </td>
              </tr>
              <tr v-if="!groups.length">
                <td colspan="6" class="text-center text-muted py-3">{{ t('users_page.group_empty') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ==================== AUTH SETTINGS TAB ==================== -->
    <div v-show="activeTab === 'auth_settings'">
      <!-- LDAP -->
      <div class="card border-0 shadow-sm mb-4">
        <div class="card-header d-flex justify-content-between align-items-center" @click="ldapExpanded = !ldapExpanded" style="cursor:pointer">
          <h6 class="mb-0"><i class="bi bi-diagram-3 me-2"></i>{{ t('auth_settings.ldap_title') }}</h6>
          <i class="bi" :class="ldapExpanded ? 'bi-chevron-up' : 'bi-chevron-down'"></i>
        </div>
        <div v-show="ldapExpanded" class="card-body">
          <div class="row g-3">
            <div class="col-md-12">
              <div class="form-check form-switch">
                <input class="form-check-input" type="checkbox" v-model="ldapForm.enabled">
                <label class="form-check-label">{{ t('auth_settings.enabled') }}</label>
              </div>
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('auth_settings.host') }}</label>
              <input v-model="ldapForm.host" type="text" class="form-control" placeholder="ldap.example.com">
            </div>
            <div class="col-md-3">
              <label class="form-label">{{ t('auth_settings.port') }}</label>
              <input v-model.number="ldapForm.port" type="number" class="form-control" placeholder="389">
            </div>
            <div class="col-md-3">
              <div class="form-check form-switch mt-4">
                <input class="form-check-input" type="checkbox" v-model="ldapForm.use_tls">
                <label class="form-check-label">{{ t('auth_settings.use_tls') }}</label>
              </div>
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('auth_settings.bind_dn') }}</label>
              <input v-model="ldapForm.bind_dn" type="text" class="form-control" placeholder="cn=admin,dc=example,dc=com">
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('auth_settings.bind_password') }}</label>
              <input v-model="ldapForm.bind_password" type="password" class="form-control">
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('auth_settings.base_dn') }}</label>
              <input v-model="ldapForm.base_dn" type="text" class="form-control" placeholder="dc=example,dc=com">
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('auth_settings.user_filter') }}</label>
              <input v-model="ldapForm.user_filter" type="text" class="form-control" placeholder="(uid=%s)">
            </div>
            <div class="col-md-4">
              <label class="form-label">{{ t('auth_settings.attr_username') }}</label>
              <input v-model="ldapForm.attributes.username" type="text" class="form-control" placeholder="uid">
            </div>
            <div class="col-md-4">
              <label class="form-label">{{ t('auth_settings.attr_email') }}</label>
              <input v-model="ldapForm.attributes.email" type="text" class="form-control" placeholder="mail">
            </div>
            <div class="col-md-4">
              <label class="form-label">{{ t('auth_settings.attr_name') }}</label>
              <input v-model="ldapForm.attributes.name" type="text" class="form-control" placeholder="cn">
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('auth_settings.sync_strategy') }}</label>
              <select v-model="ldapForm.group_sync_strategy" class="form-select">
                <option value="always">{{ t('auth_settings.sync_always') }}</option>
                <option value="first_login">{{ t('auth_settings.sync_first_login') }}</option>
              </select>
            </div>
          </div>

          <!-- LDAP Group Mappings -->
          <h6 class="mt-4 mb-3"><i class="bi bi-arrow-left-right me-1"></i>{{ t('auth_settings.group_mappings') }}</h6>
          <table class="table table-sm" v-if="ldapMappings.length">
            <thead><tr><th>{{ t('auth_settings.external_group') }}</th><th>{{ t('auth_settings.system_group') }}</th><th></th></tr></thead>
            <tbody>
              <tr v-for="(m, idx) in ldapMappings" :key="idx">
                <td><input v-model="m.external_group_name" class="form-control form-control-sm" :placeholder="t('auth_settings.external_group_placeholder')"></td>
                <td>
                  <select v-model="m.group_id" class="form-select form-select-sm">
                    <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
                  </select>
                </td>
                <td><button class="btn btn-sm btn-outline-danger" @click="ldapMappings.splice(idx, 1)"><i class="bi bi-trash"></i></button></td>
              </tr>
            </tbody>
          </table>
          <button class="btn btn-sm btn-outline-secondary me-2" @click="ldapMappings.push({ external_group_name: '', group_id: null })">
            <i class="bi bi-plus me-1"></i>{{ t('auth_settings.add_mapping') }}
          </button>

          <div class="mt-4 d-flex gap-2">
            <button class="btn btn-outline-info" @click="testLDAP" :disabled="ldapTesting">
              <i class="bi bi-plug me-1"></i>{{ ldapTesting ? '...' : t('auth_settings.test_connection') }}
            </button>
            <button class="btn btn-primary" @click="saveLDAP">{{ t('save') }}</button>
          </div>
          <div v-if="ldapTestResult" class="mt-2">
            <div class="alert small py-2" :class="ldapTestResult.success ? 'alert-success' : 'alert-danger'">
              {{ ldapTestResult.success ? ldapTestResult.message : ldapTestResult.error }}
            </div>
          </div>
        </div>
      </div>

      <!-- SAML -->
      <div class="card border-0 shadow-sm">
        <div class="card-header d-flex justify-content-between align-items-center" @click="samlExpanded = !samlExpanded" style="cursor:pointer">
          <h6 class="mb-0"><i class="bi bi-shield-lock me-2"></i>{{ t('auth_settings.saml_title') }}</h6>
          <i class="bi" :class="samlExpanded ? 'bi-chevron-up' : 'bi-chevron-down'"></i>
        </div>
        <div v-show="samlExpanded" class="card-body">
          <div class="row g-3">
            <div class="col-md-12">
              <div class="form-check form-switch">
                <input class="form-check-input" type="checkbox" v-model="samlForm.enabled">
                <label class="form-check-label">{{ t('auth_settings.enabled') }}</label>
              </div>
            </div>
            <div class="col-md-12">
              <label class="form-label">{{ t('auth_settings.idp_metadata') }}</label>
              <input v-model="samlForm.idp_metadata" type="text" class="form-control" placeholder="https://idp.example.com/metadata">
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('auth_settings.entity_id') }}</label>
              <input v-model="samlForm.entity_id" type="text" class="form-control">
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('auth_settings.acs_url') }}</label>
              <input v-model="samlForm.acs_url" type="text" class="form-control">
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('auth_settings.cert_data') }}</label>
              <textarea v-model="samlForm.cert_data" class="form-control" rows="4" placeholder="-----BEGIN CERTIFICATE-----&#10;...&#10;-----END CERTIFICATE-----&#10;&#10;or file path: /etc/saml/cert.pem"></textarea>
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('auth_settings.key_data') }}</label>
              <textarea v-model="samlForm.key_data" class="form-control" rows="4" placeholder="-----BEGIN PRIVATE KEY-----&#10;...&#10;-----END PRIVATE KEY-----&#10;&#10;or file path: /etc/saml/key.pem"></textarea>
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('auth_settings.group_attribute') }}</label>
              <input v-model="samlForm.group_attribute" type="text" class="form-control" placeholder="memberOf">
            </div>
            <div class="col-md-6">
              <label class="form-label">{{ t('auth_settings.sync_strategy') }}</label>
              <select v-model="samlForm.group_sync_strategy" class="form-select">
                <option value="always">{{ t('auth_settings.sync_always') }}</option>
                <option value="first_login">{{ t('auth_settings.sync_first_login') }}</option>
              </select>
            </div>
          </div>

          <!-- SAML Group Mappings -->
          <h6 class="mt-4 mb-3"><i class="bi bi-arrow-left-right me-1"></i>{{ t('auth_settings.group_mappings') }}</h6>
          <table class="table table-sm" v-if="samlMappings.length">
            <thead><tr><th>{{ t('auth_settings.external_group') }}</th><th>{{ t('auth_settings.system_group') }}</th><th></th></tr></thead>
            <tbody>
              <tr v-for="(m, idx) in samlMappings" :key="idx">
                <td><input v-model="m.external_group_name" class="form-control form-control-sm" :placeholder="t('auth_settings.external_group_placeholder')"></td>
                <td>
                  <select v-model="m.group_id" class="form-select form-select-sm">
                    <option v-for="g in groups" :key="g.id" :value="g.id">{{ g.name }}</option>
                  </select>
                </td>
                <td><button class="btn btn-sm btn-outline-danger" @click="samlMappings.splice(idx, 1)"><i class="bi bi-trash"></i></button></td>
              </tr>
            </tbody>
          </table>
          <button class="btn btn-sm btn-outline-secondary me-2" @click="samlMappings.push({ external_group_name: '', group_id: null })">
            <i class="bi bi-plus me-1"></i>{{ t('auth_settings.add_mapping') }}
          </button>

          <div class="mt-4">
            <button class="btn btn-primary" @click="saveSAML">{{ t('save') }}</button>
          </div>
        </div>
      </div>

    </div>

    <!-- ==================== USER MODAL ==================== -->
    <div class="modal fade" id="userModal" tabindex="-1">
      <div class="modal-dialog modal-lg modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ form.id ? t('users_page.edit_user') : t('users_page.create_user') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="row g-4">
              <div class="col-lg-5">
                <div class="mb-3">
                  <label class="form-label">{{ t('users_page.username') }}</label>
                  <input v-model="form.username" type="text" class="form-control" :disabled="!!form.id" required>
                </div>
                <div class="mb-3">
                  <label class="form-label">{{ t('users_page.display_name') }}</label>
                  <input v-model="form.display_name" type="text" class="form-control">
                </div>
                <div class="mb-3">
                  <label class="form-label">{{ t('users_page.email') }}</label>
                  <input v-model="form.email" type="email" class="form-control">
                </div>
                <div class="mb-3">
                  <label class="form-label">Password{{ form.id ? t('users_page.password_keep') : '' }}</label>
                  <input v-model="form.password" type="password" class="form-control" :required="!form.id">
                </div>
                <!-- Group selection -->
                <div class="mb-0">
                  <label class="form-label">{{ t('users_page.tab_groups') }}</label>
                  <div class="d-flex flex-wrap gap-2">
                    <button
                      v-for="g in groups"
                      :key="g.id"
                      type="button"
                      class="btn btn-sm"
                      :class="form.group_ids.includes(g.id) ? 'btn-primary' : 'btn-outline-secondary'"
                      @click="toggleUserGroup(g.id)"
                    >
                      {{ g.name }}
                    </button>
                  </div>
                </div>
              </div>

              <div class="col-lg-7">
                <div class="fm-role-panel">
                  <div class="fm-role-panel__header">
                    <div>
                      <div class="fm-role-panel__eyebrow">{{ t('users_page.role_selection') }}</div>
                      <h6 class="fm-role-panel__title mb-1">{{ t('users_page.roles') }}</h6>
                      <p class="text-muted small mb-0">{{ t('users_page.role_pick_hint') }}</p>
                    </div>
                    <button v-if="form.role_ids.length" type="button" class="btn btn-sm btn-outline-secondary" @click="clearRoles">
                      {{ t('users_page.clear_roles') }}
                    </button>
                  </div>

                  <div class="fm-role-summary">
                    <div class="fm-role-stat">
                      <span class="fm-role-stat__value">{{ selectedRoles.length }}</span>
                      <span class="fm-role-stat__label">{{ t('users_page.roles_selected') }}</span>
                    </div>
                    <div class="fm-role-stat">
                      <span class="fm-role-stat__value">{{ selectedPermissionCount }}</span>
                      <span class="fm-role-stat__label">{{ t('users_page.permissions_granted') }}</span>
                    </div>
                    <div class="fm-role-stat">
                      <span class="fm-role-stat__value">{{ selectedResourceCount }}</span>
                      <span class="fm-role-stat__label">{{ t('users_page.role_resources') }}</span>
                    </div>
                  </div>

                  <div v-if="selectedRoles.length" class="fm-role-selection-strip">
                    <div v-for="(role, idx) in selectedRoles" :key="role.id" class="fm-role-selection-pill" :style="{ '--role-accent': roleAccent(role) }">
                      <span class="fm-role-selection-pill__name">{{ role.name }}</span>
                      <span v-if="idx === 0" class="fm-role-selection-pill__tag">{{ t('users_page.primary_role') }}</span>
                    </div>
                  </div>
                  <div v-else class="fm-role-empty">
                    <i class="bi bi-shield"></i>
                    <span>{{ t('users_page.role_pick_empty') }}</span>
                  </div>

                  <div class="fm-role-card-grid">
                    <button v-for="r in allRoles" :key="r.id" type="button" class="fm-role-card" :class="{ 'is-selected': isRoleSelected(r.id) }" :style="{ '--role-accent': roleAccent(r) }" @click="toggleRole(r.id)">
                      <div class="fm-role-card__top">
                        <div>
                          <div class="fm-role-card__name">{{ r.name }}</div>
                          <div class="fm-role-card__meta">{{ isSystemRole(r) ? t('users_page.role_system') : t('users_page.role_custom') }}</div>
                        </div>
                        <span class="fm-role-card__check">
                          <i class="bi" :class="isRoleSelected(r.id) ? 'bi-check-circle-fill' : 'bi-circle'"></i>
                        </span>
                      </div>
                      <p class="fm-role-card__desc">{{ r.description || t('users_page.role_no_description') }}</p>
                      <div class="fm-role-card__stats">
                        <span>{{ rolePermissionCount(r) }} {{ t('users_page.permissions_granted') }}</span>
                        <span>{{ roleResourceCount(r) }} {{ t('users_page.role_resources') }}</span>
                      </div>
                      <div class="fm-role-card__resources">
                        <span v-for="resource in roleResourcePreview(r)" :key="resource" class="fm-role-card__resource-badge">{{ resource }}</span>
                        <span v-if="roleResourceOverflow(r) > 0" class="fm-role-card__resource-more">+{{ roleResourceOverflow(r) }}</span>
                      </div>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-primary" @click="save">{{ t('save') }}</button>
          </div>
        </div>
      </div>
    </div>

    <!-- ==================== SCOPE MODAL ==================== -->
    <div class="modal fade" id="scopeModal" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title"><i class="bi bi-shield-lock me-2"></i>{{ t('users_page.scope_title').replace('{name}', scopeUser?.display_name || scopeUser?.username || '') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="alert alert-info small mb-3">
              <i class="bi bi-info-circle me-1"></i>{{ t('users_page.scope_help') }}
            </div>
            <div class="mb-3">
              <button class="btn btn-sm btn-outline-primary me-2" @click="addScopeRow('datacenter')">{{ t('users_page.add_datacenter') }}</button>
              <button class="btn btn-sm btn-outline-info me-2" @click="addScopeRow('region')">{{ t('users_page.add_region') }}</button>
              <button class="btn btn-sm btn-outline-success" @click="addScopeRow('cluster')">{{ t('users_page.add_cluster') }}</button>
            </div>
            <div v-if="!scopeRows.length" class="text-center text-muted py-3">{{ t('users_page.no_scope_limit') }}</div>
            <table v-else class="table table-sm">
              <thead><tr><th>{{ t('users_page.scope_type') }}</th><th>{{ t('users_page.target') }}</th><th></th></tr></thead>
              <tbody>
                <tr v-for="(row, idx) in scopeRows" :key="idx">
                  <td>
                    <span class="badge" :class="{'bg-primary': row.scope_type === 'datacenter', 'bg-info': row.scope_type === 'region', 'bg-success': row.scope_type === 'cluster'}">
                      {{ { datacenter: t('topology_page.dc_label'), region: t('topology_page.region_label'), cluster: t('topology_page.cluster_label') }[row.scope_type] }}
                    </span>
                  </td>
                  <td>
                    <select v-if="row.scope_type === 'datacenter'" v-model="row.scope_id" class="form-select form-select-sm">
                      <option v-for="dc in allDCs" :key="dc.id" :value="dc.id">{{ dc.alias || dc.name }}</option>
                    </select>
                    <select v-if="row.scope_type === 'region'" v-model="row.scope_id" class="form-select form-select-sm">
                      <option v-for="r in allRegions" :key="r.id" :value="r.id">{{ r.datacenter?.alias || '' }} / {{ r.alias || r.name }}</option>
                    </select>
                    <select v-if="row.scope_type === 'cluster'" v-model="row.scope_id" class="form-select form-select-sm">
                      <option v-for="cl in allClusters" :key="cl.id" :value="cl.id">{{ cl.alias || cl.name }}</option>
                    </select>
                  </td>
                  <td><button class="btn btn-sm btn-outline-danger" @click="scopeRows.splice(idx, 1)"><i class="bi bi-trash"></i></button></td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-primary" @click="saveScopes">{{ t('users_page.save_scopes') }}</button>
          </div>
        </div>
      </div>
    </div>

    <!-- ==================== GROUP MODAL ==================== -->
    <div class="modal fade" id="groupModal" tabindex="-1">
      <div class="modal-dialog modal-lg modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ groupForm.id ? t('users_page.group_edit') : t('users_page.group_create') }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">{{ t('users_page.group_name') }}</label>
              <input v-model="groupForm.name" type="text" class="form-control" required>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('users_page.description') }}</label>
              <input v-model="groupForm.description" type="text" class="form-control">
            </div>

            <!-- Role picker -->
            <div class="mb-3">
              <label class="form-label">{{ t('users_page.roles') }}</label>
              <div class="d-flex flex-wrap gap-2">
                <button
                  v-for="r in allRoles"
                  :key="r.id"
                  type="button"
                  class="btn btn-sm"
                  :class="groupForm.role_ids.includes(r.id) ? 'btn-primary' : 'btn-outline-secondary'"
                  @click="toggleGroupRole(r.id)"
                >
                  <i class="bi bi-shield-check me-1"></i>{{ r.name }}
                </button>
              </div>
            </div>

            <!-- Scope picker -->
            <div class="mb-3">
              <label class="form-label">{{ t('users_page.scopes') }}</label>
              <div class="mb-2">
                <button class="btn btn-sm btn-outline-primary me-2" @click="addGroupScopeRow('datacenter')">{{ t('users_page.add_datacenter') }}</button>
                <button class="btn btn-sm btn-outline-info me-2" @click="addGroupScopeRow('region')">{{ t('users_page.add_region') }}</button>
                <button class="btn btn-sm btn-outline-success" @click="addGroupScopeRow('cluster')">{{ t('users_page.add_cluster') }}</button>
              </div>
              <div v-if="!groupForm.scopes.length" class="text-center text-muted py-2 small">{{ t('users_page.no_scope_limit') }}</div>
              <table v-else class="table table-sm">
                <tbody>
                  <tr v-for="(row, idx) in groupForm.scopes" :key="idx">
                    <td>
                      <span class="badge" :class="{'bg-primary': row.scope_type === 'datacenter', 'bg-info': row.scope_type === 'region', 'bg-success': row.scope_type === 'cluster'}">
                        {{ { datacenter: t('topology_page.dc_label'), region: t('topology_page.region_label'), cluster: t('topology_page.cluster_label') }[row.scope_type] }}
                      </span>
                    </td>
                    <td>
                      <select v-if="row.scope_type === 'datacenter'" v-model="row.scope_id" class="form-select form-select-sm">
                        <option v-for="dc in allDCs" :key="dc.id" :value="dc.id">{{ dc.alias || dc.name }}</option>
                      </select>
                      <select v-if="row.scope_type === 'region'" v-model="row.scope_id" class="form-select form-select-sm">
                        <option v-for="r in allRegions" :key="r.id" :value="r.id">{{ r.datacenter?.alias || '' }} / {{ r.alias || r.name }}</option>
                      </select>
                      <select v-if="row.scope_type === 'cluster'" v-model="row.scope_id" class="form-select form-select-sm">
                        <option v-for="cl in allClusters" :key="cl.id" :value="cl.id">{{ cl.alias || cl.name }}</option>
                      </select>
                    </td>
                    <td><button class="btn btn-sm btn-outline-danger" @click="groupForm.scopes.splice(idx, 1)"><i class="bi bi-trash"></i></button></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">{{ t('cancel') }}</button>
            <button type="button" class="btn btn-primary" @click="saveGroup">{{ t('save') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import {
  getUsers, createUser, updateUser, deleteUser,
  getRoles, getUserScopes, setUserScopes,
  getDataCenters, getRegions, getClusters,
  getGroups, createGroup, updateGroup, deleteGroup,
  getLDAPSettings, updateLDAPSettings, testLDAPConnection,
  getSAMLSettings, updateSAMLSettings,
  getGroupMappings, setGroupMappings,
} from '../api'
import { useI18n } from '../i18n'

const { t, dateLocale } = useI18n()
const activeTab = ref('users')

// ==================== SHARED DATA ====================
const users = ref([])
const allRoles = ref([])
const allDCs = ref([])
const allRegions = ref([])
const allClusters = ref([])
const userScopesMap = ref({})
const groups = ref([])
const systemRoleNames = ['admin', 'operator', 'viewer']

// ==================== USER STATE ====================
const form = reactive({ id: null, username: '', display_name: '', email: '', password: '', role_ids: [], group_ids: [] })
const scopeUser = ref(null)
const scopeRows = ref([])
let userModal = null
let scopeModal = null

const selectedRoles = computed(() => {
  const selected = new Set(form.role_ids)
  return allRoles.value.filter(role => selected.has(role.id))
})
const selectedPermissionCount = computed(() => {
  const permissionIDs = new Set(selectedRoles.value.flatMap(role => (role.permissions || []).map(permission => permission.id)))
  return permissionIDs.size
})
const selectedResourceCount = computed(() => {
  const resources = new Set(selectedRoles.value.flatMap(role => (role.permissions || []).map(permission => permission.resource)))
  return resources.size
})

// ==================== GROUP STATE ====================
const groupForm = reactive({ id: null, name: '', description: '', role_ids: [], scopes: [] })
let groupModal = null

// ==================== AUTH SETTINGS STATE ====================
const ldapExpanded = ref(true)
const samlExpanded = ref(false)
const ldapForm = reactive({
  enabled: false, host: '', port: 389, use_tls: false,
  bind_dn: '', bind_password: '', base_dn: '',
  user_filter: '(uid=%s)', group_filter: '',
  attributes: { username: 'uid', email: 'mail', name: 'cn' },
  group_sync_strategy: 'always',
})
const samlForm = reactive({
  enabled: false, idp_metadata: '', entity_id: '', acs_url: '',
  cert_data: '', key_data: '', group_attribute: 'memberOf',
  group_sync_strategy: 'always',
})
const ldapMappings = ref([])
const samlMappings = ref([])
const ldapTesting = ref(false)
const ldapTestResult = ref(null)

// ==================== HELPERS ====================
function formatTime(t) { return t ? new Date(t).toLocaleString(dateLocale.value) : '-' }
function scopeIcon(type) { return { datacenter: '🏢', region: '🌐', cluster: '⚙️' }[type] || '' }
function roleAccent(role) {
  const palette = { admin: '#ef4444', operator: '#2563eb', viewer: '#475569' }
  return palette[role?.name] || '#0f766e'
}
function isSystemRole(role) { return systemRoleNames.includes(role?.name) }
function rolePermissionCount(role) { return (role.permissions || []).length }
function roleResources(role) { return [...new Set((role.permissions || []).map(permission => permission.resource))] }
function roleResourceCount(role) { return roleResources(role).length }
function roleResourcePreview(role) { return roleResources(role).slice(0, 3) }
function roleResourceOverflow(role) { return Math.max(roleResourceCount(role) - 3, 0) }
function isRoleSelected(roleID) { return form.role_ids.includes(roleID) }
function toggleRole(roleID) {
  if (isRoleSelected(roleID)) { form.role_ids = form.role_ids.filter(id => id !== roleID); return }
  form.role_ids = [...form.role_ids, roleID]
}
function clearRoles() { form.role_ids = [] }
function toggleUserGroup(groupID) {
  if (form.group_ids.includes(groupID)) { form.group_ids = form.group_ids.filter(id => id !== groupID); return }
  form.group_ids = [...form.group_ids, groupID]
}

// ==================== MODALS ====================
function getUserModal() {
  if (!userModal) userModal = new window.bootstrap.Modal(document.getElementById('userModal'))
  return userModal
}
function getScopeModal() {
  if (!scopeModal) scopeModal = new window.bootstrap.Modal(document.getElementById('scopeModal'))
  return scopeModal
}
function getGroupModal() {
  if (!groupModal) groupModal = new window.bootstrap.Modal(document.getElementById('groupModal'))
  return groupModal
}

// ==================== USER CRUD ====================
async function loadUsers() {
  const { data } = await getUsers({ page_size: 100 })
  users.value = data.data || []
  for (const u of users.value) {
    try {
      const { data: scopeData } = await getUserScopes(u.id)
      userScopesMap.value[u.id] = scopeData.data || []
    } catch { userScopesMap.value[u.id] = [] }
  }
}

function openCreate() {
  Object.assign(form, { id: null, username: '', display_name: '', email: '', password: '', role_ids: [], group_ids: [] })
  getUserModal().show()
}
function openEdit(u) {
  Object.assign(form, {
    id: u.id, username: u.username, display_name: u.display_name,
    email: u.email, password: '', role_ids: (u.roles || []).map(r => r.id),
    group_ids: (u.groups || []).map(g => g.id),
  })
  getUserModal().show()
}

async function save() {
  const data = { ...form }
  if (!data.password) delete data.password
  if (form.id) await updateUser(form.id, data)
  else await createUser(data)
  getUserModal().hide()
  loadUsers()
}

async function handleDelete(u) {
  if (confirm(t('users_page.confirm_delete').replace('{name}', u.username))) {
    await deleteUser(u.id)
    loadUsers()
  }
}

function openScopeEdit(u) {
  scopeUser.value = u
  const existing = userScopesMap.value[u.id] || []
  scopeRows.value = existing.map(s => ({ scope_type: s.scope_type, scope_id: s.scope_id }))
  getScopeModal().show()
}
function addScopeRow(type) {
  let defaultID = null
  if (type === 'datacenter') defaultID = allDCs.value[0]?.id
  if (type === 'region') defaultID = allRegions.value[0]?.id
  if (type === 'cluster') defaultID = allClusters.value[0]?.id
  scopeRows.value.push({ scope_type: type, scope_id: defaultID })
}
async function saveScopes() {
  await setUserScopes(scopeUser.value.id, { scopes: scopeRows.value })
  getScopeModal().hide()
  loadUsers()
}

// ==================== GROUP CRUD ====================
async function loadGroups() {
  try {
    const { data } = await getGroups({ page_size: 100 })
    groups.value = data.data || []
  } catch { groups.value = [] }
}

function openGroupCreate() {
  Object.assign(groupForm, { id: null, name: '', description: '', role_ids: [], scopes: [] })
  getGroupModal().show()
}
function openGroupEdit(g) {
  Object.assign(groupForm, {
    id: g.id, name: g.name, description: g.description,
    role_ids: (g.roles || []).map(r => r.id),
    scopes: (g.scopes || []).map(s => ({ scope_type: s.scope_type, scope_id: s.scope_id })),
  })
  getGroupModal().show()
}
function toggleGroupRole(roleID) {
  if (groupForm.role_ids.includes(roleID)) { groupForm.role_ids = groupForm.role_ids.filter(id => id !== roleID); return }
  groupForm.role_ids = [...groupForm.role_ids, roleID]
}
function addGroupScopeRow(type) {
  let defaultID = null
  if (type === 'datacenter') defaultID = allDCs.value[0]?.id
  if (type === 'region') defaultID = allRegions.value[0]?.id
  if (type === 'cluster') defaultID = allClusters.value[0]?.id
  groupForm.scopes.push({ scope_type: type, scope_id: defaultID })
}
async function saveGroup() {
  const data = { name: groupForm.name, description: groupForm.description, role_ids: groupForm.role_ids, scopes: groupForm.scopes }
  if (groupForm.id) await updateGroup(groupForm.id, data)
  else await createGroup(data)
  getGroupModal().hide()
  loadGroups()
  loadUsers()
}
async function handleGroupDelete(g) {
  if (confirm(t('users_page.group_confirm_delete').replace('{name}', g.name))) {
    await deleteGroup(g.id)
    loadGroups()
  }
}

// ==================== AUTH SETTINGS ====================
async function loadAuthSettings() {
  try {
    const { data: ldap } = await getLDAPSettings()
    Object.assign(ldapForm, ldap)
    if (!ldapForm.attributes) ldapForm.attributes = { username: 'uid', email: 'mail', name: 'cn' }
    if (!ldapForm.group_sync_strategy) ldapForm.group_sync_strategy = 'always'
  } catch {}
  try {
    const { data: saml } = await getSAMLSettings()
    Object.assign(samlForm, saml)
    if (!samlForm.group_sync_strategy) samlForm.group_sync_strategy = 'always'
  } catch {}
  try {
    const { data: ldapM } = await getGroupMappings('ldap')
    ldapMappings.value = (ldapM.data || []).map(m => ({ external_group_name: m.external_group_name, group_id: m.group_id }))
  } catch { ldapMappings.value = [] }
  try {
    const { data: samlM } = await getGroupMappings('saml')
    samlMappings.value = (samlM.data || []).map(m => ({ external_group_name: m.external_group_name, group_id: m.group_id }))
  } catch { samlMappings.value = [] }
}

async function saveLDAP() {
  await updateLDAPSettings({ ...ldapForm })
  await setGroupMappings({ source: 'ldap', mappings: ldapMappings.value.filter(m => m.external_group_name && m.group_id) })
  alert(t('auth_settings.save_success'))
}
async function saveSAML() {
  await updateSAMLSettings({ ...samlForm })
  await setGroupMappings({ source: 'saml', mappings: samlMappings.value.filter(m => m.external_group_name && m.group_id) })
  alert(t('auth_settings.save_success'))
}
async function testLDAP() {
  ldapTesting.value = true
  ldapTestResult.value = null
  try {
    const { data } = await testLDAPConnection({ ...ldapForm })
    ldapTestResult.value = data
  } catch (e) {
    ldapTestResult.value = { success: false, error: e.message || 'Connection failed' }
  }
  ldapTesting.value = false
}

// ==================== MOUNT ====================
onMounted(async () => {
  const [, rolesRes, dcRes, regRes, clRes] = await Promise.all([
    loadUsers(), getRoles(), getDataCenters(), getRegions(), getClusters(),
  ])
  allRoles.value = rolesRes.data.data || []
  allDCs.value = dcRes.data.data || []
  allRegions.value = regRes.data.data || []
  allClusters.value = clRes.data.data || []
  await loadGroups()
  loadAuthSettings()
})
</script>

<style scoped>
.bg-purple { background-color: #7c3aed !important; }

.fm-role-chip-list { display: flex; flex-wrap: wrap; gap: 0.4rem; }

.fm-role-chip {
  display: inline-flex; align-items: center; gap: 0.35rem;
  padding: 0.4rem 0.7rem; border-radius: 999px;
  background: color-mix(in srgb, var(--role-accent) 10%, white);
  color: color-mix(in srgb, var(--role-accent) 72%, #0f172a);
  border: 1px solid color-mix(in srgb, var(--role-accent) 18%, white);
  font-size: 0.77rem; font-weight: 600;
}

.fm-role-panel {
  border: 1px solid #dbe6f3; border-radius: 18px;
  background: radial-gradient(circle at top right, rgba(59, 130, 246, 0.12), transparent 34%), linear-gradient(180deg, #f8fbff 0%, #ffffff 100%);
  padding: 1rem;
}
.fm-role-panel__header { display: flex; justify-content: space-between; gap: 1rem; align-items: flex-start; margin-bottom: 1rem; }
.fm-role-panel__eyebrow { font-size: 0.72rem; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; color: #2563eb; margin-bottom: 0.35rem; }
.fm-role-panel__title { font-weight: 700; color: #0f172a; }
.fm-role-summary { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.75rem; margin-bottom: 1rem; }
.fm-role-stat { padding: 0.85rem 0.9rem; border-radius: 14px; background: rgba(255, 255, 255, 0.88); border: 1px solid rgba(148, 163, 184, 0.18); }
.fm-role-stat__value { display: block; font-size: 1.2rem; font-weight: 700; color: #0f172a; }
.fm-role-stat__label { display: block; margin-top: 0.2rem; font-size: 0.76rem; color: #64748b; text-transform: uppercase; letter-spacing: 0.05em; }
.fm-role-selection-strip { display: flex; flex-wrap: wrap; gap: 0.5rem; margin-bottom: 1rem; }
.fm-role-selection-pill { display: inline-flex; align-items: center; gap: 0.45rem; padding: 0.5rem 0.7rem; border-radius: 999px; background: color-mix(in srgb, var(--role-accent) 12%, white); border: 1px solid color-mix(in srgb, var(--role-accent) 18%, white); color: #0f172a; }
.fm-role-selection-pill__name { font-weight: 700; font-size: 0.82rem; }
.fm-role-selection-pill__tag { padding: 0.12rem 0.45rem; border-radius: 999px; background: rgba(255, 255, 255, 0.85); font-size: 0.68rem; color: #475569; }
.fm-role-empty { display: flex; align-items: center; gap: 0.55rem; padding: 0.8rem 0.9rem; margin-bottom: 1rem; border-radius: 14px; background: rgba(255, 255, 255, 0.78); border: 1px dashed rgba(148, 163, 184, 0.4); color: #64748b; }
.fm-role-card-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.75rem; }
.fm-role-card { text-align: left; border: 1px solid rgba(148, 163, 184, 0.22); border-radius: 16px; background: rgba(255, 255, 255, 0.9); padding: 0.95rem; transition: transform 0.18s ease, box-shadow 0.18s ease, border-color 0.18s ease, background 0.18s ease; }
.fm-role-card:hover { transform: translateY(-1px); box-shadow: 0 12px 24px rgba(15, 23, 42, 0.08); border-color: color-mix(in srgb, var(--role-accent) 28%, white); }
.fm-role-card.is-selected { border-color: color-mix(in srgb, var(--role-accent) 40%, white); background: linear-gradient(180deg, color-mix(in srgb, var(--role-accent) 10%, white) 0%, rgba(255, 255, 255, 0.96) 100%); box-shadow: 0 14px 32px color-mix(in srgb, var(--role-accent) 14%, transparent); }
.fm-role-card__top { display: flex; justify-content: space-between; gap: 0.75rem; }
.fm-role-card__name { font-size: 0.98rem; font-weight: 700; color: #0f172a; }
.fm-role-card__meta { font-size: 0.72rem; color: #64748b; text-transform: uppercase; letter-spacing: 0.05em; margin-top: 0.18rem; }
.fm-role-card__check { color: var(--role-accent); font-size: 1.1rem; }
.fm-role-card__desc { color: #64748b; font-size: 0.82rem; line-height: 1.45; min-height: 2.35rem; margin: 0.75rem 0; }
.fm-role-card__stats { display: flex; gap: 0.75rem; flex-wrap: wrap; margin-bottom: 0.7rem; font-size: 0.76rem; color: #475569; }
.fm-role-card__resources { display: flex; flex-wrap: wrap; gap: 0.35rem; }
.fm-role-card__resource-badge, .fm-role-card__resource-more { display: inline-flex; align-items: center; padding: 0.22rem 0.5rem; border-radius: 999px; background: rgba(241, 245, 249, 0.9); color: #334155; font-size: 0.72rem; }

@media (max-width: 991.98px) { .fm-role-card-grid { grid-template-columns: 1fr; } }
@media (max-width: 767.98px) { .fm-role-summary { grid-template-columns: 1fr; } }
</style>
