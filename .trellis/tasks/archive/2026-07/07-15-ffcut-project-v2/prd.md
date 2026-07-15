# FFcut Project v2 协议

## Goal

建立一个纯数据、可校验、可稳定序列化的 `ffcut.Project v2`，作为 FFVMix 与未来 FFmpeg renderer 之间唯一共享的时间线接口。

## Requirements

- Project 必须表达画布宽高、FPS、背景、主视频序列、转场、音频轨、有序全局图层和溯源 metadata。
- 时间在 Go 内使用 `time.Duration`，JSON 中使用带类型的 `int64` 微秒值。
- 视频片段必须表达本地源、源区间、时间线区间、播放速度、循环、末帧定格、画面适配和原声音量。
- 转场必须独立引用前后片段，表达类型、时间和是否处理原声。
- BGM 必须作为独立音频轨表达循环、截断、音量和淡入淡出。
- 图层必须支持字幕和本地图片，使用绝对时间范围以及像素/百分比空间值。
- 背景必须支持纯色、本地图片和模糊主视频三种类型。
- 联合类型必须有明确 discriminator 和唯一匹配的 typed payload，不得使用无限制 `interface{}` 或 `map[string]any`。
- Marshal、Unmarshal 和 Validate 必须返回上下文错误，不得静默丢失错误或修改输入对象。
- Project 必须记录模板摘要、种子、组合指纹、选择结果、素材指纹和约束摘要。
- 不要求兼容当前 `ffcut/fusion` JSON。

## Acceptance Criteria

- [x] 有效 Project 可以 JSON round-trip，深度比较语义一致。
- [x] 无效时间范围、引用、图层 payload、单位、画布或 FPS 都返回可定位错误。
- [x] 时间协议可表达微秒并拒绝溢出或负持续时间。
- [x] 转场引用必须连接相邻主视频片段，且不得超过片段可用范围。
- [x] 图层顺序在 round-trip 后保持不变。
- [x] 序列化不改变输入 Project。
- [x] `go test ./ffcut` 和 `go vet ./ffcut` 通过。

## Out of Scope

- 模板、槽位、候选权重和约束。
- FFprobe 和本地文件存在性检查。
- FFmpeg 参数生成或视频渲染。
- 迁移旧 fusion JSON。
