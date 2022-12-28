import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom';
import { useLocation } from 'react-router';
import { HashRouter as Router } from 'react-router-dom';
import { Spin } from 'antd';
import moment from 'moment';
import routesConfig, { findMenus } from './routes.config';
import renderRoutes from './utils/renderRoutes';
import useGlobalModel from '@/models/useGlobalModel';
import 'moment/locale/zh-cn';
import './style.less';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/es/locale/zh_CN';
import PortalLayout from '@/components/Layout';

// hot reload
if (process.env.NODE_ENV === 'development') {
  if (module.hot) {
    module.hot.accept();
  }
}

moment.locale('zh-cn');

const App = () => {
  const [menus, setMenus] = useState([]);
  const [sysTitle, setSysTitle] = useState();
  const { pathname } = useLocation();
  const { loading, setUser } = useGlobalModel();

  // console.log(user);

  if (loading) {
    return (
      <div
        style={{
          width: '100vw',
          height: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}>
        <Spin size="large" />
      </div>
    );
  }

  useEffect(() => {
    const info = findMenus(pathname);
    setMenus(info.menus);
    setSysTitle(info.title);
  }, [pathname]);

  return (
    <PortalLayout customMenu={menus} sysTitle={sysTitle} onUserChange={setUser}>
      {renderRoutes(routesConfig)}
    </PortalLayout>
  );
};

ReactDOM.render(
  <Router>
    <ConfigProvider locale={zhCN}>
      <App />
    </ConfigProvider>
  </Router>,
  document.getElementById('app'),
);
