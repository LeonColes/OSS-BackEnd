#!/bin/bash

# 数据库初始化脚本
# 用法：./scripts/init_db.sh [配置文件路径]

CONFIG_FILE=${1:-"./configs/config.yaml"}

# 读取配置文件中的数据库连接信息
if [ -f "$CONFIG_FILE" ]; then
    echo "正在读取配置文件: $CONFIG_FILE"
    
    # 使用grep和awk提取数据库配置信息
    DB_DSN=$(grep -A 5 "database:" "$CONFIG_FILE" | grep "dsn:" | awk -F': ' '{print $2}')
    
    # 从DSN中提取用户名、密码、主机和数据库名
    DB_USER=$(echo $DB_DSN | awk -F':' '{print $1}')
    DB_PASS=$(echo $DB_DSN | awk -F':' '{print $2}' | awk -F'@' '{print $1}')
    DB_HOST=$(echo $DB_DSN | awk -F'@tcp(' '{print $2}' | awk -F':' '{print $1}')
    DB_PORT=$(echo $DB_DSN | awk -F':' '{print $3}' | awk -F')/' '{print $1}')
    DB_NAME=$(echo $DB_DSN | awk -F'/' '{print $2}' | awk -F'?' '{print $1}')
    
    echo "数据库连接信息:"
    echo "  用户: $DB_USER"
    echo "  主机: $DB_HOST"
    echo "  端口: $DB_PORT"
    echo "  数据库: $DB_NAME"
    
    # 检查MySQL客户端是否可用
    if command -v mysql >/dev/null 2>&1; then
        echo "MySQL客户端已安装"
        
        # 检查数据库连接
        if mysql -u"$DB_USER" -p"$DB_PASS" -h"$DB_HOST" -P"$DB_PORT" -e "SELECT 1" >/dev/null 2>&1; then
            echo "数据库连接成功"
            
            # 检查数据库是否存在
            if mysql -u"$DB_USER" -p"$DB_PASS" -h"$DB_HOST" -P"$DB_PORT" -e "USE $DB_NAME" >/dev/null 2>&1; then
                echo "数据库 '$DB_NAME' 已存在"
            else
                echo "数据库 '$DB_NAME' 不存在，正在创建..."
                mysql -u"$DB_USER" -p"$DB_PASS" -h"$DB_HOST" -P"$DB_PORT" -e "CREATE DATABASE $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
                if [ $? -eq 0 ]; then
                    echo "数据库 '$DB_NAME' 创建成功"
                else
                    echo "数据库创建失败"
                    exit 1
                fi
            fi
        else
            echo "数据库连接失败，请检查配置"
            exit 1
        fi
    else
        echo "MySQL客户端未安装，请先安装MySQL客户端"
        exit 1
    fi
else
    echo "配置文件不存在: $CONFIG_FILE"
    exit 1
fi

echo "数据库初始化完成，现在可以启动应用"
echo "应用将在首次启动时自动创建表结构并初始化基础数据"
exit 0 