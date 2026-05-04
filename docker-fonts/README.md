# docker-fonts - 自定义字体目录

本目录用于存放 Docker 构建时使用的中文字体。项目已自带 Noto 开源字体（思源宋体 + 思源黑体），
开箱即用；如果有 Windows 原版字体可以替换以获得更精确的渲染效果。

## 已自带的开源字体

| 文件名 | 字体名称 | 用途 | 许可 |
|---------|----------|------|------|
| `NotoSerifSC-Regular.otf` | 思源宋体 Regular | 替代宋体/仿宋 Regular | OFL |
| `NotoSerifSC-Bold.otf` | 思源宋体 Bold | 替代宋体/仿宋 Bold | OFL |
| `NotoSansSC-Regular.otf` | 思源黑体 Regular | 替代黑体 Regular | OFL |
| `NotoSansSC-Bold.otf` | 思源黑体 Bold | 替代黑体 Bold | OFL |

这些字体由 Google/Adobe 联合开发，采用 SIL Open Font License (OFL) 许可，可自由分发。
来源：[noto-cjk](https://github.com/notofonts/noto-cjk)

## 可选：使用 Windows 原版字体（优先级更高）

如果需要更精确的字体渲染效果（例如与 Windows 上的 PDF 完全一致），可以将 Windows 字体放入本目录。
构建时 Windows 字体会**优先于** Noto 字体生效。

| 文件名 | 字体名称 | 说明 |
|---------|----------|------|
| `simsun.ttc` | 宋体 (SimSun) | Windows 默认中文衬线字体 |
| `simhei.ttf` | 黑体 (SimHei) | Windows 默认中文无衬线字体 |
| `simkai.ttf` | 楷体 (SimKai) | 手写风格字体 |
| `simfang.ttf` | 仿宋 (SimFang) | 印刷风格字体 |

从 Windows 系统复制：

```
C:\Windows\Fonts\simsun.ttc
C:\Windows\Fonts\simhei.ttf
C:\Windows\Fonts\simkai.ttf
C:\Windows\Fonts\simfang.ttf
```

## 字体优先级

构建时 Ghostscript cidfmap 映射的优先级（从高到低）：

1. **Windows 原版字体**（simsun/simhei/simkai/simfang）—— 最精确
2. **Noto 开源字体**（NotoSerifSC/NotoSansSC）—— 已自带，质量优秀
3. **系统默认字体**（arphic-uming/arphic-ukai/wqy-zenhei）—— 兜底

## 使用

1. 正常执行 `docker build` 或 `make docker-build`
2. 构建过程会自动安装字体并更新 Ghostscript 的字体映射

## 注意事项

- Noto 字体为 OFL 许可，可以自由分发和提交到 Git
- **SimSun/SimHei 等 Windows 字体为微软版权所有**，仅限个人/内部使用，请勿将包含这些字体的 Docker 镜像公开分发
- 支持的字体格式：`.ttf`、`.ttc`、`.otf`
