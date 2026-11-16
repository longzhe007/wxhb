# wxhb

`wxhb` 是一个面向图片、视频和音频创作的本地无限画布工作台。它将画布编排、AI 生成、素材管理、提示词库和项目保存集中在同一个界面中。

仓库地址：[github.com/longzhe007/wxhb](https://github.com/longzhe007/wxhb)

## 主要功能

- 无限画布：节点创建、连线、分组、复制粘贴、撤销重做、导入导出和项目管理。
- 图片、视频和音频生成：支持 OpenAI 兼容接口及多种模型渠道。
- Muse 提示词库：同步 Muse Video 提示词、作者信息和原帖链接；可公开播放的示例视频会显示缩略图并可在详情中播放。
- 素材管理：将图片、视频、音频和文本素材复用到画布与生成任务。
- 3D 导演台：布置场景、角色、机位和全景环境，并将截图送回画布。
- 提示词与工作流：支持提示词查询、复制、素材沉淀与工作流复用。

## 本地启动

### 环境要求

- Windows 10/11
- Go 1.25 或更高版本
- Node.js 22 或更高版本
- Bun

### 配置

在项目根目录创建本地配置：

```powershell
Copy-Item .env.example .env
```

编辑 `.env`，设置管理员账号、密码和 `JWT_SECRET`。默认使用 SQLite，数据文件保存在 `data/infinite-canvas.db`。

### 开发模式

启动后端：

```powershell
go run .
```

在另一个终端启动前端：

```powershell
Set-Location web
bun install
bun run dev
```

打开：

- 应用：`http://127.0.0.1:3000`
- 管理后台：`http://127.0.0.1:3000/admin`
- 后端健康检查：`http://127.0.0.1:8080/api/health`

### 生产构建

```powershell
Set-Location web
bun run build
```

构建会生成 Next.js standalone 产物，并自动复制静态资源。后端可执行文件可使用以下命令生成：

```powershell
go build -o .runtime/infinite-canvas-api.exe .
```

## 配置 AI 渠道

登录管理后台后，在“设置”中添加模型渠道、基础地址、API Key 和对应模型。图片、视频、文本和音频可分别选择默认模型。

Muse 提示词库中的公开视频使用本地媒体代理播放；代理仅接受受限的公开视频地址，并支持浏览器视频所需的分段读取。

## 文档

- [功能介绍](docs/overview/features.md)
- [画布节点操作手册](docs/canvas/canvas-node-manual.md)
- [画布快捷键](docs/canvas/canvas-shortcuts.md)
- [后端数据库说明](docs/backend/backend-database.md)
- [接口响应约定](docs/backend/api-response.md)

## 开源协议

本项目使用 GNU Affero General Public License v3.0，详见 [LICENSE](LICENSE)。
