package stack

// 跨模块事件 Topic 常量。
// 模块间通过 EventBus 发布/订阅这些事件，实现解耦通信。
const (
	// 玩家事件
	EventPlayerLogin   = "player:login"    // 玩家登录 (payload: uid int64)
	EventPlayerLogout  = "player:logout"   // 玩家登出 (payload: uid int64)
	EventPlayerCreated = "player:created"  // 玩家创建 (payload: uid int64)
	EventPlayerLevelUp = "player:level_up" // 玩家升级 (payload: uid int64, new_level int32)

	// 物品事件
	EventItemAcquire = "item:acquire" // 获得物品 (payload: uid int64, item_id int32, count int32)
	EventItemConsume = "item:consume" // 消耗物品 (payload: uid int64, item_id int32, count int32)
	EventItemEquip   = "item:equip"   // 装备物品 (payload: uid int64, item_id int32)
	EventItemUnequip = "item:unequip" // 卸下物品 (payload: uid int64, item_id int32)

	// 邮件事件
	EventMailNew        = "mail:new"    // 新邮件 (payload: uid int64, mail_id int64)
	EventMailRead       = "mail:read"   // 邮件已读 (payload: uid int64, mail_id int64)
	EventMailAttachRecv = "mail:attach" // 附件领取 (payload: uid int64, mail_id int64)

	// 公会事件
	EventGuildCreate      = "guild:create"       // 公会创建 (payload: guild_id int64)
	EventGuildDissolve    = "guild:dissolve"     // 公会解散 (payload: guild_id int64)
	EventGuildMemberJoin  = "guild:member_join"  // 成员加入 (payload: guild_id int64, uid int64)
	EventGuildMemberLeave = "guild:member_leave" // 成员离开 (payload: guild_id int64, uid int64)
	EventGuildLevelUp     = "guild:level_up"     // 公会升级 (payload: guild_id int64, new_level int32)

	// 匹配事件
	EventMatchFound   = "match:found"   // 匹配成功 (payload: match_id string)
	EventMatchTimeout = "match:timeout" // 匹配超时 (payload: uid int64)
	EventMatchCancel  = "match:cancel"  // 匹配取消 (payload: uid int64)

	// 战斗事件
	EventCombatStart  = "combat:start"  // 战斗开始 (payload: combat_id string)
	EventCombatEnd    = "combat:end"    // 战斗结束 (payload: combat_id string)
	EventCombatDamage = "combat:damage" // 造成伤害 (payload: combat_id string, uid int64, target int64, damage int32)

	// 商城事件
	EventShopPurchase = "shop:purchase" // 购买成功 (payload: uid int64, item_id int32, count int32)

	// 活动事件
	EventActivityStart = "activity:start" // 活动开始 (payload: activity_id int32)
	EventActivityEnd   = "activity:end"   // 活动结束 (payload: activity_id int32)

	// 排行榜事件
	EventLeaderboardUpdate = "leaderboard:update" // 排行榜更新 (payload: board_name string, uid int64, new_rank int32)

	// 社交事件
	EventFriendAdd    = "social:friend_add"    // 添加好友 (payload: uid int64, friend_uid int64)
	EventFriendRemove = "social:friend_remove" // 删除好友 (payload: uid int64, friend_uid int64)
)
