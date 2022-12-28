import React, { Fragment, useState, useEffect } from 'react';
import { Card, Row, Col, Image, Avatar, Space, Tooltip} from 'antd';
import { GithubOutlined } from '@ant-design/icons';
import Logo from '@/assets/td-redis-operator-logo.jpg';
import fetch from 'isomorphic-fetch';


export default () => {
  const [avatars, setAvatars] = useState([]);

  const loadContributors = async () => {
    const res = await fetch('https://api.github.com/repos/tongdun/td-redis-operator/contributors').then(response => {
      if(response.status === 200) {
        return response.json();
      }
      return [];
    });
    setAvatars(res);
  };

  useEffect(() => {
    loadContributors();
  }, []);

  return (
    <Card bodyStyle={{padding: "10px 50px"}}>
      <div style={{textAlign: 'center'}}>
        <Image height="243px" width="280px" preview={false} src={Logo} />
        <p>Version 0.2.0</p>
      </div>
      <div style={{margin:'50px 50px'}}>
        <p>
        一款强大的云原生redis-operator，经过大规模生产级运行考验，支持分布式集群、支持主备切换、友好的web管理界面等缓存集群解决方案。
        </p>
        <p>
        The powerful cloud-native redis-operator, which has passed the test of large-scale production-level operation, supports distributed clusters and active/standby switching,friendly web manager.
        </p>
      </div>

      <div style={{margin:'50px 50px', textAlign: 'center'}}>
        <p>
          <GithubOutlined /> Github 地址： 
          <a href='https://github.com/tongdun/td-redis-operator' target='_blank'>
          https://github.com/tongdun/td-redis-operator
          </a>
        </p>
        <Space>
          {avatars.map(a => {
            return <Tooltip key={a.id} title={a.login}><a href={a.html_url} target='_blank'><Avatar src={a.avatar_url} /></a></Tooltip>
          })
          }
        </Space>
      </div>

    </Card>
  );
};
