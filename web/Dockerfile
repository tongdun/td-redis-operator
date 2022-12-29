FROM node:18.12-slim

RUN npm install pm2 -g
# 工作目录/home/admin/td-redis-operator-ui，不要改

WORKDIR /home/admin/td-redis-operator-ui
# 按此处参考的顺序放入需要的文件
COPY config ./config
COPY dist ./dist
COPY public ./public
COPY node_modules  ./node_modules
COPY pm2.json start.sh ./

# 运行启动脚本
CMD ["/bin/bash", "-c" ,"sh /home/admin/td-redis-operator-ui/start.sh" ]
