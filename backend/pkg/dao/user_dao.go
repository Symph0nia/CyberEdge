// pkg/dao/user_dao.go

package dao

import (
	"context"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserDAO struct {
	collection *mongo.Collection
}

func NewUserDAO(collection *mongo.Collection) *UserDAO {
	return &UserDAO{collection: collection}
}

func (d *UserDAO) GetQRCodeStatus() (bool, error) {
	logging.Info("正在获取二维码状态")
	var config struct {
		QRCodeEnabled bool `bson:"qrcode_enabled"`
	}

	err := d.collection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logging.Info("未找到二维码状态配置，使用默认值 true")
			return true, nil // 默认值，如果没有文档
		}
		logging.Error("获取二维码状态失败: %v", err)
		return false, err
	}
	logging.Info("成功获取二维码状态: %v", config.QRCodeEnabled)
	return config.QRCodeEnabled, nil
}

func (d *UserDAO) GetAllUsers() ([]models.User, error) {
	logging.Info("正在获取所有用户")
	cursor, err := d.collection.Find(context.Background(), bson.M{})
	if err != nil {
		logging.Error("获取所有用户失败: %v", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var users []models.User
	if err := cursor.All(context.Background(), &users); err != nil {
		logging.Error("解析用户数据失败: %v", err)
		return nil, err
	}
	logging.Info("成功获取所有用户，共 %d 个", len(users))
	return users, nil
}

func (d *UserDAO) GetUserByAccount(account string) (*models.User, error) {
	logging.Info("正在获取用户信息: %s", account)
	var user models.User
	err := d.collection.FindOne(context.Background(), bson.M{"account": account}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logging.Warn("用户不存在: %s", account)
			return nil, err
		}
		logging.Error("获取用户信息失败: %s, 错误: %v", account, err)
		return nil, err
	}
	logging.Info("成功获取用户信息: %s", account)
	return &user, nil
}

func (d *UserDAO) CreateUser(user *models.User) error {
	logging.Info("正在创建新用户: %s", user.Account)
	_, err := d.collection.InsertOne(context.Background(), user)
	if err != nil {
		logging.Error("创建用户失败: %s, 错误: %v", user.Account, err)
		return err
	}
	logging.Info("用户创建成功: %s", user.Account)
	return nil
}

type DeleteUsersResult struct {
	DeletedAccounts []string          // 成功删除的账号
	FailedAccounts  map[string]string // 失败的账号及原因
}

func (d *UserDAO) DeleteUsers(accounts []string) (*DeleteUsersResult, error) {
	logging.Info("正在批量删除用户: %v", accounts)

	result := &DeleteUsersResult{
		DeletedAccounts: make([]string, 0),
		FailedAccounts:  make(map[string]string),
	}

	// 创建批量删除条件
	filter := bson.M{"account": bson.M{"$in": accounts}}

	// 执行批量删除
	_, err := d.collection.DeleteMany(context.Background(), filter)
	if err != nil {
		logging.Error("批量删除用户失败: %v", err)
		return nil, err
	}

	// 查询剩余的账号（未被删除的）
	var remainingAccounts []string
	cursor, err := d.collection.Find(context.Background(),
		bson.M{"account": bson.M{"$in": accounts}})
	if err != nil {
		logging.Error("查询剩余账号失败: %v", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var users []struct{ Account string }
	if err := cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}
	for _, user := range users {
		remainingAccounts = append(remainingAccounts, user.Account)
	}

	// 计算成功和失败的账号
	successMap := make(map[string]bool)
	for _, account := range accounts {
		found := false
		for _, remaining := range remainingAccounts {
			if account == remaining {
				result.FailedAccounts[account] = "删除失败"
				found = true
				break
			}
		}
		if !found {
			result.DeletedAccounts = append(result.DeletedAccounts, account)
			successMap[account] = true
		}
	}

	logging.Info("批量删除用户完成，成功删除: %d, 删除失败: %d",
		len(result.DeletedAccounts), len(result.FailedAccounts))

	return result, nil
}

func (d *UserDAO) IncrementLoginCount(account string) (int, error) {
	logging.Info("正在增加用户登录次数: %s", account)
	result := d.collection.FindOneAndUpdate(
		context.Background(),
		bson.M{"account": account},
		bson.M{"$inc": bson.M{"loginCount": 1}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var updatedUser models.User
	err := result.Decode(&updatedUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logging.Warn("尝试更新不存在的用户登录次数: %s", account)
		} else {
			logging.Error("更新用户登录次数失败: %s, 错误: %v", account, err)
		}
		return 0, err
	}
	logging.Info("成功更新用户登录次数: %s, 新的登录次数: %d", account, updatedUser.LoginCount)
	return updatedUser.LoginCount, nil
}
