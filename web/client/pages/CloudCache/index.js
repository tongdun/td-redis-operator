import React, { Fragment, useState, useEffect } from 'react';
import { Table, Space, Button, Card, Popconfirm, Dropdown, Input, Row, Col, Select, message, Menu, Modal } from 'antd';
import { DownOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import cn from 'classnames';
import * as api from '@/services/cache';
import CreateModal from './CreateModal';
import UpgradeModal from './UpgradeModal';
import styles from './style.less';
import Detail from './Detail';
import SlowLogModal from './SlowLogModal';

const { confirm } = Modal;

const dcMap = {
  hz: '杭州',
  sh: '上海',
};
const envMap = {
  production: '生产',
};
const kindMap = {
  standby: '主备',
  cluster: '分布式',
};

export default () => {
  const [dataSource, setDataSource] = useState([]);
  const [list, setList] = useState([]);
  const [visibleCreation, setVisibleCreation] = useState(false);
  const [visibleUpgrade, setVisibleUpgrade] = useState(false);
  const [instance, setInstance] = useState();
  const [filter, setFilter] = useState();
  const [filterKind, setFilterKind] = useState();
  const [filterPhase, setFilterPhase] = useState();
  const [detailVisible, setDetailVisible] = useState(false);
  const [slowLogVisible, setSlowLogVisible] = useState(false);

  const loadDataSource = async () => {
    const res = await api.getInstanceList();
    setDataSource(res || []);
  };

  const showSecret = secret => () => message.info(secret, 5);

  const onRefresh = async () => {
    await loadDataSource();
  };

  const confirmDelete = instance => () => {
    confirm({
      title: '确定要删除此实例么',
      icon: <ExclamationCircleOutlined />,
      content: '删除操作无法撤销',
      okText: '确定',
      okType: 'danger',
      cancelText: '取消',
      onOk() {
        onDelete(instance);
      },
    });
  };

  const confirmFlush = instance => () => {
    confirm({
      title: '确定要清空此实例数据么',
      icon: <ExclamationCircleOutlined />,
      content: '清空数据操作无法撤销',
      okText: '确定',
      okType: 'danger',
      cancelText: '取消',
      onOk() {
        onFlush(instance);
      },
    });
  };

  const onDelete = async instance => {
    await api.deleteInstance(instance);
    await loadDataSource();
  };

  const onFlush = async instance => {
    const res = await api.flush(instance);
    if (res.success) {
      message.success('清理完成');
    } else {
      message.error(res.message);
    }
  };

  const columns = [
    {
      title: '实例名称',
      dataIndex: 'name',
      key: 'name',
      width: 200,
      ellipsis: true,
      render: name => name.substr(8),
    },
    {
      title: 'IP和端口',
      dataIndex: 'clusterIp',
      key: 'clusterIp',
      width: 200,
      ellipsis: true,
    },
    {
      title: '外部IP和端口',
      dataIndex: 'externalIp',
      key: 'externalIp',
      width: 200,
      ellipsis: true,
      render: externalIp => externalIp || '无',
    },
    {
      title: '机房',
      dataIndex: 'dc',
      key: 'dc',
      width: 60,
      render: dc => dcMap[dc],
    },
    {
      title: '环境',
      dataIndex: 'env',
      key: 'env',
      width: 60,
      render: env => envMap[env],
    },
    {
      title: '机型',
      dataIndex: 'kind',
      key: 'kind',
      width: 80,
      render: kind => kindMap[kind],
    },
    {
      title: '状态',
      dataIndex: 'phase',
      key: 'phase',
      width: 120,
      render: phase => {
        if (phase === 'Ready') {
          return (
            <Fragment>
              <i className={cn(styles.circle, styles.green)} />
              运行
            </Fragment>
          );
        }

        if (phase === 'UpdateQuota') {
          return (
            <Fragment>
              <i className={cn(styles.circle, styles.yellow)} />
              升降配置中
            </Fragment>
          );
        }

        return (
          <Fragment>
            <i className={cn(styles.circle, styles.red)} />
            未就绪
          </Fragment>
        );
      },
    },
    {
      title: '实例容量',
      dataIndex: 'capacity',
      key: 'capacity',
      render: (capacity, record) => {
        if (capacity < 1024) {
          return `${record.memoryused}/${capacity}MB`;
        }
        return `${record.memoryused}/${(capacity / 1024).toFixed(2)}GB`;
      },
      width: 220,
      ellipsis: true,
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right',
      render: (_text, record) => (
        <Space size="middle">
          <Button
            type="link"
            block
            onClick={() => {
              setInstance(record);
              setDetailVisible(true);
            }}>
            详情
          </Button>
          <Dropdown
            overlay={
              <Menu>
                <Menu.Item>
                  <Button type="link" block onClick={showSecret(record.secret)}>
                    查看密码
                  </Button>
                </Menu.Item>
                <Menu.Item>
                  <Button
                    type="link"
                    block
                    onClick={() => {
                      setInstance(record);
                      setSlowLogVisible(true);
                    }}>
                    慢查询分析
                  </Button>
                </Menu.Item>
                <Menu.Item>
                  <Button
                    type="link"
                    block
                    onClick={() => {
                      setInstance(record);
                      setVisibleUpgrade(true);
                    }}>
                    升降级
                  </Button>
                </Menu.Item>
                <Menu.Item>
                  <Button type="link" block onClick={confirmFlush(record)}>
                    清理数据
                  </Button>
                </Menu.Item>
                <Menu.Item>
                  <Button type="link" block onClick={confirmDelete(record)}>
                    删除
                  </Button>
                </Menu.Item>
              </Menu>
            }>
            <a className="ant-dropdown-link" onClick={e => e.preventDefault()}>
              更多 <DownOutlined />
            </a>
          </Dropdown>
        </Space>
      ),
    },
  ];

  useEffect(() => {
    loadDataSource();
  }, []); // 第二个参数数组为空，则只会在组件mount时执行一次

  useEffect(() => {
    setList(
      dataSource.filter(
        o =>
          (!filter || o.name.substr(8).indexOf(filter) >= 0) &&
          (!filterKind || o.kind === filterKind) &&
          (!filterPhase ||
            (filterPhase === 'Ready' && o.phase === 'Ready') ||
            (filterPhase !== 'Ready' && o.phase !== 'Ready')),
      ),
    );
  }, [dataSource, filter, filterKind, filterPhase]);

  return (
    <Card
      title="实例列表"
      extra={
        <div style={{ width: 700 }}>
          <Row gutter={10}>
            <Col span={8}>
              <Input.Search placeholder="输入实例名称查找" onSearch={setFilter} />
            </Col>
            <Col span={6}>
              <Select
                options={[
                  { value: 'standby', label: '主备' },
                  { value: 'cluster', label: '分布式' },
                ]}
                style={{ width: '100%' }}
                placeholder="选择机型"
                allowClear
                value={filterKind}
                onChange={setFilterKind}
              />
            </Col>
            <Col span={6}>
              <Select
                options={[
                  { value: 'Ready', label: '运行' },
                  { value: 'NotReady', label: '未就绪' },
                ]}
                style={{ width: '100%' }}
                placeholder="选择状态"
                allowClear
                value={filterPhase}
                onChange={setFilterPhase}
              />
            </Col>
            <Col span={4}>
              <Button
                type="primary"
                block
                onClick={() => {
                  setInstance(undefined);
                  setVisibleCreation(true);
                }}>
                创建实例
              </Button>
            </Col>
          </Row>
        </div>
      }>
      <Table dataSource={list} columns={columns} rowKey={r => r.name} scroll={{ x: 1360 }} />
      <CreateModal visible={visibleCreation} setVisible={setVisibleCreation} onRefresh={onRefresh} />
      <UpgradeModal visible={visibleUpgrade} setVisible={setVisibleUpgrade} onRefresh={onRefresh} instance={instance} />
      <SlowLogModal visible={slowLogVisible} setVisible={setSlowLogVisible} instance={instance} />
      <Detail visible={detailVisible} setVisible={setDetailVisible} instance={instance} />
    </Card>
  );
};
