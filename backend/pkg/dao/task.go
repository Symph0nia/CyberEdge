package dao

import (
	"cyberedge/pkg/models"
	"gorm.io/gorm"
)

// TaskDAO 任务数据访问对象
type TaskDAO struct {
	*BaseDAO
}

// NewTaskDAO 创建任务DAO
func NewTaskDAO(db *gorm.DB) *TaskDAO {
	return &TaskDAO{
		BaseDAO: NewBaseDAO(db),
	}
}

// Create 创建任务
func (d *TaskDAO) Create(task *models.Task) error {
	return d.db.Create(task).Error
}

// GetByID 根据ID获取任务
func (d *TaskDAO) GetByID(id uint) (*models.Task, error) {
	var task models.Task
	err := d.db.Preload("Target").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Update 更新任务
func (d *TaskDAO) Update(id uint, updates map[string]interface{}) error {
	return d.db.Model(&models.Task{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateStatus 更新任务状态
func (d *TaskDAO) UpdateStatus(id uint, status models.TaskStatus, result string) error {
	updates := map[string]interface{}{
		"status": status,
		"result": result,
	}

	if status == models.TaskStatusCompleted || status == models.TaskStatusFailed {
		// TODO: 设置实际的时间戳
		updates["completed_at"] = int64(1000)
	}

	return d.Update(id, updates)
}

// GetAll 获取所有任务
func (d *TaskDAO) GetAll() ([]*models.Task, error) {
	var tasks []*models.Task
	err := d.db.Preload("Target").Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetByTargetID 根据目标ID获取任务
func (d *TaskDAO) GetByTargetID(targetID uint) ([]*models.Task, error) {
	var tasks []*models.Task
	err := d.db.Where("target_id = ?", targetID).
		Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetByType 根据类型获取任务
func (d *TaskDAO) GetByType(taskType models.TaskType) ([]*models.Task, error) {
	var tasks []*models.Task
	err := d.db.Where("type = ?", taskType).
		Preload("Target").
		Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetByStatus 根据状态获取任务
func (d *TaskDAO) GetByStatus(status models.TaskStatus) ([]*models.Task, error) {
	var tasks []*models.Task
	err := d.db.Where("status = ?", status).
		Preload("Target").
		Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetRunningTasks 获取运行中的任务
func (d *TaskDAO) GetRunningTasks() ([]*models.Task, error) {
	return d.GetByStatus(models.TaskStatusRunning)
}

// GetRecentTasks 获取最近的任务
func (d *TaskDAO) GetRecentTasks(limit int) ([]*models.Task, error) {
	var tasks []*models.Task
	err := d.db.Preload("Target").
		Order("created_at DESC").
		Limit(limit).Find(&tasks).Error
	return tasks, err
}

// Delete 删除任务
func (d *TaskDAO) Delete(id uint) error {
	return d.db.Delete(&models.Task{}, id).Error
}

// DeleteByTargetID 删除目标相关的所有任务
func (d *TaskDAO) DeleteByTargetID(targetID uint) error {
	return d.db.Where("target_id = ?", targetID).Delete(&models.Task{}).Error
}

// Count 获取任务总数
func (d *TaskDAO) Count() (int64, error) {
	var count int64
	err := d.db.Model(&models.Task{}).Count(&count).Error
	return count, err
}

// CountByStatus 按状态统计任务数量
func (d *TaskDAO) CountByStatus() (map[models.TaskStatus]int64, error) {
	type StatusCount struct {
		Status models.TaskStatus `json:"status"`
		Count  int64             `json:"count"`
	}

	var results []StatusCount
	err := d.db.Model(&models.Task{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[models.TaskStatus]int64)
	for _, result := range results {
		counts[result.Status] = result.Count
	}

	return counts, nil
}