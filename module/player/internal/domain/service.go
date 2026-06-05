package domain

// CalcLevel 根据总经验值计算等级（领域服务，纯函数）。
//
// 升级公式：level=1 需要 100 exp，之后每级递增 100。
//
//	level 1: 0-99 exp
//	level 2: 100-299 exp
//	level 3: 300-599 exp
func CalcLevel(exp Exp) Level {
	e := exp.Int64()
	level := int32(1)
	needed := int64(100)
	for e >= needed {
		e -= needed
		level++
		needed = int64(level) * 100
	}
	lv, _ := NewLevel(level)
	return lv
}
