# FFVMix 惰性生成器与约束

## Goal

基于不可变 `CompiledTemplate`，按需、确定地生成唯一且满足约束的 `ffcut.Project`，不预先展开完整组合空间。

## Requirements

- 组合维度包括每槽位一个视频候选、每连接一个转场候选和可选 BGM 池中的一个候选。
- 完全相同的组合永远只枚举一次。
- 权重只影响出现顺序，不改变可行性，也不允许重复。
- 种子可选；自动生成的实际种子必须可读取。
- 相同编译模板、约束和种子产生相同结果顺序。
- `Next(ctx)` 串行消费；同一生成器不支持并发调用或第一版 checkpoint。
- 生成状态必须区分 `Yielded / Exhausted / BudgetExceeded`。
- 单次调用受搜索预算和 context 控制，预算用尽后保留进度。
- 筛选采用贪心语义，不回溯、不承诺全局最大集合。
- 约束必须是无副作用插件，只读取 `CandidateView` 和 `HistoryView`。
- 内置相似度按相同槽位、相同素材的重叠视频时长计算。
- 视频路径复用和 BGM 复用必须是独立约束。
- 通过约束后才构建完整 Project；图层时间锚点必须解析为绝对范围。
- Project metadata 必须包含完整生成溯源。

## Acceptance Criteria

- [x] 小组合空间完整遍历后，每个组合只出现一次且正确返回 Exhausted。
- [x] 相同种子结果完全一致，不同种子可以改变顺序但不改变无约束全集。
- [x] 高权重候选在统计测试中更倾向于较早出现，同时所有候选仍可到达。
- [x] BudgetExceeded 不等同于 Exhausted，继续调用可以恢复扫描。
- [x] 约束拒绝不会更新接受历史，约束错误会中止调用并保留一致状态。
- [x] MaxSimilarity、MaxVideoAssetUses 和 MaxBGMUses 均有边界测试。
- [x] 不可行转场/素材组合不会生成 Project，并记录稳定拒绝原因。
- [x] 生成 Project 的时长、转场重叠、BGM 和全局图层绝对时间正确。
- [x] 并发调用得到明确错误或被文档化的竞态保护，不产生调度相关结果。

## Out of Scope

- 全局最优或最大数量求解。
- 图层候选参与组合。
- 跨进程恢复。
- FFmpeg 渲染。
