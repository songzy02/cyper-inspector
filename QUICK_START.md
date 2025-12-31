# Cyber Inspector 快速开始指南

## 🎯 概述

本指南将帮助您在 10 分钟内快速部署并运行 Cyber Inspector 系统。

## 📋 环境准备

### 必需组件

- **Go 1.21+** (仅源码部署需要)
- **MySQL 5.7+**
- **Linux/Unix 系统** (推荐 Ubuntu 20.04+)

### 可选组件

- **Docker & Docker Compose** (推荐，最简单)

## 🚀 三种部署方式

### 方式一：Docker 部署（推荐 ⭐）

**适合场景**：快速部署、开发测试、生产环境

#### 1. 安装 Docker

```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com | bash -s docker

# 安装 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### 2. 克隆项目

```bash
git clone https://github.com/yourusername/cyber-inspector.git
cd cyber-inspector
```

#### 3. 一键启动

```bash
# 使用 Makefile
make docker-run

# 或者直接使用 Docker Compose
docker-compose up -d
```

#### 4. 验证部署

```bash
# 查看容器状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

#### 5. 访问系统

打开浏览器访问：http://localhost:8080

- 用户名：admin
- 密码：admin123

> **注意**：首次登录后请立即修改密码

---

### 方式二：源码部署

**适合场景**：二次开发、学习研究、自定义配置

#### 1. 安装 Go

```bash
# 下载 Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz

# 解压
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# 设置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

#### 2. 安装 MySQL

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install mysql-server mysql-client

# CentOS/RHEL
sudo yum install mysql-server mysql

# 启动 MySQL
sudo systemctl start mysql
sudo systemctl enable mysql
```

#### 3. 创建数据库

```bash
# 登录 MySQL
mysql -u root -p

# 执行初始化脚本
source scripts/mysql/init/init.sql
```

#### 4. 克隆并编译项目

```bash
# 克隆项目
git clone https://github.com/yourusername/cyber-inspector.git
cd cyber-inspector

# 下载依赖
go mod download

# 编译
make build
```

#### 5. 配置项目

编辑配置文件 `configs/config.yaml`：

```yaml
mysql:
  dsn: "root:your_password@tcp(127.0.0.1:3306)/cyber_inspector?charset=utf8mb4&parseTime=True&loc=Local"
```

#### 6. 启动 Master

```bash
# 使用 Makefile
make run-master

# 或者直接运行
./bin/cyber-inspector --config=configs/config.yaml
```

#### 7. 启动 Agent（在每台被监控服务器上）

```bash
# 复制 cyber-agent 到目标服务器
scp bin/cyber-agent user@server:/path/to/

# 在目标服务器上运行
./cyber-agent --port=8083
```

---

### 方式三：二进制部署

**适合场景**：无 Go 环境、快速部署、生产环境

#### 1. 下载预编译二进制文件

```bash
# 从 Release 页面下载
wget https://github.com/yourusername/cyber-inspector/releases/download/v2.0.0/cyber-inspector-linux-amd64.tar.gz
wget https://github.com/yourusername/cyber-inspector/releases/download/v2.0.0/cyber-agent-linux-amd64.tar.gz

# 解压
tar -xzf cyber-inspector-linux-amd64.tar.gz
tar -xzf cyber-agent-linux-amd64.tar.gz
```

#### 2. 创建数据库

同方式二的步骤 2-3

#### 3. 配置并运行

```bash
# 创建配置文件目录
mkdir -p configs

# 复制配置模板
cp config.yaml.example configs/config.yaml

# 编辑配置
vi configs/config.yaml

# 运行 Master
./cyber-inspector --config=configs/config.yaml

# 运行 Agent
./cyber-agent --port=8083
```

---

## 🎮 使用指南

### 1. 添加监控节点

1. 登录系统（admin/admin123）
2. 点击"添加节点"按钮
3. 填写节点信息：
   - **节点名称**：如 "Web服务器-01"
   - **IP 地址**：如 "192.168.1.100"
   - **Agent URL**：如 "http://192.168.1.100:8083"
   - **巡检间隔**：建议 300 秒（5分钟）
4. 点击保存

### 2. 查看节点状态

