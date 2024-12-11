package service

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"errors"
)

type TargetService struct {
	targetDAO *dao.TargetDAO
}

// NewTargetService 创建一个新的 TargetService 实例
func NewTargetService(targetDAO *dao.TargetDAO) *TargetService {
	return &TargetService{
		targetDAO: targetDAO,
	}
}

// CreateTarget 创建新的目标
func (s *TargetService) CreateTarget(target *models.Target) error {
	if target == nil {
		return errors.New("无效的目标")
	}

	logging.Info("正在通过 Service 创建新的目标")
	err := s.targetDAO.CreateTarget(target)
	if err != nil {
		logging.Error("通过 Service 创建目标失败: %v", err)
		return err
	}

	logging.Info("通过 Service 成功创建目标")
	return nil
}

// GetTargetByID 根据 ID 获取目标
func (s *TargetService) GetTargetByID(id string) (*models.Target, error) {
	if id == "" {
		return nil, errors.New("无效的 ID")
	}

	logging.Info("正在通过 Service 获取目标: %s", id)
	target, err := s.targetDAO.GetTargetByID(id)
	if err != nil {
		logging.Error("通过 Service 获取目标失败: %v", err)
		return nil, err
	}

	logging.Info("通过 Service 成功获取目标: %s", id)
	return target, nil
}

// GetAllTargets 获取所有目标
func (s *TargetService) GetAllTargets() ([]*models.Target, error) {
	logging.Info("正在通过 Service 获取所有目标")
	targets, err := s.targetDAO.GetAllTargets()
	if err != nil {
		logging.Error("通过 Service 获取目标列表失败: %v", err)
		return nil, err
	}

	logging.Info("通过 Service 成功获取目标列表，共 %d 个", len(targets))
	return targets, nil
}

// UpdateTarget 更新目标信息
func (s *TargetService) UpdateTarget(id string, updatedTarget *models.Target) error {
	if id == "" || updatedTarget == nil {
		return errors.New("无效的参数")
	}

	logging.Info("正在通过 Service 更新目标: %s", id)
	err := s.targetDAO.UpdateTarget(id, updatedTarget)
	if err != nil {
		logging.Error("通过 Service 更新目标失败: %v", err)
		return err
	}

	logging.Info("通过 Service 成功更新目标: %s", id)
	return nil
}

// DeleteTarget 删除目标
func (s *TargetService) DeleteTarget(id string) error {
	if id == "" {
		return errors.New("无效的 ID")
	}

	logging.Info("正在通过 Service 删除目标: %s", id)
	err := s.targetDAO.DeleteTarget(id)
	if err != nil {
		logging.Error("通过 Service 删除目标失败: %v", err)
		return err
	}

	logging.Info("通过 Service 成功删除目标: %s", id)
	return nil
}
