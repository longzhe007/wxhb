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

## Windows 非 Docker 部署

以下流程适用于将项目部署在 Windows 电脑或 Windows 服务器上。应用会运行在本机 `3000` 端口，后端 API 运行在 `8080` 端口。

### 1. 获取代码与安装依赖

```powershell
git clone https://github.com/longzhe007/wxhb.git
Set-Location wxhb
Copy-Item .env.example .env
go mod download

Set-Location web
bun install --frozen-lockfile
Set-Location ..
```

首次部署前编辑根目录 `.env`，至少修改以下配置：

```text
ADMIN_USERNAME=admin
ADMIN_PASSWORD=请设置强密码
JWT_SECRET=请设置一段随机长字符串
DATABASE_DSN=data/infinite-canvas.db
```

`data/infinite-canvas.db` 保存 SQLite 数据库。升级前请备份整个 `data` 目录。

### 2. 构建生产产物

```powershell
Set-Location web
bun run build
Set-Location ..

New-Item -ItemType Directory -Force .runtime | Out-Null
go build -o .runtime/infinite-canvas-api.exe .
```

`bun run build` 会生成 `web/.next/standalone`，并自动复制 Next.js 静态资源与 `public` 目录。

### 3. 后台启动

```powershell
New-Item -ItemType Directory -Force .runtime/logs | Out-Null

Start-Process `
  -FilePath "$PWD/.runtime/infinite-canvas-api.exe" `
  -WorkingDirectory $PWD `
  -WindowStyle Hidden `
  -RedirectStandardOutput "$PWD/.runtime/logs/api.stdout.log" `
  -RedirectStandardError "$PWD/.runtime/logs/api.stderr.log"

$env:PORT = "3000"
$env:HOSTNAME = "127.0.0.1"
Start-Process `
  -FilePath "node" `
  -ArgumentList "server.js" `
  -WorkingDirectory "$PWD/web/.next/standalone" `
  -WindowStyle Hidden `
  -RedirectStandardOutput "$PWD/.runtime/logs/web.stdout.log" `
  -RedirectStandardError "$PWD/.runtime/logs/web.stderr.log"
```

如需让局域网设备访问，将 `HOSTNAME` 改为 `0.0.0.0`，并在 Windows 防火墙中仅向可信网络开放 `3000` 端口。

### 4. 验证部署

```powershell
(Invoke-WebRequest http://127.0.0.1:3000/).StatusCode
(Invoke-WebRequest http://127.0.0.1:8080/api/health).StatusCode
```

两个命令都应返回 `200`。运行日志位于 `.runtime/logs/`。

### 5. 停止服务

```powershell
Get-NetTCPConnection -LocalPort 3000,8080 -State Listen -ErrorAction SilentlyContinue |
  Select-Object -ExpandProperty OwningProcess -Unique |
  ForEach-Object { Stop-Process -Id $_ -Force }
```

### 6. 更新部署

停止服务后执行：

```powershell
git pull --ff-only origin main
go mod download

Set-Location web
bun install --frozen-lockfile
bun run build
Set-Location ..

go build -o .runtime/infinite-canvas-api.exe .
```

然后重复“后台启动”和“验证部署”两节。

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
