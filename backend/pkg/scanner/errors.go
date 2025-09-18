package scanner

import "errors"

// 扫描相关错误定义
var (
	// 配置错误
	ErrInvalidConfig    = errors.New("无效的扫描配置")
	ErrInvalidTarget    = errors.New("无效的扫描目标")
	ErrInvalidProjectID = errors.New("无效的项目ID")
	ErrInvalidTimeout   = errors.New("无效的超时时间")

	// 工具错误
	ErrScannerNotFound      = errors.New("扫描工具未找到")
	ErrScannerNotAvailable  = errors.New("扫描工具不可用")
	ErrScannerAlreadyExists = errors.New("扫描工具已存在")
	ErrScannerExecuteFailed = errors.New("扫描工具执行失败")

	// 执行错误
	ErrScanTimeout      = errors.New("扫描执行超时")
	ErrScanCancelled    = errors.New("扫描被取消")
	ErrScanFailed       = errors.New("扫描执行失败")
	ErrInvalidResult    = errors.New("无效的扫描结果")

	// 流水线错误
	ErrPipelineInvalid           = errors.New("无效的扫描流水线")
	ErrStageDependencyNotMet     = errors.New("阶段依赖条件未满足")
	ErrStageExecutionFailed      = errors.New("阶段执行失败")
	ErrCircularDependency        = errors.New("发现循环依赖")

	// 数据转换错误
	ErrDataConversionFailed = errors.New("数据转换失败")
	ErrInvalidScanResult    = errors.New("无效的扫描结果数据")
)