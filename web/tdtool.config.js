const path = require('path');
const Config = require('tdtool').Config;
const isDebug = process.env.NODE_ENV !== 'production';

const clientConfig = new Config({
  entry: {
    cip: './client/index',
  },
  sourceMap: true,
  devtool: 'source-map',
  filename: '[name].[hash].js',
  minimize: !isDebug,
  disableCSSModules: true,
  alias: {
    '@': path.resolve(__dirname, 'client'),
  },
  extends: [
    [
      'react',
      {
        plugins: [['import', { libraryName: 'antd', style: true }]],
        source: [path.resolve(__dirname, 'client')],
      },
    ],
    [
      'less',
      {
        extractCss: {
          filename: '[name].[hash].css',
          allChunks: true,
        },
        happypack: true,
        theme: {
          '@primary-color': '#567bff',
        },
      },
    ],
  ],
  env: {
    __DEV__: isDebug,
  },
});

clientConfig.add('output.path', path.join(process.cwd(), 'dist', 'client'));
clientConfig.add('output.publicPath', '/');
clientConfig.add('output.chunkFilename', '[name].[chunkhash].chunk.js');
const AssetsPlugin = require('assets-webpack-plugin');
clientConfig.add(
  'plugin.AssetsPlugin',
  new AssetsPlugin({
    path: './dist/client',
    filename: 'assets.json',
    prettyPrint: true,
  }),
);

const serverConfig = new Config({
  entry: './server/index.js',
  target: 'node',
  filename: 'server.js',
  sourceMap: true,
  devServer: isDebug,
  externals: [/^\.\/client\/assets\.json$/, require('webpack-node-externals')()],
});

serverConfig.add('resolve.extensions', ['.js']);

module.exports = [clientConfig.resolve(), serverConfig.resolve()];
