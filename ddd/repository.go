package ddd

import "context"

// Repository 是聚合的持久化仓储接口。
//
// 泛型参数 T 约束为 Aggregate 类型，确保仓储只操作聚合根。
// 仓储负责聚合的整体加载和保存，不暴露内部实体。
//
// 设计决策：
//   - 以聚合为单位：Load 加载完整聚合，Save 保存完整聚合
//   - Save 应为幂等操作
//   - Delete 删除数据
//
// 使用泛型而非 any + 类型断言，编译期保证类型安全。
type Repository[T Aggregate] interface {
	// Load 按 ID 加载聚合。不存在时返回错误。
	Load(ctx context.Context, id int64) (T, error)

	// Save 持久化聚合。
	Save(ctx context.Context, aggregate T) error

	// Delete 删除聚合。已删除的数据再次删除应返回 nil（幂等）。
	Delete(ctx context.Context, id int64) error
}
