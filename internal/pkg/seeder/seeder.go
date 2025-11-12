package seeder

import (
	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/app/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SeedData 负责填充初始数据
func SeedData(db *gorm.DB) {
	//  检查是否需要填充 
	var count int64
	db.Model(&model.User{}).Where("role = ?", "bot").Count(&count)

	if count > 0 {
		zap.S().Info("Seeder: 数据库中已存在 'bot' 用户，跳过数据填充。")
		return
	}
	zap.S().Info("Seeder: 未发现 'bot' 用户，开始执行数据填充...")

	//  创建机器人用户 ---
	bots := []model.User{
		{Username: "bot_1", Password: "bot_fake_password_hash", Nickname: "许愿星", Role: "bot"},
		{Username: "bot_2", Password: "bot_fake_password_hash", Nickname: "小雪花", Role: "bot"},
		{Username: "bot_3", Password: "bot_fake_password_hash", Nickname: "幸运草", Role: "bot"},
		{Username: "bot_4", Password: "bot_fake_password_hash", Nickname: "奶农大人", Role: "bot"},
		{Username: "bot_5", Password: "bot_fake_password_hash", Nickname: "奶农小人", Role: "bot"},
		{Username: "bot_6", Password: "bot_fake_password_hash", Nickname: "杰伦", Role: "bot"},
		{Username: "bot_7", Password: "bot_fake_password_hash", Nickname: "好好", Role: "bot"},
		{Username: "bot_8", Password: "bot_fake_password_hash", Nickname: "坏坏", Role: "bot"},
		{Username: "bot_9", Password: "bot_fake_password_hash", Nickname: "河大妈", Role: "bot"},
		{Username: "bot_10", Password: "bot_fake_password_hash", Nickname: "海的女儿", Role: "bot"},
	}
	if err := db.Create(&bots).Error; err != nil {
		zap.S().Fatalf("Seeder: 创建 'bot' 用户失败: %v", err)
	}
	zap.S().Infof("Seeder: 成功创建 %d 个 'bot' 用户。", len(bots))

	//  创建 50 条假的心愿 
	wishContents := []string{
		"希望期末考试顺利通过!", "欸欸我要谈外国帅哥", "家人身体健康，万事如意。",
		"希望能找到一份好实习", "什么时候能中彩票啊", "好想去旅游...", "我要吃十斤咖喱饭",
		"大家都要开开心心健康幸福呀", "我想当富二代", "希望明天是个好天气",
		"希望世界和平呀", "希望今天我的linux do注册成功", "希望能交到更多朋友",
		"希望能考上理想的研究生院", "希望能有个甜甜的恋爱", "希望能养一只小猫咪",
		"希望能买到心仪的电子产品", "希望能早日实现财务自由", "希望能多睡会儿觉",
		"希望能学会弹吉他", "希望能有机会出国留学", "希望能找到理想的工作",
		"希望能多读几本好书", "希望能改善人际关系", "希望能提升自己的厨艺",
		"好想早点回家", "小学生愿望成真",
	}

	var fakeWishes []model.Wish
	for i := 0; i < 55; i++ {
		fakeWishes = append(fakeWishes, model.Wish{
			UserID:   bots[i%len(bots)].ID,
			Content:  wishContents[i%len(wishContents)],
			IsPublic: true,
		})
	}
	if err := db.Create(&fakeWishes).Error; err != nil {
		zap.S().Fatalf("Seeder: 创建 'fake' 愿望失败: %v", err)
	}
	zap.S().Infof("Seeder: 成功创建 %d 条 'fake' 愿望。", len(fakeWishes))
	zap.S().Info("Seeder: 数据填充成功完成！")
}
