import request, { stringify } from '@/utils/request';

export function createInstance(body) {
  return request('/cacheApi/api/v1alpha2/redis', {
    method: 'POST',
    body,
  });
}

export function modifyInstance(body) {
  return request('/cacheApi/api/v1alpha2/redis', {
    method: 'PUT',
    body,
  });
}

export function getInstanceList() {
  return request('/cacheApi/api/v1alpha2/redis');
}

export function deleteInstance(body) {
  return request('/cacheApi/api/v1alpha2/redis', {
    method: 'DELETE',
    body,
  });
}

export function getSlowLog(name) {
  return request('/cacheApi/api/v1alpha2/redis/slowlog/' + name);
}

export function getOperLog(name) {
  return request('/cacheApi/api/v1alpha2/redis/operlog/' + name);
}

export function flush(body) {
  return request('/cacheApi/api/v1alpha2/redis/flush', {
    method: 'PUT',
    body,
  });
}

export function getConfig(name) {
  return request('/cacheApi/api/v1alpha2/redis/config/' + name);
}

export function updateConfig(name, body) {
  return request('/cacheApi/api/v1alpha2/redis/config/' + name, {
    method: 'PUT',
    body,
  });
}