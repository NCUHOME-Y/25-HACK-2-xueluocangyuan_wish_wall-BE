
package seeder

import (
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SeedData 负责填充初始数据
func SeedData(db *gorm.DB) {
	// --- 1. 检查是否需要填充 ---
	var count int64
	db.Model(&model.User{}).Where("role = ?", "bot").Count(&count)
	
	if count > 0 {
		zap.S().Info("Seeder: 数据库中已存在 'bot' 用户，跳过数据填充。")
		return
	}
	zap.S().Info("Seeder: 未发现 'bot' 用户，开始执行数据填充...")

	// --- 2. 创建机器人用户 ---
	// (V12 统一模型 需要 Username 和 Password)
	bots := []model.User{
		{Username: "bot_1", Password: "bot_fake_password_hash", Nickname: "许愿星", Role: "bot"},
		{Username: "bot_2", Password: "bot_fake_password_hash", Nickname: "小雪花", Role: "bot"},
		{Username: "bot_3", Password: "bot_fake_password_hash", Nickname: "幸运草", Role: "bot"},
	}
	if err := db.Create(&bots).Error; err != nil {
		zap.S().Fatalf("Seeder: 创建 'bot' 用户失败: %v", err)
	}
	zap.S().Infof("Seeder: 成功创建 %d 个 'bot' 用户。", len(bots))

	// --- 3. 创建 50 条假的心愿 ---
	wishContents := []string{
		"希望期末考试顺利通过!", "想在冬天谈一场甜甜的恋爱", "家人身体健康，万事如意。",
		"希望能找到一份好实习", "什么时候能中彩票啊", "好想去旅游...",
	}
	backgrounds := []string{"star_blue", "moon_purple", "default", "sun_orange"}

	var fakeWishes []model.Wish
	for i := 0; i < 50; i++ {
		fakeWishes = append(fakeWishes, model.Wish{
			UserID:   bots[i%len(bots)].ID,
			Content:  wishContents[i%len(wishContents)],
			Background: backgrounds[i%len(backgrounds)],
			IsPublic: true,
			
		})
	}
	if err := db.Create(&fakeWishes).Error; err != nil {
		zap.S().Fatalf("Seeder: 创建 'fake' 愿望失败: %v", err)
	}
	zap.S().Infof("Seeder: 成功创建 %d 条 'fake' 愿望。", len(fakeWishes))
	zap.S().Info("Seeder: 数据填充成功完成！")
}