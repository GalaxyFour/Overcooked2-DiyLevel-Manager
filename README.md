# Overcooked2 DIY 关卡管理器

基于 [Overcooked2-LevelEditor](https://github.com/gua248/Overcooked2-LevelEditor) 导出包的社区关卡分发平台。

## 功能

- 作者上传 Level Editor 导出的 zip，后台串行解析、校验、提取截图与元数据
- 腾讯云 COS 私有存储，包下载 6h 预签名、图片 48h 预签名（自动滚动刷新）
- 公开主页浏览与下载关卡集
- 三角色权限：作者 / 管理员 / 超级管理员
- 日/夜主题、可爱厨房风格 UI

## 端口

| 环境 | 端口 | 说明 |
|------|------|------|
| 后端 API | 14556 | 开发 API |
| 前端 Vite | 14557 | 开发前端（proxy `/api` → 14556） |
| 生产 | 14558 | 静态前端 + API 同进程 |

## 快速开始

```bash
# 1. 安装依赖
make deps

# 2. 复制并编辑配置（填入 COS 密钥）
cp configs/config.example.yaml configs/config.yaml

# 3. 开发模式（两个终端）
make dev-api   # :14556
make dev-fe    # :14557

# 4. 生产构建
make prod      # :14558
```

默认超级管理员：`admin` / `admin123`（可在 config 中修改，仅首次启动时创建）

## 配置说明

见 [`configs/config.example.yaml`](configs/config.example.yaml)

- **COS 留空**：使用本地 `data/cos-local/` 模拟存储（开发用）
- **parser**：需安装 `UnityPy` 和 `Pillow`（`pip install -r tools/requirements.txt`）

## 上传格式

zip 命名：`{slug}_v{version}_{yyyyMMdd}.zip`

```
levels/<slug>/info_<slug>
levels/<slug>/s_*
commonW1 / commonW2 / runtime（可选）
OC2LevelRuntimeLoader.dll（可选）
```

## API 文档

- 开发：http://localhost:14556/swagger/index.html
- Swagger JSON：http://localhost:14556/swagger/doc.json

## 技术栈

- 后端：Go + chi + SQLite (WAL) + JWT
- 前端：React + Vite + Tailwind + framer-motion
- 解析：Python UnityPy 子进程
