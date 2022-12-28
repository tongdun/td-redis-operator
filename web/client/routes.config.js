import React, { lazy } from 'react';

const siderMenus = {
  '/cloud/cache': {
    title: '云缓存',
    menus: [
      {
        cnName: 'Redis实例',
        enName: 'Redis Instance',
        code: 'redisInstance',
        url: '/#/cloud/cache/redis',
        isMenu: true,
      },
      {
        cnName: '版本说明',
        enName: 'Release Notes',
        code: 'releaseNotes',
        url: '/#/cloud/cache/releaseNotes',
        isMenu: true,
      },
      {
        cnName: '帮助文档',
        enName: 'Document',
        code: 'document',
        url: 'https://github.com/tongdun/td-redis-operator/wiki',
        isMenu: true,
        target: '_blank'
      },
    ],
  },
};

export function findMenus(location) {
  const prefix = Object.keys(siderMenus).find(o => location.startsWith(o));
  if (prefix) {
    return siderMenus[prefix];
  }

  return { menus: [] };
}

export default [
  {
    path: '/',
    routes: [
      // 首页
      { path: '/', exact: true, redirect: '/cloud/cache' },

      {
        path: '/403',
        exact: true,
        component: lazy(() => import('@/components/Exception/403')),
      },
      {
        path: '/404',
        exact: true,
        component: lazy(() => import('@/components/Exception/404')),
      },
      {
        path: '/500',
        exact: true,
        component: lazy(() => import('@/components/Exception/500')),
      },
      // 云缓存首页
      { path: '/cloud/cache', exact: true, redirect: '/cloud/cache/redis' },
      {
        path: '/cloud/cache/redis',
        exact: true,
        component: lazy(() => import('@/pages/CloudCache/index')),
      },
      {
        path: '/cloud/cache/releaseNotes',
        exact: true,
        component: lazy(() => import('@/pages/CloudCache/releaseNotes')),
      },
      { path: '*', exact: true, redirect: '/404' },
    ],
  },
];
