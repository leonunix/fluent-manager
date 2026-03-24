package models

import "gorm.io/gorm"

type builtinConfigModuleSeed struct {
	Name        string
	Description string
	ModuleType  string
	FluentType  string
	Content     string
	Variables   string
	PresetKind  string
	PresetKey   string
}

func seedBuiltinConfigModules(db *gorm.DB) {
	modules := []builtinConfigModuleSeed{
		{
			Name:        "guided-fb-service-standard",
			Description: "Enterprise baseline service tuning for Fluent Bit edge collectors.",
			ModuleType:  "service",
			FluentType:  "fluentbit",
			Variables:   `{"flush":1,"log_level":"info","storage_path":"/var/lib/fluent-bit/state"}`,
			Content: `[SERVICE]
    Flush         {{ .flush }}
    Daemon        Off
    Log_Level     {{ .log_level }}
    storage.path  {{ .storage_path }}`,
		},
		{
			Name:        "guided-fb-filter-enrich",
			Description: "Add standard enterprise metadata before forwarding to the central pipeline.",
			ModuleType:  "filter",
			FluentType:  "fluentbit",
			Variables:   `{"match":"*","source_kind":"edge","pipeline_owner":"platform"}`,
			Content: `[FILTER]
    Name   modify
    Match  {{ .match }}
    Add    source_kind {{ .source_kind }}
    Add    pipeline_owner {{ .pipeline_owner }}`,
		},
		{
			Name:        "guided-fb-input-nginx-access",
			Description: "Tail Nginx access logs with a ready-to-use enterprise preset.",
			ModuleType:  "input",
			FluentType:  "fluentbit",
			Variables:   `{"path":"/var/log/nginx/access.log","tag":"nginx.access","db_path":"/var/lib/fluent-bit/nginx-access.db","parser":"nginx"}`,
			PresetKind:  "input",
			PresetKey:   "nginx_access",
			Content: `[INPUT]
    Name              tail
    Path              {{ .path }}
    Tag               {{ .tag }}
    DB                {{ .db_path }}
    Parser            {{ .parser }}
    Refresh_Interval  5`,
		},
		{
			Name:        "guided-fb-input-nginx-error",
			Description: "Tail Nginx error logs with multiline-friendly defaults.",
			ModuleType:  "input",
			FluentType:  "fluentbit",
			Variables:   `{"path":"/var/log/nginx/error.log","tag":"nginx.error","db_path":"/var/lib/fluent-bit/nginx-error.db"}`,
			PresetKind:  "input",
			PresetKey:   "nginx_error",
			Content: `[INPUT]
    Name              tail
    Path              {{ .path }}
    Tag               {{ .tag }}
    DB                {{ .db_path }}
    Read_from_Head    True
    Refresh_Interval  5`,
		},
		{
			Name:        "guided-fb-input-systemd",
			Description: "Collect journald entries from systemd-managed services.",
			ModuleType:  "input",
			FluentType:  "fluentbit",
			Variables:   `{"tag":"host.systemd","systemd_filter":"_SYSTEMD_UNIT=nginx.service","read_from_tail":"On"}`,
			PresetKind:  "input",
			PresetKey:   "systemd_journald",
			Content: `[INPUT]
    Name                systemd
    Tag                 {{ .tag }}
    Systemd_Filter      {{ .systemd_filter }}
    Read_From_Tail      {{ .read_from_tail }}`,
		},
		{
			Name:        "guided-fb-input-docker-json",
			Description: "Collect Docker JSON logs from the host runtime.",
			ModuleType:  "input",
			FluentType:  "fluentbit",
			Variables:   `{"path":"/var/lib/docker/containers/*/*-json.log","tag":"docker.containers","db_path":"/var/lib/fluent-bit/docker-json.db","parser":"docker"}`,
			PresetKind:  "input",
			PresetKey:   "docker_json",
			Content: `[INPUT]
    Name              tail
    Path              {{ .path }}
    Tag               {{ .tag }}
    DB                {{ .db_path }}
    Parser            {{ .parser }}
    Docker_Mode       On`,
		},
		{
			Name:        "guided-fb-output-opensearch",
			Description: "Send logs to an OpenSearch cluster with reusable destination variables.",
			ModuleType:  "output",
			FluentType:  "fluentbit",
			Variables:   `{"match":"*","host":"opensearch.internal","port":9200,"index":"logs-%Y.%m.%d","http_user":"admin","http_password":"changeme","tls":"On","replace_dots":"On"}`,
			PresetKind:  "output",
			PresetKey:   "opensearch",
			Content: `[OUTPUT]
    Name           opensearch
    Match          {{ .match }}
    Host           {{ .host }}
    Port           {{ .port }}
    Index          {{ .index }}
    HTTP_User      {{ .http_user }}
    HTTP_Passwd    {{ .http_password }}
    tls            {{ .tls }}
    Replace_Dots   {{ .replace_dots }}`,
		},
		{
			Name:        "guided-fb-output-loki",
			Description: "Forward logs to Grafana Loki using the HTTP output plugin.",
			ModuleType:  "output",
			FluentType:  "fluentbit",
			Variables:   `{"match":"*","host":"loki.internal","port":3100,"tenant_id":"platform","labels":"job=fluent-manager"}`,
			PresetKind:  "output",
			PresetKey:   "loki",
			Content: `[OUTPUT]
    Name       loki
    Match      {{ .match }}
    Host       {{ .host }}
    Port       {{ .port }}
    Tenant_ID  {{ .tenant_id }}
    Labels     {{ .labels }}`,
		},
		{
			Name:        "guided-fb-output-kafka",
			Description: "Forward logs into Kafka for downstream stream processing.",
			ModuleType:  "output",
			FluentType:  "fluentbit",
			Variables:   `{"match":"*","brokers":"kafka-1:9092,kafka-2:9092","topics":"logs.platform","format":"json"}`,
			PresetKind:  "output",
			PresetKey:   "kafka",
			Content: `[OUTPUT]
    Name      kafka
    Match     {{ .match }}
    Brokers   {{ .brokers }}
    Topics    {{ .topics }}
    Format    {{ .format }}`,
		},
		{
			Name:        "guided-fb-output-http",
			Description: "Deliver logs to a generic HTTP endpoint.",
			ModuleType:  "output",
			FluentType:  "fluentbit",
			Variables:   `{"match":"*","host":"collector.internal","port":8080,"uri":"/ingest","format":"json_lines"}`,
			PresetKind:  "output",
			PresetKey:   "http",
			Content: `[OUTPUT]
    Name     http
    Match    {{ .match }}
    Host     {{ .host }}
    Port     {{ .port }}
    URI      {{ .uri }}
    Format   {{ .format }}`,
		},
		{
			Name:        "guided-fd-service-standard",
			Description: "Enterprise baseline workers and log level for Fluentd aggregators.",
			ModuleType:  "service",
			FluentType:  "fluentd",
			Variables:   `{"workers":2,"log_level":"info"}`,
			Content: `<system>
  workers {{ .workers }}
  log_level {{ .log_level }}
</system>`,
		},
		{
			Name:        "guided-fd-input-nginx-access",
			Description: "Tail Nginx access logs in Fluentd with a ready-made preset.",
			ModuleType:  "input",
			FluentType:  "fluentd",
			Variables:   `{"path":"/var/log/nginx/access.log","pos_file":"/var/log/fluentd/nginx-access.pos","tag":"nginx.access"}`,
			PresetKind:  "input",
			PresetKey:   "nginx_access",
			Content: `<source>
  @type tail
  path {{ .path }}
  pos_file {{ .pos_file }}
  tag {{ .tag }}
  <parse>
    @type nginx
  </parse>
</source>`,
		},
		{
			Name:        "guided-fd-input-systemd",
			Description: "Collect journald logs in Fluentd for host and service observability.",
			ModuleType:  "input",
			FluentType:  "fluentd",
			Variables:   `{"tag":"host.systemd","path":"/var/log/journal","matches":"[{\"_SYSTEMD_UNIT\":\"nginx.service\"}]"}`,
			PresetKind:  "input",
			PresetKey:   "systemd_journald",
			Content: `<source>
  @type systemd
  tag {{ .tag }}
  path {{ .path }}
  matches {{ .matches }}
</source>`,
		},
		{
			Name:        "guided-fd-output-opensearch",
			Description: "Send Fluentd events to OpenSearch with reusable destination settings.",
			ModuleType:  "output",
			FluentType:  "fluentd",
			Variables:   `{"match":"**","host":"opensearch.internal","port":9200,"scheme":"https","index_name":"logs","user":"admin","password":"changeme"}`,
			PresetKind:  "output",
			PresetKey:   "opensearch",
			Content: `<match {{ .match }}>
  @type opensearch
  host {{ .host }}
  port {{ .port }}
  scheme {{ .scheme }}
  index_name {{ .index_name }}
  user {{ .user }}
  password {{ .password }}
</match>`,
		},
		{
			Name:        "guided-fd-output-loki",
			Description: "Ship Fluentd events to Loki.",
			ModuleType:  "output",
			FluentType:  "fluentd",
			Variables:   `{"match":"**","url":"http://loki.internal:3100","tenant":"platform","extra_labels":"{\"job\":\"fluent-manager\"}"}`,
			PresetKind:  "output",
			PresetKey:   "loki",
			Content: `<match {{ .match }}>
  @type loki
  url {{ .url }}
  tenant {{ .tenant }}
  extra_labels {{ .extra_labels }}
</match>`,
		},
		{
			Name:        "guided-fd-output-kafka",
			Description: "Stream Fluentd events into Kafka.",
			ModuleType:  "output",
			FluentType:  "fluentd",
			Variables:   `{"match":"**","brokers":"kafka-1:9092,kafka-2:9092","default_topic":"logs.platform","output_data_type":"json"}`,
			PresetKind:  "output",
			PresetKey:   "kafka",
			Content: `<match {{ .match }}>
  @type kafka2
  brokers {{ .brokers }}
  default_topic {{ .default_topic }}
  output_data_type {{ .output_data_type }}
</match>`,
		},
		{
			Name:        "guided-fd-output-http",
			Description: "Deliver Fluentd events to a generic HTTP endpoint.",
			ModuleType:  "output",
			FluentType:  "fluentd",
			Variables:   `{"match":"**","endpoint":"http://collector.internal:8080/ingest","http_method":"post","serializer":"json"}`,
			PresetKind:  "output",
			PresetKey:   "http",
			Content: `<match {{ .match }}>
  @type http
  endpoint {{ .endpoint }}
  http_method {{ .http_method }}
  serializer {{ .serializer }}
</match>`,
		},
	}

	for _, seed := range modules {
		var existing ConfigModule
		err := db.Where("name = ? AND module_type = ? AND fluent_type = ?", seed.Name, seed.ModuleType, seed.FluentType).First(&existing).Error
		if err == nil {
			continue
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			continue
		}
		db.Create(&ConfigModule{
			Name:        seed.Name,
			Description: seed.Description,
			ModuleType:  seed.ModuleType,
			FluentType:  seed.FluentType,
			Content:     seed.Content,
			Variables:   seed.Variables,
			IsBuiltin:   true,
			PresetKind:  seed.PresetKind,
			PresetKey:   seed.PresetKey,
		})
	}
}
