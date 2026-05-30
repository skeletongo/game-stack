package ddd

// ValueObject 表示一个没有独立标识的不可变值。
//
// 值对象的核心特征：
//   - 无 ID：由属性值定义相等性
//   - 不可变：创建后不可修改，修改 = 创建新对象
//   - 自校验：构造时检查不变量，拒绝非法值
//
// 在 game-stack 中，典型的 ValueObject 包括：
//   - PlayerID, Nickname, Level, Gold, Diamond
//   - ItemID, ItemCount
//   - Token, Password
type ValueObject interface {
	// Equals 按值比较两个值对象是否相等。
	Equals(other ValueObject) bool
}

// Scalar 是值对象的可选接口。实现了 Scalar 的值对象可被 debug 服务
// 自动展开为 JSON 原始类型（Gold(999) → 999），无需手写序列化。
type Scalar interface {
	// Scalar 返回值对象的底层 Go 值。
	Scalar() any
}
