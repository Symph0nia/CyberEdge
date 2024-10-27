package service

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
)

type ConfigService struct {
	configDAO *dao.ConfigDAO
}

func NewConfigService(configDAO *dao.ConfigDAO) *ConfigService {
	return &ConfigService{configDAO: configDAO}
}

func (s *ConfigService) GetQRCodeStatus() (bool, error) {
	logging.Info("正在获取二维码状态")
	status, err := s.configDAO.GetQRCodeStatus()
	if err != nil {
		logging.Error("获取二维码状态失败: %v", err)
		return false, err
	}
	logging.Info("成功获取二维码状态: %v", status)
	return status, nil
}

func (s *ConfigService) SetQRCodeStatus(enabled bool) error {
	logging.Info("正在设置二维码状态为: %v", enabled)
	err := s.configDAO.SetQRCodeStatus(enabled)
	if err != nil {
		logging.Error("设置二维码状态失败: %v", err)
		return err
	}
	logging.Info("成功设置二维码状态为: %v", enabled)
	return nil
}
