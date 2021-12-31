package configurable

const (
	DefaultThread = 12 // 默认协程数量

	// 配置中上下文中的一些标志

	CollyConfName      = "__conf.name"      // 配置名称
	CollyConfStepName  = "__conf.step.name" // 步骤名称
	CollyConfExt       = "__conf.ext"       // 额外信息
	CollyConfStepStart = "start"            // 开始步骤名称
	CollyConfStepEnd   = "final"            // 结束步骤名称
)
