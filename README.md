# 三国志11交易系统

基于三国志11-血色衣冠mod的联赛交易管理系统。

## 功能特性

### 第一阶段（MVP）
- ✅ 用户注册/登录
- ✅ 玩家报名
- ✅ 保底抽将（3次）
- ✅ 普通抽将（7次）
- ✅ 选秀系统
- ✅ 交易系统（发起/接受/拒绝/撤回）
- ✅ 武将/宝物/俱乐部查看
- ✅ 玩家阵容查看
- ✅ 管理员后台（阶段控制/数据导入/重置赛季）

## 技术栈

- **后端**: Go 1.21 + Gin + GORM + SQLite
- **前端**: Vue3 + Element Plus + Pinia + Axios
- **部署**: Docker + Docker Compose

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose（生产部署）

### 本地开发

#### 1. 启动后端

```bash
cd backend

# 下载依赖
go mod tidy

# 创建管理员账号并启动服务
go run ./cmd/server -create-admin -admin-user=admin -admin-pass=admin123

# 或者直接启动
go run ./cmd/server
```

后端默认运行在 `http://localhost:8080`

#### 2. 启动前端

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端默认运行在 `http://localhost:3000`

### 生产部署

#### 使用 Docker Compose（推荐）

```bash
# 构建并启动所有服务
docker-compose up -d --build

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

访问 `http://your-server-ip` 即可使用系统。

#### 手动部署

##### 编译后端

```bash
cd backend
go build -o san11-trade ./cmd/server
./san11-trade -create-admin -admin-user=admin -admin-pass=your_password
```

##### 构建前端

```bash
cd frontend
npm install
npm run build
# dist 目录即为静态文件，可部署到 nginx
```

## 使用说明

### 1. 初始化数据

1. 登录管理员账号
2. 进入"管理后台" -> "数据导入"
3. 上传包含武将/宝物/俱乐部数据的Excel文件

Excel文件格式要求：
- **武将sheet**: 姓名, 统率, 武力, 智力, 政治, 魅力, [薪资], [池类型], [档次], [特技]
- **宝物sheet**: 名称, 类型, [价值], [效果], [特技]
- **俱乐部sheet**: 名称, [描述], [国策], [底价]

### 2. 游戏流程

1. **报名阶段**: 玩家注册并报名
2. **保底抽将**: 每人3次保底抽将机会
3. **普通抽将**: 每人7次普通抽将机会
4. **选秀阶段**: 按顺序选择武将
5. **自由交易**: 玩家之间自由交易武将
6. **比赛阶段**: 进行实际比赛

管理员可在"管理后台"切换各阶段。

## API文档

### 公开接口

| 方法 | 路径 | 说明 |
|-----|-----|-----|
| POST | /api/auth/register | 用户注册 |
| POST | /api/auth/login | 用户登录 |
| GET | /api/phase | 获取当前阶段 |
| GET | /api/generals | 获取所有武将 |
| GET | /api/treasures | 获取所有宝物 |
| GET | /api/clubs | 获取所有俱乐部 |
| GET | /api/players | 获取已报名玩家 |
| GET | /api/players/:id/roster | 获取玩家阵容 |

### 需要登录

| 方法 | 路径 | 说明 |
|-----|-----|-----|
| GET | /api/me | 获取当前用户信息 |
| GET | /api/me/roster | 获取我的阵容 |
| POST | /api/signup | 报名参赛 |
| POST | /api/draw/guarantee | 保底抽将 |
| POST | /api/draw/normal | 普通抽将 |
| POST | /api/draft/pick | 选秀选择 |
| POST | /api/trades | 发起交易 |
| POST | /api/trades/:id/accept | 接受交易 |
| POST | /api/trades/:id/reject | 拒绝交易 |

### 管理员接口

| 方法 | 路径 | 说明 |
|-----|-----|-----|
| POST | /api/admin/phase | 设置游戏阶段 |
| POST | /api/admin/reset | 重置赛季 |
| POST | /api/admin/import | 导入Excel数据 |

## 配置说明

### 环境变量

| 变量 | 默认值 | 说明 |
|-----|-------|-----|
| SERVER_PORT | 8080 | 服务端口 |
| SERVER_HOST | 0.0.0.0 | 监听地址 |
| DB_PATH | ./data/san11trade.db | 数据库路径 |
| JWT_SECRET | san11-trade-secret-key... | JWT密钥（生产环境必须修改） |

### 命令行参数

```bash
./san11-trade [options]

Options:
  -port string        服务端口
  -db string          数据库路径
  -create-admin       创建管理员账号
  -admin-user string  管理员用户名 (default "admin")
  -admin-pass string  管理员密码 (default "admin123")
```

## 目录结构

```
San11-trade/
├── backend/                    # 后端代码
│   ├── cmd/server/            # 主程序入口
│   ├── internal/
│   │   ├── api/               # API路由和处理器
│   │   ├── config/            # 配置管理
│   │   ├── database/          # 数据库连接
│   │   ├── model/             # 数据模型
│   │   └── service/           # 业务逻辑
│   ├── Dockerfile
│   └── go.mod
├── frontend/                   # 前端代码
│   ├── src/
│   │   ├── api/               # API调用
│   │   ├── layouts/           # 布局组件
│   │   ├── router/            # 路由配置
│   │   ├── stores/            # 状态管理
│   │   └── views/             # 页面组件
│   ├── Dockerfile
│   └── package.json
├── data/                       # Excel数据文件
├── docker-compose.yml
└── README.md
```

## 后续规划

- **第二阶段**: 国策系统、对阵配置、对战表导出
- **第三阶段**: 剧本制作集成
- **第四阶段**: QQ机器人、实时拍卖

## 许可证

MIT License
