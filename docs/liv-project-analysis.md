# Liv项目代码整理 - FFmpeg Go封装库

## 项目概览

Liv是一个用Go语言编写的FFmpeg友好封装库，提供了视频转码、截图、音频处理等多媒体处理功能。项目采用模块化设计，支持多种视频格式和滤镜效果。

## 项目结构

```
liv/
├── doc/                    # 文档目录
├── examples/               # 示例代码
│   ├── ffcut/             # FFcut示例
│   ├── ffmpeg/            # FFmpeg示例
│   ├── ffprobe/           # FFprobe示例
│   └── service/           # 服务示例
├── ffcut/                 # 视频剪切模块
├── ffmpeg/                # FFmpeg核心模块
│   ├── codec/             # 编码器定义
│   ├── filter/            # 滤镜模块
│   ├── input/             # 输入处理
│   ├── output/            # 输出处理
│   ├── stream/            # 流处理
│   └── util/              # 工具函数
├── ffprobe/               # 媒体信息探测模块
├── fftool/                # FFmpeg工具集
├── pkg/                   # 通用工具包
│   ├── conv/              # 类型转换
│   ├── encoding/          # 编码处理
│   ├── math/              # 数学计算
│   └── sugar/             # 语法糖
├── liv.go                 # 主入口文件
├── transcode.go           # 转码功能实现
├── snapshot.go            # 截图功能实现
└── 其他配置和参数文件
```

## 核心模块详解

### 1. 主入口模块 (liv.go)

定义了核心接口：

```go
type Transcoder interface {
    Transcode(context.Context, *TranscodeParams) error
}

type VideoFilter interface{}
```

### 2. 转码模块 (transcode.go)

#### 核心功能：
- **SimpleMP4**: 转码为MP4格式，支持多路输出
- **SimpleMP3**: 转码为MP3音频格式
- **SimpleJPEG**: 转码为JPEG图片
- **SimpleHLS**: 转码为HLS流媒体格式
- **SimpleTS**: 转码为TS格式
- **ConvertContainer**: 容器格式转换
- **Concat**: 视频合并
- **ExtractAudio**: 提取音频
- **MergeByFrames**: 帧序列合并

#### 支持的滤镜效果：
- **视频缩放**: 自动调整分辨率，支持偶数像素对齐
- **水印添加**: 支持多个水印，可设置位置、缩放
- **去水印**: 指定矩形区域进行遮标处理
- **视频裁剪**: 支持时间段裁剪
- **元数据添加**: 支持自定义键值对元数据

### 3. 截图模块 (snapshot.go)

#### 核心功能：
- **Simple**: 基础截图功能
  - 关键帧截图 (FrameType=0)
  - 等间隔截图 (FrameType=1)
  - 指定帧序列截图 (FrameType=2)
- **Sprite**: 生成雪碧图
- **SVGMark**: SVG标注功能，支持矩形、画笔、箭头、文字标注

### 4. FFmpeg核心模块

#### Filter滤镜系统：
```go
// 视频滤镜
- Overlay: 图像覆盖
- Scale: 视频缩放
- Delogo: 去水印
- Crop: 视频裁剪
- Select: 帧选择
- Split: 流分割
- Tile: 瓦片布局

// 音频滤镜
- ASplit: 音频流分割
- ASetPTS: 音频时间戳设置
```

#### 编码器支持 (codec/):
```go
// 视频编码器
X264 = "libx264"
X265 = "libx265" 
VP9 = "libvpx-vp9"
MJPEG = "mjpeg"

// 音频编码器
AAC = "aac"
MP3Lame = "libmp3lame"

// 容器格式
MP4 = "mp4"
HLS = "hls"
```

### 5. FFprobe媒体探测模块

提供媒体文件信息探测功能：
- 获取视频流信息
- 获取音频流信息
- 获取格式信息
- 支持JSON格式输出

### 6. 参数结构体系统

#### 转码参数 (transcode_params.go)：
```go
type TranscodeParams struct {
    Infile string                    // 输入文件
    Subs   []*SubTranscodeParams     // 多路输出配置
}

type Filters struct {
    Video     *Video      // 视频参数
    Audio     *Audio      // 音频参数
    Logo      []*Logo     // 水印配置
    Delogo    []*Delogo   // 去水印配置
    Clip      *Clip       // 裁剪配置
    Metadata  []*KV       // 元数据
}
```

#### 截图参数 (snapshot_params.go)：
```go
type SnapshotParams struct {
    Infile         string    // 输入文件
    Outfile        string    // 输出文件
    FrameType      int32     // 截图类型
    Interval       int32     // 间隔时间
    IntervalFrames int32     // 间隔帧数
    StartTime      float32   // 开始时间
    Width, Height  int32     // 输出尺寸
}
```

## 使用示例

### 1. 视频转码示例

```go
// MP4转码
tc := liv.NewTranscode(
    liv.FFmpegOptions(
        ffmpeg.WithBin("ffmpeg"),
        ffmpeg.WithDebug(true),
    ),
)

params := &liv.TranscodeParams{
    Infile: "input.mp4",
    Subs: []*liv.SubTranscodeParams{
        {
            Outfile: "output.mp4",
            Filters: &liv.Filters{
                Video: &liv.Video{
                    Codec:  codec.X264,
                    Height: 720,
                    Crf:    23,
                },
                Audio: &liv.Audio{
                    Codec: codec.AAC,
                },
                Logo: []*liv.Logo{
                    {
                        File: "logo.png",
                        Pos:  "TopRight",
                        Dx:   10,
                        Dy:   10,
                    },
                },
                Clip: &liv.Clip{
                    Seek:     5.0,
                    Duration: 60.0,
                },
            },
        },
    },
}

err := tc.SimpleMP4(ctx, params)
```

