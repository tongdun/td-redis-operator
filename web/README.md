# td-redis-operator - ui

## 依赖安装
npm install

## 本地开发
npm start 启动服务后自动打开开发页面http://localhost:9090

## 生产环境启动
pm2-runtime start pm2.json

## 相关文档
- [Antd组件库](https://ant-design.gitee.io/components/overview-cn/)

## 配置说明

配置文件路径： ./config
默认配置：./config/default.json
本地开发配置：./config/development.json
生产环境配置：./config/production.json

## 状态库说明

当前react主要使用react hooks api，状态库主要使用[hox](https://github.com/umijs/hox/blob/master/README-cn.md);
react hooks的使用参考 [hooks最佳实践](http://wiki.tongdun.me/pages/viewpage.action?pageId=33236516)