- 绿色：正常运行
- 黄色：警告级别告警
- 红色：严重级别告警
- 灰色：节点离线

### 3. 手动触发巡检

点击"立即巡检"按钮，系统会立即对所有活跃节点执行一次巡检。

### 4. 查看告警记录

点击"告警记录"菜单，查看历史告警信息。

---

## 🔧 常见问题

### 问题1：数据库连接失败

**症状**：启动时报 "database connection failed"

**解决方案**：

1. 检查 MySQL 是否运行：
   ```bash
   sudo systemctl status mysql
   ```

2. 检查数据库是否创建：
   ```bash
   mysql -u root -p -e "SHOW DATABASES;"
   ```

3. 检查配置文件中的 DSN：
   ```yaml
   mysql:
     dsn: "username:password@tcp(127.0.0.1:3306)/cyber_inspector?charset=utf8mb4&parseTime=True&loc=Local"
   ```

### 问题2：Agent 无法连接

**症状**：节点显示离线状态

**解决方案**：

1. 检查 Agent 是否运行：
   ```bash
   ps aux | grep cyber-agent
   ```

2. 检查端口是否监听：
   ```bash
   netstat -lnp | grep 8083
   ```

3. 检查防火墙：
   ```bash
   sudo ufw allow 8083
   ```

4. 测试 Agent 接口：
   ```bash
   curl http://agent-ip:8083/health
   ```

### 问题3：邮件告警不发送

**症状**：告警产生但没有收到邮件

**解决方案**：

1. 检查邮件配置：
   ```yaml
   mail:
     enabled: true
     host: "smtp.163.com"
     port: 994
     user: "your_email@163.com"
     pass: "your_password"  # 注意：这里是授权码，不是邮箱密码
   ```

2. 检查网络连接：
   ```bash
   telnet smtp.163.com 994
   ```

3. 查看日志：
   ```bash
   # Docker
   docker-compose logs master
   
   # 源码
   tail -f logs/app.log
   ```

---

## 📊 性能优化建议

### 1. 数据库优化

```sql
-- 为常用查询添加索引
CREATE INDEX idx_inspections_agent_created ON inspections(agent_id, created_at);
CREATE INDEX idx_alerts_status ON alerts(status);
```

### 2. 配置优化

```yaml
# 增加并发数
check:
  max_concurrent: 20
  
# 减少巡检间隔
check:
  interval: "2m"
```

### 3. 系统优化

```bash
# 增加文件描述符限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# 优化 TCP 连接
sysctl -w net.core.somaxconn=1024
```

---

## 🔄 升级指南

### 从 v1.x 升级到 v2.0

1. **备份数据库**：
   ```bash
   mysqldump -u root -p cyber_inspector > backup.sql
   ```

2. **更新代码**：
   ```bash
   git pull origin main
   ```

3. **执行数据库迁移**：
   ```bash
   mysql -u root -p cyber_inspector < scripts/mysql/migrate_v1_to_v2.sql
   ```

4. **重新编译**：
   ```bash
   make clean
   make build
   ```

5. **更新配置**：
   ```bash
   cp configs/config.yaml.example configs/config.yaml.new
   # 合并您的自定义配置到新文件
   ```

6. **重启服务**：
   ```bash
   ./bin/cyber-inspector --config=configs/config.yaml
   ```

---

## 📞 技术支持

### 获取帮助

1. **查看日志**：
   ```bash
   # Docker
   docker-compose logs master
   docker-compose logs agent1
   
   # 源码
   tail -f logs/app.log
   ```

2. **开启调试模式**：
   ```yaml
   log:
     level: "debug"
   ```

3. **查看 API 文档**：
   ```bash
   # 启动服务后访问
   http://localhost:8080/swagger/index.html
   ```

### 联系我们

- 📧 Email: support@cyber-inspector.com
- 💬 微信: CyberInspector
- 🐛 Issue: https://github.com/yourusername/cyber-inspector/issues

---

## 📚 相关文档

- [详细部署指南](DEPLOYMENT.md)
- [API 接口文档](API.md)
- [配置说明](CONFIG.md)
- [开发指南](DEVELOPMENT.md)

---

**祝您使用愉快！** 🎉
