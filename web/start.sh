#!/bin/bash
APPOUT_PATH="/home/admin/output/td-redis-operator-ui/logs"
APP_HOME="/home/admin/td-redis-operator-ui"

mkdir -p $APPOUT_PATH

cd $APP_HOME
# 容器的启动
pm2-runtime start pm2.json
