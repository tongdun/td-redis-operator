import express from 'express';
import path from 'path';
import morgan from 'morgan';
import fs from 'fs';
import config from 'config';
import fetch from 'isomorphic-fetch';

const proxy = require('express-http-proxy');

let env = process.env.ENV;
const isDebug = env === 'dev' || env === undefined;

let assets = {};
if (fs.existsSync(path.resolve(__dirname, './client/assets.json'))) {
  assets = JSON.parse(fs.readFileSync(path.resolve(__dirname, './client/assets.json')).toString());
}

const time = new Date().getTime();
const app = express();
const port = process.env.port || 8088; // 默认端口

//////////////////////////////////////// log config ////////////////////////////////////////////
// 1. console log
app.use(
  morgan('combined', {
    skip: (req, _res) => req.url && req.url.indexOf('/ok.htm') >= 0,
  }),
);

///////////////////////////////////// end of log config ///////////////////////////////////////

function renderView(conf) {
  return `
  <!DOCTYPE html>
  <html lang="en">
    <head>
      <title>td-redis-operator</title>
      <meta charset="utf-8>
      <meta http-equiv="content-type" content="text/html;charset=utf-8">
      <meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <link rel="stylesheet" type="text/css" href="${conf.cip.css}?v=${time}">
    </head>
    <body>
      <div id="app" />
      <script src="${conf.cip.js}?v=${time}"></script>
    </body>
  </html>
  `;
}

app.use(express.static(path.resolve(__dirname, 'client')));
app.use(express.static(path.resolve(__dirname, '..', 'public')));

app.get('/', (req, res) => {
  res.send(
    renderView({
      ...assets,
    }),
  );
});

// TODO: 接口代理加密传递用户信息
app.use(
  '/cacheApi',
  proxy(config.get('proxy.cloudCache'), {
    proxyReqOptDecorator: function(proxyReqOpts, srcReq) {
      return proxyReqOpts;
    },
  }),
);

app.get('/ok.htm', (_req, res) => res.send('ok')); // 服务是否可用的检查接口

app
  .listen(port, () => {
    if (isDebug) {
      console.log(`The server is running at http://localhost.tongdun.cn:${port}/`);
    } else {
      console.log(`The server is running at http://127.0.0.1:${port}/`);
    }
  })
  .setTimeout(20 * 60 * 1000);
