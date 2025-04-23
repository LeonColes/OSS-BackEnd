# 数据库初始化脚本 (PowerShell版)
# 用法：.\scripts\init_db.ps1 [配置文件路径]

param (
    [string]$ConfigFile = ".\configs\config.yaml"
)

# 检查配置文件是否存在
if (-not (Test-Path $ConfigFile)) {
    Write-Host "配置文件不存在: $ConfigFile" -ForegroundColor Red
    exit 1
}

Write-Host "正在读取配置文件: $ConfigFile" -ForegroundColor Cyan

# 读取配置文件中的数据库连接信息
$configContent = Get-Content $ConfigFile -Raw
$dsnMatch = [regex]::Match($configContent, 'dsn:\s*(.*?)(?=\r\n|\n)')

if (-not $dsnMatch.Success) {
    Write-Host "无法在配置文件中找到数据库DSN" -ForegroundColor Red
    exit 1
}

$dsn = $dsnMatch.Groups[1].Value.Trim()

# 提取数据库连接信息
$userPassMatch = [regex]::Match($dsn, '(.*?):(.*?)@')
$hostPortMatch = [regex]::Match($dsn, '@tcp\((.*?):(.*?)\)')
$dbNameMatch = [regex]::Match($dsn, '\/([^?]*)(?:\?|$)')

if (-not ($userPassMatch.Success -and $hostPortMatch.Success -and $dbNameMatch.Success)) {
    Write-Host "DSN格式不正确: $dsn" -ForegroundColor Red
    exit 1
}

$dbUser = $userPassMatch.Groups[1].Value
$dbPass = $userPassMatch.Groups[2].Value
$dbHost = $hostPortMatch.Groups[1].Value
$dbPort = $hostPortMatch.Groups[2].Value
$dbName = $dbNameMatch.Groups[1].Value

Write-Host "数据库连接信息:" -ForegroundColor Green
Write-Host "  用户: $dbUser"
Write-Host "  主机: $dbHost"
Write-Host "  端口: $dbPort"
Write-Host "  数据库: $dbName"

# 检查MySQL客户端是否可用
try {
    $mysqlPath = (Get-Command mysql -ErrorAction Stop).Source
    Write-Host "MySQL客户端已安装: $mysqlPath" -ForegroundColor Green
} catch {
    Write-Host "MySQL客户端未安装，请先安装MySQL客户端" -ForegroundColor Red
    Write-Host "可以从 https://dev.mysql.com/downloads/mysql/ 下载安装" -ForegroundColor Yellow
    exit 1
}

# 测试数据库连接
try {
    $output = & mysql --user="$dbUser" --password="$dbPass" --host="$dbHost" --port="$dbPort" -e "SELECT 1;" 2>&1
    if ($LASTEXITCODE -ne 0) { throw "数据库连接失败" }
    Write-Host "数据库连接成功" -ForegroundColor Green
} catch {
    Write-Host "数据库连接失败，请检查配置: $_" -ForegroundColor Red
    exit 1
}

# 检查数据库是否存在
try {
    $output = & mysql --user="$dbUser" --password="$dbPass" --host="$dbHost" --port="$dbPort" -e "USE $dbName;" 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "数据库 '$dbName' 已存在" -ForegroundColor Green
    } else {
        Write-Host "数据库 '$dbName' 不存在，正在创建..." -ForegroundColor Yellow
        $output = & mysql --user="$dbUser" --password="$dbPass" --host="$dbHost" --port="$dbPort" -e "CREATE DATABASE ``$dbName`` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "数据库 '$dbName' 创建成功" -ForegroundColor Green
        } else {
            Write-Host "数据库创建失败: $output" -ForegroundColor Red
            exit 1
        }
    }
} catch {
    Write-Host "检查数据库时出错: $_" -ForegroundColor Red
    exit 1
}

Write-Host "`n数据库初始化完成，现在可以启动应用" -ForegroundColor Cyan
Write-Host "应用将在首次启动时自动创建表结构并初始化基础数据"

exit 0 