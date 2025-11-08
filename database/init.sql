-- 创建数据库
CREATE DATABASE IF NOT EXISTS wish_wall CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE wish_wall;

-- 在这里放置上面所有的 CREATE TABLE 语句

-- 插入初始皮肤数据
INSERT INTO skins (name, description, price, image_url) VALUES
('星空蓝', '深邃的蓝色星空主题', 100, '/skins/star_blue.png'),
('月光紫', '神秘的紫色月光主题', 150, '/skins/moon_purple.png'),
('晨曦金', '温暖的晨曦金色主题', 200, '/skins/sunrise_gold.png'),
('森林绿', '清新的森林绿色主题', 180, '/skins/forest_green.png');

-- 创建索引优化查询性能
CREATE INDEX idx_wishes_user_public ON wishes(user_id, is_public);
CREATE INDEX idx_comments_wish_parent ON comments(wish_id, parent_id);
CREATE INDEX idx_likes_user_wish ON likes(user_id, wish_id);