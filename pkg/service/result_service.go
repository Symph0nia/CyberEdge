package service

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResultService struct {
	resultDAO *dao.ResultDAO
}

// NewResultService 创建一个新的 ResultService 实例
func NewResultService(resultDAO *dao.ResultDAO) *ResultService {
	return &ResultService{
		resultDAO: resultDAO,
	}
}

// CreateResult 创建新的扫描结果
func (s *ResultService) CreateResult(result *models.Result) error {
	if result == nil {
		return errors.New("无效的扫描结果")
	}

	logging.Info("正在通过 Service 创建新的扫描结果")
	err := s.resultDAO.CreateResult(result)
	if err != nil {
		logging.Error("通过 Service 创建扫描结果失败: %v", err)
		return err
	}

	logging.Info("通过 Service 成功创建扫描结果")
	return nil
}

// GetResultByID 根据 ID 获取扫描结果
func (s *ResultService) GetResultByID(id string) (*models.Result, error) {
	if id == "" {
		return nil, errors.New("无效的 ID")
	}

	logging.Info("正在通过 Service 获取扫描结果: %s", id)
	result, err := s.resultDAO.GetResultByID(id)
	if err != nil {
		logging.Error("通过 Service 获取扫描结果失败: %v", err)
		return nil, err
	}

	logging.Info("通过 Service 成功获取扫描结果: %s", id)
	return result, nil
}

// GetResultsByType 根据类型获取扫描结果列表
func (s *ResultService) GetResultsByType(resultType string) ([]*models.Result, error) {
	if resultType == "" {
		return nil, errors.New("无效的类型")
	}

	logging.Info("正在通过 Service 获取类型为 %s 的扫描结果", resultType)
	results, err := s.resultDAO.GetResultsByType(resultType)
	if err != nil {
		logging.Error("通过 Service 获取类型为 %s 的扫描结果失败: %v", resultType, err)
		return nil, err
	}

	logging.Info("通过 Service 成功获取类型为 %s 的扫描结果，共 %d 个", resultType, len(results))
	return results, nil
}

// UpdateResult 更新指定 ID 的扫描结果
func (s *ResultService) UpdateResult(id string, updatedResult *models.Result) error {
	if id == "" || updatedResult == nil {
		return errors.New("无效的参数")
	}

	logging.Info("正在通过 Service 更新扫描结果: %s", id)
	err := s.resultDAO.UpdateResult(id, updatedResult)
	if err != nil {
		logging.Error("通过 Service 更新扫描结果失败: %v", err)
		return err
	}

	logging.Info("通过 Service 成功更新扫描结果: %s", id)
	return nil
}

// DeleteResult 删除指定 ID 的扫描结果
func (s *ResultService) DeleteResult(id string) error {
	if id == "" {
		return errors.New("无效的 ID")
	}

	logging.Info("正在通过 Service 删除扫描结果: %s", id)
	err := s.resultDAO.DeleteResult(id)
	if err != nil {
		logging.Error("通过 Service 删除扫描结果失败: %v", err)
		return err
	}

	logging.Info("通过 Service 成功删除扫描结果: %s", id)
	return nil
}

// MarkResultAsRead 根据任务 ID 修改任务的已读状态（支持已读/未读切换）
func (s *ResultService) MarkResultAsRead(resultID string, isRead bool) error {
	if resultID == "" {
		return errors.New("无效的任务 ID")
	}

	// 获取扫描结果
	result, err := s.resultDAO.GetResultByID(resultID)
	if err != nil {
		logging.Error("获取扫描结果失败: %v", err)
		return err
	}

	// 更新已读状态
	result.IsRead = isRead

	// 保存更新后的扫描结果
	err = s.resultDAO.UpdateResult(resultID, result)
	if err != nil {
		logging.Error("更新扫描结果失败: %v", err)
		return err
	}

	logging.Info("成功更新任务 %s 的已读状态为: %v", resultID, isRead)
	return nil
}

// MarkEntryAsRead 根据任务 ID 和条目 ID 修改条目（端口/指纹/路径）的已读状态
func (s *ResultService) MarkEntryAsRead(resultID string, entryID string) error {
	if resultID == "" || entryID == "" {
		return errors.New("无效的任务 ID 或条目 ID")
	}

	// 获取扫描结果
	result, err := s.resultDAO.GetResultByID(resultID)
	if err != nil {
		logging.Error("获取扫描结果失败: %v", err)
		return err
	}

	// 将 entryID 转换为 ObjectID
	entryObjID, err := primitive.ObjectIDFromHex(entryID)
	if err != nil {
		logging.Error("无效的条目 ID: %v", err)
		return err
	}

	// 遍历扫描结果中的数据，查找对应条目并更新已读状态
	switch result.Type {
	case "Port":
		for _, port := range result.Data.([]*models.Port) {
			if port.ID == entryObjID {
				port.IsRead = true
				break
			}
		}
	case "Fingerprint":
		for _, fingerprint := range result.Data.([]*models.Fingerprint) {
			if fingerprint.ID == entryObjID {
				fingerprint.IsRead = true
				break
			}
		}
	case "Path":
		for _, path := range result.Data.([]*models.Path) {
			if path.ID == entryObjID {
				path.IsRead = true
				break
			}
		}
	default:
		return errors.New("未知的数据类型")
	}

	// 保存更新后的扫描结果
	err = s.resultDAO.UpdateResult(resultID, result)
	if err != nil {
		logging.Error("更新条目已读状态失败: %v", err)
		return err
	}

	logging.Info("成功更新任务 %s 中条目 %s 的已读状态", resultID, entryID)
	return nil
}
