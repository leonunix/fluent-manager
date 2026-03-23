package services

import (
	"github.com/fluent-manager/fluent-manager/internal/cache"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

type OverviewResult struct {
	AvgCPU        float64 `json:"avg_cpu"`
	AvgMem        float64 `json:"avg_mem"`
	AvgDisk       float64 `json:"avg_disk"`
	FluentRunning int64   `json:"fluent_running"`
	FluentTotal   int64   `json:"fluent_total"`
	FluentRunRate float64 `json:"fluent_run_rate"`
	TotalNodes    int64   `json:"total_nodes"`
}

type TopNodeResult struct {
	NodeID      uint    `json:"node_id"`
	Hostname    string  `json:"hostname"`
	IPAddress   string  `json:"ip_address"`
	ClusterName string  `json:"cluster_name"`
	CPU         float64 `json:"cpu"`
	Mem         float64 `json:"mem"`
	Disk        float64 `json:"disk"`
}

type DCMetricsResult struct {
	DCID      uint    `json:"dc_id"`
	DCName    string  `json:"dc_name"`
	DCAlias   string  `json:"dc_alias"`
	AvgCPU    float64 `json:"avg_cpu"`
	AvgMem    float64 `json:"avg_mem"`
	NodeCount int64   `json:"node_count"`
}

type MetricsService interface {
	Overview() (*OverviewResult, error)
	TopNodes() ([]TopNodeResult, error)
	ByDatacenter() ([]DCMetricsResult, error)
}

type metricsService struct {
	db *gorm.DB
}

func NewMetricsService(db *gorm.DB) MetricsService {
	return &metricsService{db: db}
}

func (s *metricsService) Overview() (*OverviewResult, error) {
	const cacheKey = "metrics:overview"
	var resp OverviewResult
	if cache.Get(cacheKey, &resp) {
		return &resp, nil
	}

	var total int64
	s.db.Model(&models.NodeMetrics{}).
		Joins("JOIN nodes ON nodes.id = node_metrics.node_id AND nodes.status = 'online' AND nodes.deleted_at IS NULL").
		Count(&total)

	if total > 0 {
		row := s.db.Model(&models.NodeMetrics{}).
			Select("AVG(node_metrics.cpu_usage_percent) as avg_cpu, AVG(node_metrics.mem_usage_percent) as avg_mem, AVG(node_metrics.disk_usage_percent) as avg_disk").
			Joins("JOIN nodes ON nodes.id = node_metrics.node_id AND nodes.status = 'online' AND nodes.deleted_at IS NULL").
			Row()
		row.Scan(&resp.AvgCPU, &resp.AvgMem, &resp.AvgDisk)

		var running int64
		s.db.Model(&models.NodeMetrics{}).
			Joins("JOIN nodes ON nodes.id = node_metrics.node_id AND nodes.status = 'online' AND nodes.deleted_at IS NULL").
			Where("node_metrics.fluent_running = ?", true).
			Count(&running)

		resp.FluentRunning = running
		resp.FluentTotal = total
		resp.FluentRunRate = float64(running) / float64(total) * 100
	}
	resp.TotalNodes = total

	cache.Set(cacheKey, resp)
	return &resp, nil
}

func (s *metricsService) TopNodes() ([]TopNodeResult, error) {
	const cacheKey = "metrics:top_nodes"
	var resp []TopNodeResult
	if cache.Get(cacheKey, &resp) {
		return resp, nil
	}

	s.db.Model(&models.NodeMetrics{}).
		Select("node_metrics.node_id, nodes.hostname, nodes.ip_address, COALESCE(clusters.name, '') as cluster_name, node_metrics.cpu_usage_percent as cpu, node_metrics.mem_usage_percent as mem, node_metrics.disk_usage_percent as disk").
		Joins("JOIN nodes ON nodes.id = node_metrics.node_id AND nodes.status = 'online' AND nodes.deleted_at IS NULL").
		Joins("LEFT JOIN clusters ON clusters.id = nodes.cluster_id AND clusters.deleted_at IS NULL").
		Order("node_metrics.cpu_usage_percent DESC").
		Limit(5).
		Scan(&resp)

	if resp == nil {
		resp = []TopNodeResult{}
	}

	cache.Set(cacheKey, resp)
	return resp, nil
}

func (s *metricsService) ByDatacenter() ([]DCMetricsResult, error) {
	const cacheKey = "metrics:by_dc"
	var resp []DCMetricsResult
	if cache.Get(cacheKey, &resp) {
		return resp, nil
	}

	s.db.Model(&models.NodeMetrics{}).
		Select("data_centers.id as dc_id, data_centers.name as dc_name, data_centers.alias as dc_alias, AVG(node_metrics.cpu_usage_percent) as avg_cpu, AVG(node_metrics.mem_usage_percent) as avg_mem, COUNT(*) as node_count").
		Joins("JOIN nodes ON nodes.id = node_metrics.node_id AND nodes.status = 'online' AND nodes.deleted_at IS NULL").
		Joins("JOIN clusters ON clusters.id = nodes.cluster_id AND clusters.deleted_at IS NULL").
		Joins("JOIN regions ON regions.id = clusters.region_id AND regions.deleted_at IS NULL").
		Joins("JOIN data_centers ON data_centers.id = regions.data_center_id AND data_centers.deleted_at IS NULL").
		Group("data_centers.id").
		Order("data_centers.name").
		Scan(&resp)

	if resp == nil {
		resp = []DCMetricsResult{}
	}

	cache.Set(cacheKey, resp)
	return resp, nil
}
