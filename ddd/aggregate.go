// Package ddd 提供领域驱动设计的核心抽象。
//
// 本包定义 Aggregate、Entity、ValueObject、DomainEvent、Repository、Command 等
// DDD 战术设计的基础接口，供各模块的 domain 层使用。
//
// 设计原则：
//   - 接口最小化，只定义必要的契约
//   - Go 惯用风格，避免过度抽象
//   - 与 Actor 模型正交：Actor 提供物理串行化，Aggregate 提供逻辑一致性
package ddd

// Aggregate 是领域模型的聚合根。
// 聚合是业务一致性边界：对聚合内任何实体的修改都必须通过聚合根。
// 聚合的 ID 在限界上下文中全局唯一。
//
// 与 Actor 的关系：
// 同一玩家的所有聚合由同一个 Actor 保护（物理串行化边界），
// 但每个聚合独立维护自身的业务不变量（逻辑一致性边界）。
type Aggregate interface {
	// ID 返回聚合的唯一标识。
	ID() int64
}

// Entity 是拥有唯一标识的领域对象。
// 与 ValueObject 不同：两个 Entity 即使属性完全相同，只要 ID 不同就是不同对象。
// Entity 总是属于某个 Aggregate，不应跨聚合直接引用。
type Entity interface {
	// ID 返回实体在其所属聚合内的唯一标识。
	ID() int64
}
