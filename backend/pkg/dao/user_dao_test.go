package dao

import (
	"cyberedge/pkg/models"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTestDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})

	if err != nil {
		return nil, nil, err
	}

	return gormDB, mock, nil
}

func TestUserDAO_Create(t *testing.T) {
	db, mock, err := setupTestDB()
	assert.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	userDAO := NewUserDAO(db)

	user := &models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         "user",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
	}

	// 期望的SQL执行
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `users`").
		WithArgs(user.Username, user.Email, user.PasswordHash, false, "", "user", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = userDAO.Create(user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_GetByUsername(t *testing.T) {
	db, mock, err := setupTestDB()
	assert.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	userDAO := NewUserDAO(db)

	tests := []struct {
		name     string
		username string
		setup    func()
		wantErr  bool
	}{
		{
			name:     "User found",
			username: "testuser",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "is_2fa_enabled", "totp_secret", "role", "created_at", "updated_at"}).
					AddRow(1, "testuser", "test@example.com", "hashedpassword", false, "", "user", time.Now().Unix(), time.Now().Unix())
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE username = \\? ORDER BY `users`.`id` LIMIT \\?").
					WithArgs("testuser", 1).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:     "User not found",
			username: "nonexistent",
			setup: func() {
				mock.ExpectQuery("SELECT \\* FROM `users` WHERE username = \\? ORDER BY `users`.`id` LIMIT \\?").
					WithArgs("nonexistent", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			user, err := userDAO.GetByUsername(tt.username)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.username, user.Username)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserDAO_GetByEmail(t *testing.T) {
	db, mock, err := setupTestDB()
	assert.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	userDAO := NewUserDAO(db)

	email := "test@example.com"
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "is_2fa_enabled", "totp_secret", "role", "created_at", "updated_at"}).
		AddRow(1, "testuser", email, "hashedpassword", false, "", "user", time.Now().Unix(), time.Now().Unix())

	mock.ExpectQuery("SELECT \\* FROM `users` WHERE email = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(email, 1).
		WillReturnRows(rows)

	user, err := userDAO.GetByEmail(email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_GetByID(t *testing.T) {
	db, mock, err := setupTestDB()
	assert.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	userDAO := NewUserDAO(db)

	userID := uint(1)
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "is_2fa_enabled", "totp_secret", "role", "created_at", "updated_at"}).
		AddRow(userID, "testuser", "test@example.com", "hashedpassword", false, "", "user", time.Now().Unix(), time.Now().Unix())

	mock.ExpectQuery("SELECT \\* FROM `users` WHERE `users`.`id` = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs(userID, 1).
		WillReturnRows(rows)

	user, err := userDAO.GetByID(userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_Update(t *testing.T) {
	db, mock, err := setupTestDB()
	assert.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	userDAO := NewUserDAO(db)

	user := &models.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "newhashedpassword",
		Is2FAEnabled: true,
		TOTPSecret:   "newsecret",
		Role:         "user",
		CreatedAt:    time.Now().Unix(),
		UpdatedAt:    time.Now().Unix(),
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `users` SET").
		WithArgs(user.Username, user.Email, user.PasswordHash, user.Is2FAEnabled, user.TOTPSecret, user.Role, sqlmock.AnyArg(), sqlmock.AnyArg(), user.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = userDAO.Update(user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_Delete(t *testing.T) {
	db, mock, err := setupTestDB()
	assert.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	userDAO := NewUserDAO(db)

	userID := uint(1)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `users` WHERE `users`.`id` = \\?").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = userDAO.Delete(userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_GetAll(t *testing.T) {
	db, mock, err := setupTestDB()
	assert.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	userDAO := NewUserDAO(db)

	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "is_2fa_enabled", "totp_secret", "role", "created_at", "updated_at"}).
		AddRow(1, "user1", "user1@example.com", "hash1", false, "", "user", time.Now().Unix(), time.Now().Unix()).
		AddRow(2, "user2", "user2@example.com", "hash2", false, "", "user", time.Now().Unix(), time.Now().Unix())

	mock.ExpectQuery("SELECT \\* FROM `users`").
		WillReturnRows(rows)

	users, err := userDAO.GetAll()

	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "user1", users[0].Username)
	assert.Equal(t, "user2", users[1].Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserDAO_DatabaseError(t *testing.T) {
	db, mock, err := setupTestDB()
	assert.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	userDAO := NewUserDAO(db)

	// 模拟数据库连接错误
	mock.ExpectQuery("SELECT \\* FROM `users` WHERE username = \\? ORDER BY `users`.`id` LIMIT \\?").
		WithArgs("testuser", 1).
		WillReturnError(sql.ErrConnDone)

	user, err := userDAO.GetByUsername("testuser")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}