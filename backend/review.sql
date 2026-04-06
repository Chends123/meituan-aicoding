CREATE DATABASE IF NOT EXISTS meituan_review_ai DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE meituan_review_ai;

DROP TABLE IF EXISTS reviews;

CREATE TABLE reviews (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  score TINYINT NOT NULL,
  content TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  deleted_at DATETIME NULL,
  INDEX idx_created_id (created_at DESC, id DESC),
  INDEX idx_score (score)
);

INSERT INTO reviews (username, score, content, created_at, updated_at, deleted_at) VALUES
('小张', 2, '环境不错但是服务态度很差，菜品分量也比较少。', '2026-03-31 18:10:00', '2026-03-31 18:10:00', NULL),
('阿宁', 5, '味道很好，尤其是招牌烤鱼，朋友都很满意。', '2026-03-31 19:40:00', '2026-03-31 19:40:00', NULL),
('Momo', 3, '套餐里的饮料不好喝，其他菜都挺满意的。', '2026-04-01 12:15:00', '2026-04-01 12:15:00', NULL),
('老王', 1, '等位时间太长了，服务员爱答不理，体验很差。', '2026-04-01 19:05:00', '2026-04-01 19:05:00', NULL),
('Carrie', 4, '整体环境干净，菜品口味在线，上菜速度也可以。', '2026-04-02 13:25:00', '2026-04-02 13:25:00', NULL),
('阿泽', 2, '分量偏少，不太值这个价格，而且米饭还是冷的。', '2026-04-02 18:30:00', '2026-04-02 18:30:00', NULL),
('可乐', 5, '服务热情，招牌菜很有特色，性价比不错。', '2026-04-03 11:22:00', '2026-04-03 11:22:00', NULL),
('Luna', 4, '味道不错，装修很有氛围，就是排队稍微久一点。', '2026-04-03 17:18:00', '2026-04-03 17:18:00', NULL),
('子墨', 2, '套餐搭配不合理，饮料太甜不好喝，服务响应也慢。', '2026-04-04 12:45:00', '2026-04-04 12:45:00', NULL),
('张小北', 5, '团购非常划算，菜品新鲜，家人都说下次还来。', '2026-04-04 18:55:00', '2026-04-04 18:55:00', NULL),
('Yuki', 3, '菜品总体还行，但是高峰期体验一般，等位和出餐都慢。', '2026-04-05 12:10:00', '2026-04-05 12:10:00', NULL),
('老刘', 1, '服务态度太差了，催菜也没人理，不会再来了。', '2026-04-05 19:48:00', '2026-04-05 19:48:00', NULL),
('Ann', 4, '烤鱼入味，配菜也不错，整体比较满意。', '2026-04-06 11:05:00', '2026-04-06 11:05:00', NULL),
('阿杰', 5, '朋友聚餐选这里挺合适，环境和味道都在线。', '2026-04-06 13:50:00', '2026-04-06 13:50:00', NULL),
('Wendy', 2, '菜量有点少，套餐饮料一般，桌面清理也不及时。', '2026-04-06 18:12:00', '2026-04-06 18:12:00', NULL);
