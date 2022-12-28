import fetch from 'isomorphic-fetch';
import { notification } from 'antd';
import { stringify } from 'qs';

export { stringify } from 'qs';

const codeMessage = {
  200: '服务器成功返回请求的数据。',
  201: '新建或修改数据成功。',
  202: '一个请求已经进入后台排队（异步任务）。',
  204: '删除数据成功。',
  400: '发出的请求有错误，服务器没有进行新建或修改数据的操作。',
  401: '用户没有权限（令牌、用户名、密码错误）。',
  403: '用户得到授权，但是访问是被禁止的。',
  404: '发出的请求针对的是不存在的记录，服务器没有进行操作。',
  406: '请求的格式不可得。',
  409: '请求添加的资源已存在',
  410: '请求的资源被永久删除，且不会再得到的。',
  422: '当创建一个对象时，发生一个验证错误。',
  500: '服务器发生错误，请检查服务器。',
  502: '网关错误。',
  503: '服务不可用，服务器暂时过载或维护。',
  504: '网关超时。',
};

const checkStatus = response => {
  if (response.status >= 200 && response.status < 300) {
    return response;
  }
  let errortext = response.statusText || codeMessage[response.status];
  if (!response.body) {
    notification.error({
      message: '请求失败',
      description: errortext,
    });
  } else {
    response.json().then(res => {
      errortext = res.msg || res.message || response.statusText || codeMessage[response.status];
      notification.error({
        message: '请求失败',
        description: errortext,
      });
    });
  }
  const error = new Error(errortext);
  error.name = `${response.status}`;
  error.response = response;
  throw error;
};

const middlewares = [];

/**
 * Requests a URL, returning a promise.
 *
 * @param  {string} url       The URL we want to request
 * @param  {object} [option] The options we want to pass to "fetch"
 * @return {object}           An object containing either "data" or "err"
 */
export default function request(url, option) {
  const options = {
    ...option,
  };

  // 模块市场专用
  if (typeof TDESIGN_SITE_MOCK_DATA !== 'undefined') {
    // eslint-disable-next-line
    return TDESIGN_SITE_MOCK_DATA[`${(options.method || 'GET').toUpperCase()} ${url}`];
  }

  const defaultOptions = {
    credentials: 'include',
  };

  const newOptions = { ...defaultOptions, ...options };
  if (newOptions.method === 'POST' || newOptions.method === 'PUT' || newOptions.method === 'DELETE') {
    if (!(newOptions.body instanceof FormData)) {
      newOptions.headers = {
        Accept: 'application/json',
        'Content-Type': 'application/json; charset=utf-8',
        ...newOptions.headers,
      };
      newOptions.body = JSON.stringify(newOptions.body);
    } else {
      // newOptions.body is FormData
      newOptions.headers = {
        Accept: 'application/json',
        ...newOptions.headers,
      };
    }
  }

  let fetchRes = fetch(url, newOptions)
    .then(checkStatus)
    .then(response => {
      if (newOptions.method === 'DELETE' || response.status === 204) {
        return response.text();
      }
      return response.json();
    });

  middlewares.forEach(handler => {
    fetchRes = fetchRes.then(res => handler(res, newOptions));
  });

  return fetchRes;
}

['get', 'post', 'patch', 'put', 'delete'].forEach(method => {
  request[method] = (url, options) => request(url, { ...options, method: method.toUpperCase() });
});

export function use(cb) {
  middlewares.push(cb);
}

request.use = use;
request.stringify = stringify;

export function postForm(url, data) {
  return fetch(url, {
    method: 'POST',
    headers: {
      'Content-type': 'application/x-www-form-urlencoded; charset=UTF-8',
    },
    credentials: 'include',
    body: Object.keys(data)
      .filter(key => !!data[key])
      .map(key => `${key}=${data[key]}`)
      .join('&'),
  }).then(res => {
    if (res.redirected) {
      if (res.url.indexOf('&failed') >= 0) {
        return false;
      }

      window.location.href = res.url;
      return true;
    }

    return false;
  });
}

export function download(url, options = {}, filename) {
  if (typeof options === 'string') {
    // eslint-disable-next-line
    filename = options;

    // eslint-disable-next-line
    options = {};
  }

  return fetch(url)
    .then(response => {
      try {
        // eslint-disable-next-line
        filename = response.headers.get('content-disposition').match(/filename=([^;]+)/)[1];
      } catch (e) {
        // eslint-disable-next-line
        console.log(e);
      }

      return response.blob();
    })
    .then(blob => {
      const a = document.createElement('a');
      a.href = window.URL.createObjectURL(blob);
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      a.remove();
    });
}
