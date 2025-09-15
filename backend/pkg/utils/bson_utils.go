package utils

import (
	"cyberedge/pkg/logging"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
)

// UnmarshalData 是一个辅助函数，用于将 interface{} 类型的数据解析为指定的结构体
func UnmarshalData(data interface{}, target interface{}) error {
	bsonData, err := bson.Marshal(data)
	if err != nil {
		logging.Error("序列化数据失败: %v", err)
		return errors.New("序列化数据失败")
	}

	if err := bson.Unmarshal(bsonData, target); err != nil {
		logging.Error("解析数据失败: %v", err)
		return errors.New("解析数据失败")
	}

	return nil
}
