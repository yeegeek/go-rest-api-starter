package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yeegeek/go-rest-api-starter/internal/testutil"
)

// TestUserRepository_Integration 使用真实 PostgreSQL 容器进行集成测试
func TestUserRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	ctx := context.Background()

	// 启动 PostgreSQL 容器
	pgContainer, err := testutil.NewPostgresContainer(ctx)
	require.NoError(t, err)
	defer pgContainer.Close(ctx)

	// 运行迁移
	err = pgContainer.DB.AutoMigrate(&User{})
	require.NoError(t, err)

	// 创建仓储
	repo := NewRepository(pgContainer.DB)

	t.Run("CreateUser", func(t *testing.T) {
		user := &User{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "hashedpassword",
		}

		err := repo.Create(ctx, user)
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		// 创建测试用户
		user := &User{
			Name:     "Find Me",
			Email:    "findme@example.com",
			Password: "hashedpassword",
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// 通过邮箱查找
		found, err := repo.GetByEmail(ctx, "findme@example.com")
		assert.NoError(t, err)
		assert.Equal(t, user.Email, found.Email)
		assert.Equal(t, user.Name, found.Name)
	})

	t.Run("Update", func(t *testing.T) {
		// 创建测试用户
		user := &User{
			Name:     "Original Name",
			Email:    "update@example.com",
			Password: "hashedpassword",
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// 更新用户
		user.Name = "Updated Name"
		err = repo.Update(ctx, user)
		assert.NoError(t, err)

		// 验证更新
		found, err := repo.GetByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
	})

	t.Run("Delete", func(t *testing.T) {
		// 创建测试用户
		user := &User{
			Name:     "To Be Deleted",
			Email:    "delete@example.com",
			Password: "hashedpassword",
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// 删除用户
		err = repo.Delete(ctx, user.ID)
		assert.NoError(t, err)

		// 验证已删除
		_, err = repo.GetByID(ctx, user.ID)
		assert.Error(t, err)
		assert.Equal(t, ErrUserNotFound, err)
	})
}
