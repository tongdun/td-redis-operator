import React from 'react';
import { Layout, Menu, Row, Col, Dropdown } from 'antd';
import { LogoutOutlined, MenuOutlined } from '@ant-design/icons';
import Icon from '@ant-design/icons';
import './style.less';

const { Header } = Layout;

const AvatarSvg = () => (
  <svg
    width="24px"
    height="24px"
    viewBox="0 0 24 24"
    version="1.1"
    xmlns="http://www.w3.org/2000/svg"
    style={{ marginBottom: '-5px' }}>
    <path
      d="M12.0196588,23.9836685 C18.6403853,23.9836685 24.0098525,18.6142013 24.0098525,11.9912952 C24.0098525,5.36838906 18.6414868,-3.55271368e-15 12.020737,-3.55271368e-15 C5.39672925,-3.55271368e-15 0.0283636268,5.36836562 0.0283636268,11.9912952 C0.0283636268,18.6142013 5.39672925,23.9836685 12.0196588,23.9836685 Z"
      id="路径"
      fill="#A4BEFF"></path>
    <path
      d="M14.6345561,16.332019 C14.514556,16.28511 13.7705551,15.9480186 14.2483736,14.5571068 C15.4472852,13.3101952 16.3832863,11.3192827 16.3832863,9.3534601 C16.3832863,6.33163627 14.3694659,4.74873599 12.0425625,4.74873599 C9.6927394,4.74873599 7.70183865,6.33164799 7.70183865,9.3534601 C7.70183865,11.3192804 8.61383971,13.3341835 9.83675129,14.5811068 C10.3145698,15.8301981 9.451661,16.2851088 9.28475065,16.356019 C6.83892749,17.2440201 3.96110383,18.8509204 3.96110383,20.4578441 L3.96110383,20.8898446 C6.07201644,22.7836648 8.90182833,23.9836685 12.0196523,23.9836685 C15.1374762,23.9836685 17.9661982,22.7836671 20.1011229,20.8407664 L20.1011229,20.4098558 C20.1022135,18.7778539 17.2232989,17.1709536 14.6345772,16.3320307"
      id="路径"
      fill="#E8F2FE"
      fillRule="nonzero"></path>
  </svg>
);

const AvatarIcon = props => <Icon component={AvatarSvg} {...props} />;
const LogoIcon = props => <Icon component={LogoSvg} {...props} />;

export default ({ user, toggleProductions, group }) => {
  const menuList = [
    // {
    //   name: '退出登录',
    //   icon: <LogoutOutlined />,
    //   url: '/logout',
    // },
  ];

  const menu = (
    <Menu>
      {menuList.map(({ name, icon, url }) => (
        <Menu.Item key={name}>
          <a href={url}>
            <span style={{ marginRight: 5 }}>{icon}</span> {name}
          </a>
        </Menu.Item>
      ))}
    </Menu>
  );

  return (
    <Header className="portal-header">
      <Row type="flex" justify="space-between">
        <Col style={{ display: 'flex', flex: '0 0 250' }}>
          <Row type="flex" justify="flex-start">
            <Col style={{ display: 'flex', flex: '0 0 200' }}>
              <span
                style={{
                  color: 'white',
                  fontSize: '22px',
                  fontFamily: 'auto',
                  marginLeft: '10px',
                  fontWeight: 900,
                }}>
                TD-REDIS-OPERATOR
              </span>
            </Col>
          </Row>
        </Col>
        <Col style={{ display: 'flex', flex: '0 0 135px' }}>
          <Row style={{ width: '100%' }}>
            <Col span={24}>
              <Dropdown overlay={menu} trigger={['click']} placement="bottomCenter">
                <span style={{ cursor: 'pointer' }}>
                  <AvatarIcon />
                  <span
                    style={{
                      marginLeft: 8,
                      color: '#fff',
                    }}>
                    {user.alias || 'admin'}
                  </span>
                </span>
              </Dropdown>
            </Col>
          </Row>
        </Col>
      </Row>
    </Header>
  );
};