### 2. 截图示例

```go
// 视频截图
ss := liv.NewSnapshot(
    liv.FFmpegOptions(
        ffmpeg.WithBin("ffmpeg"),
        ffmpeg.WithDebug(true),
    ),
)

params := &liv.SnapshotParams{
    Infile:    "input.mp4",
    Outfile:   "output-%05d.jpg",
    FrameType: 1,           // 等间隔截图
    Interval:  5,           // 每5秒截一张
    Width:     1280,
    Height:    720,
}

err := ss.Simple(ctx, params)
```

## Rust实现建议

基于Go版本的架构，Rust实现应考虑以下设计：

### 1. 模块结构
```rust
src/
├── lib.rs              // 库入口
├── transcoder/         // 转码模块
│   ├── mod.rs
│   ├── params.rs       // 参数定义
│   └── spec.rs         // 规格验证
├── snapshot/           // 截图模块
├── ffmpeg/             // FFmpeg核心
│   ├── filter/         // 滤镜
│   ├── codec/          // 编码器
│   ├── input/          // 输入处理
│   └── output/         // 输出处理
├── ffprobe/            // 媒体探测
└── utils/              // 工具函数
```

### 2. 核心特性利用

#### 类型安全
```rust
// 使用枚举确保编码器类型安全
#[derive(Debug, Clone)]
pub enum VideoCodec {
    X264,
    X265,
    VP9,
    Copy,
}

// 使用泛型和特征约束
pub trait Filter {
    fn params(&self) -> Vec<String>;
    fn name(&self) -> String;
}
```

#### 错误处理
```rust
use thiserror::Error;

#[derive(Error, Debug)]
pub enum LivError {
    #[error("Invalid parameters: {0}")]
    InvalidParams(String),
    #[error("FFmpeg execution failed: {0}")]
    FFmpegError(String),
    #[error("IO error: {0}")]
    IoError(#[from] std::io::Error),
}
```

#### 异步支持
```rust
use tokio::process::Command;

pub struct Transcoder {
    options: TranscoderOptions,
}

impl Transcoder {
    pub async fn simple_mp4(&self, params: &TranscodeParams) -> Result<(), LivError> {
        let mut cmd = Command::new("ffmpeg");
        cmd.args(&self.build_args(params)?);
        
        let output = cmd.output().await?;
        if !output.status.success() {
            return Err(LivError::FFmpegError(
                String::from_utf8_lossy(&output.stderr).into_owned()
            ));
        }
        Ok(())
    }
}
```

### 3. 性能优化

#### 零拷贝字符串处理
```rust
use std::borrow::Cow;

pub struct FilterChain<'a> {
    filters: Vec<Cow<'a, str>>,
}

impl<'a> FilterChain<'a> {
    pub fn add_filter(&mut self, filter: impl Into<Cow<'a, str>>) {
        self.filters.push(filter.into());
    }
}
```

#### 构建器模式
```rust
pub struct TranscodeParamsBuilder {
    infile: Option<String>,
    subs: Vec<SubTranscodeParams>,
}

impl TranscodeParamsBuilder {
    pub fn new() -> Self {
        Self {
            infile: None,
            subs: Vec::new(),
        }
    }
    
    pub fn input_file<S: Into<String>>(mut self, file: S) -> Self {
        self.infile = Some(file.into());
        self
    }
    
    pub fn add_output(mut self, sub: SubTranscodeParams) -> Self {
        self.subs.push(sub);
        self
    }
    
    pub fn build(self) -> Result<TranscodeParams, LivError> {
        Ok(TranscodeParams {
            infile: self.infile.ok_or(LivError::InvalidParams("Input file required".into()))?,
            subs: self.subs,
        })
    }
}
```

### 4. 并发处理
```rust
use tokio::sync::Semaphore;
use std::sync::Arc;

pub struct TranscodePool {
    semaphore: Arc<Semaphore>,
}

impl TranscodePool {
    pub fn new(max_concurrent: usize) -> Self {
        Self {
            semaphore: Arc::new(Semaphore::new(max_concurrent)),
        }
    }
    
    pub async fn transcode(&self, params: TranscodeParams) -> Result<(), LivError> {
        let _permit = self.semaphore.acquire().await.unwrap();
        // 执行转码任务
        self.do_transcode(params).await
    }
}
```

### 5. 配置管理
```rust
use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize, Serialize)]
pub struct LivConfig {
    pub ffmpeg_bin: String,
    pub ffprobe_bin: String,
    pub max_concurrent: usize,
    pub default_quality: u8,
    pub temp_dir: String,
}

impl Default for LivConfig {
    fn default() -> Self {
        Self {
            ffmpeg_bin: "ffmpeg".to_string(),
            ffprobe_bin: "ffprobe".to_string(),
            max_concurrent: 4,
            default_quality: 23,
            temp_dir: "/tmp".to_string(),
        }
    }
}
```

## 总结

Liv项目是一个功能完善的FFmpeg Go封装库，提供了丰富的视频处理功能。Rust实现时应该：

1. **保持API设计的简洁性**：延续Go版本的简单易用特点
2. **利用Rust的类型系统**：提供更好的类型安全和错误处理
3. **支持异步操作**：充分利用Rust的async/await特性
4. **优化性能**：减少不必要的内存分配和复制
5. **提供良好的配置管理**：支持序列化/反序列化配置

通过这样的设计，Rust版本将能够提供更好的性能、安全性和并发处理能力。