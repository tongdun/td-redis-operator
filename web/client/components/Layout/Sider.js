import React, { useState, useEffect } from 'react';
import { Layout, Menu } from 'antd';
import { addHistoryListener, removeHistoryListener } from './historyListener';
import './style.less';

const { Sider } = Layout;
const { SubMenu } = Menu;

export default ({ title, collapsible, menus, collapsed }) => {
  const [openKeys, setOpenKeys] = useState([]);
  const [selectedKeys, setSelectedKeys] = useState([]);

  const onSelect = data => {
    setSelectedKeys(data.selectedKeys);
  };

  const onRouteChange = () => {
    let needOpen = false;

    for (let i = 0; i < menus.length; i += 1) {
      const subMenu = menus[i];
      const location = window.location.href.split('?')[0];
      const item = subMenu.isMenu
        ? location.endsWith(subMenu.url) && subMenu
        : subMenu.children.find(o => location.endsWith(o.url));
      if (item) {
        needOpen = true;
        setSelectedKeys([item.code]);
        setOpenKeys([subMenu.code]);
        break;
      }
    }

    if (!needOpen) {
      setSelectedKeys([]);
      setOpenKeys([]);
    }
  };

  useEffect(() => {
    onRouteChange();
    if (menus.length) {
      addHistoryListener(onRouteChange);
    }

    return () => removeHistoryListener(onRouteChange);
  }, [menus]);

  return (
    <Sider
      trigger={null}
      collapsible={collapsible}
      collapsed={collapsed}
      className="portal-sider"
      width={216}
      collapsedWidth={0}
      theme="light">
      {title && <div className="portal-sider-title">{title}</div>}
      <div className="portal-sider-divider" />
      <Menu
        mode="inline"
        style={{ paddingLeft: 0, marginBottom: 0 }}
        selectedKeys={selectedKeys}
        openKeys={openKeys}
        onOpenChange={setOpenKeys}
        onSelect={onSelect}>
        {menus.map(subMenu =>
          !subMenu.isMenu ? (
            <SubMenu
              key={subMenu.code}
              title={
                <span>
                  <span>{subMenu.cnName}</span>
                </span>
              }>
              {subMenu.children.map(item => (
                <Menu.Item key={item.code}>
                  <a href={item.url}>
                    <span>
                      <span>{item.cnName}</span>
                    </span>
                  </a>
                </Menu.Item>
              ))}
            </SubMenu>
          ) : (
            <Menu.Item key={subMenu.code}>
              <a href={subMenu.url} target={subMenu.target?subMenu.target:null}>
                <span>
                  <span>{subMenu.cnName}</span>
                </span>
              </a>
            </Menu.Item>
          ),
        )}
      </Menu>
    </Sider>
  );
};
