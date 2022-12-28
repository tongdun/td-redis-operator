import React, { useState, useEffect } from 'react';
import { Layout, Spin, Drawer, Menu, message } from 'antd';
import * as icons from '@ant-design/icons';
import axios from 'axios';
import store from 'store';
import sortBy from 'lodash/sortBy';
import Sider from './Sider';
import Header from './Header';
import './style.less';

const { Content } = Layout;
const { SubMenu } = Menu;

function getMenus(tree) {
  return sortBy(tree.children, item => -item.ext.priority).map(group => ({
    id: group.id,
    code: group.name,
    cnName: group.ext.cnName,
    enName: group.ext.enName,
    icon: group.ext.icon,
    url: group.ext.url,
    priority: group.ext.priority,
    isMenu: group.ext.isMenu,
    target: group.ext.target || '_self',
    children:
      group.ext.showChildren !== 'false'
        ? group.children.map(menu => ({
            id: menu.id,
            code: menu.name,
            cnName: menu.ext.cnName,
            enName: menu.ext.enName,
            url: menu.ext.url,
            icon: menu.ext.icon,
            target: menu.ext.target || '_self',
          }))
        : [],
  }));
}

export default ({
  children,
  token,
  solution,
  withPerms = true,
  withRealms = false,
  luc,
  onUserChange,
  collapsible,
  placeholder,
  className,
  customMenu,
  showSider = true,
  defaultOpenProductionKeys,
  sysTitle, // 子系统名称
  headerItems, // header 菜单
}) => {
  const [menus, setMenus] = useState(customMenu || []);
  const [productions, setProductions] = useState([]);
  const [user, setUser] = useState({});
  const [collapsed, setCollapsed] = useState(solution ? store.get(`${solution}_collapsed`) || false : false);
  const [showProductions, setShowProductions] = useState(false);
  const [loading, setLoading] = useState(false);
  const group = (user.groups || [])[0];

  const getUserInfo = async () => {
    setLoading(true);
    try {
      const res = await axios.post(`${luc}/oauth/user/`, {
        with_groups: solution,
        with_perms: withPerms,
        with_realms: withRealms,
        token,
      });
      if (!res.data.msg) {
        setUser(res.data || {});
        window.UserInfo = res.data || {};
        setProductions(getMenus(res.data.groups[0]));
        if (!customMenu) {
          if (showSider && res.data.groups.length > 1) {
            setMenus(getMenus(res.data.groups[1]));
          } else {
            setMenus([]);
          }
        }
        if (onUserChange) {
          onUserChange(res.data || {});
        }
      } else {
        message.error('用户不存在或token已失效，请重新登录');
        window.location.replace('/logout');
      }
    } catch (e) {
      message.error('获取用户信息请求异常或token已失效，请重新登录');
      window.location.replace('/logout');
    }
    setLoading(false);
  };

  const toggle = () => {
    store.set(`${solution}_collapsed`, !collapsed);
    setCollapsed(!collapsed);
  };

  const toggleProductions = () => setShowProductions(!showProductions);
  const onClose = () => setShowProductions(false);
  const openProduction = (url, target) => () => {
    window.open(url, target);
    setShowProductions(false);
  };

  useEffect(() => {
    if (token && luc !== undefined && luc !== null) {
      getUserInfo();
    }
  }, [token, luc]);

  useEffect(() => {
    if (customMenu) {
      setMenus(customMenu);
    }
  }, [customMenu]);

  return (
    <Spin spinning={loading}>
      <Layout className={className}>
        <Header
          user={user}
          group={group}
          toggle={toggle}
          collapsed={collapsed}
          placeholder={placeholder}
          productions={productions}
          toggleProductions={toggleProductions}
          collapsible={showSider && menus.length > 0 && collapsible}
          headerItems={headerItems}
        />
        <Layout>
          {showSider && menus.length > 0 && (
            <Sider title={sysTitle} menus={menus} collapsible={collapsible} collapsed={collapsed} />
          )}
          <Content className="portal-content">{children}</Content>
          <Drawer
            title="产品与服务"
            visible={showProductions}
            closable={false}
            placement="left"
            onClose={onClose}
            bodyStyle={{ padding: '0' }}
            className="portal-productions">
            <Menu
              mode="inline"
              theme="dark"
              style={{ paddingLeft: 0, marginBottom: 0 }}
              defaultOpenKeys={defaultOpenProductionKeys}>
              {productions.map(subMenu =>
                !subMenu.isMenu ? (
                  <SubMenu
                    key={subMenu.code}
                    title={
                      <span>
                        {icons[subMenu.icon] ? React.createElement(icons[subMenu.icon]) : undefined}
                        <span>{subMenu.cnName}</span>
                      </span>
                    }>
                    <div
                      style={{
                        height: '1px',
                        background: 'rgba(255,255,255,0.19)',
                        margin: '0 20px',
                      }}
                    />
                    {subMenu.children.map(item => (
                      <Menu.Item key={item.code}>
                        <a onClick={openProduction(item.url, item.target)}>
                          <span>
                            {icons[item.icon] ? React.createElement(icons[item.icon]) : undefined}
                            <span>{item.cnName}</span>
                          </span>
                        </a>
                      </Menu.Item>
                    ))}
                  </SubMenu>
                ) : (
                  <Menu.Item key={subMenu.code}>
                    <a onClick={openProduction(subMenu.url, subMenu.target)}>
                      <span>
                        {icons[subMenu.icon] ? React.createElement(icons[subMenu.icon]) : undefined}
                        <span>{subMenu.cnName}</span>
                      </span>
                    </a>
                  </Menu.Item>
                ),
              )}
            </Menu>
          </Drawer>
        </Layout>
      </Layout>
    </Spin>
  );
};
