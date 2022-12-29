import React, { Fragment, useEffect, useState } from 'react';
import {
  Button,
  Descriptions,
  Drawer,
  Space,
  Badge,
  Table,
  Tabs,
  Input,
  Col,
  Row,
  message,
  InputNumber,
  Modal,
} from 'antd';
import * as api from '@/services/cache';
import styles from './style.less';
import cn from 'classnames';
import UpgradeModal from './UpgradeModal';
import SlowLogModal from './SlowLogModal';
import { ExclamationCircleOutlined } from '@ant-design/icons';

const { TabPane } = Tabs;
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

const operStatusMap = {
  0: (
    <div>
      <Badge status="processing" />
      执行中
    </div>
  ),
  1: (
    <div>
      <Badge status="success" />
      成功
    </div>
  ),
  2: (
    <div>
      <Badge status="error" />
      失败
    </div>
  ),
};

export default ({ visible, setVisible, instance = {} }) => {
  const [operLog, setOperLog] = useState([]);
  const [config, setConfig] = useState([]);
  const [slowLogVisible, setSlowLogVisible] = useState(false);
  const [visibleUpgrade, setVisibleUpgrade] = useState(false);

  const getOperLog = async () => {
    const res = await api.getOperLog(instance.name);
    setOperLog(res || []);
  };

  const getConfig = async () => {
    const res = await api.getConfig(instance.name);
    setConfig(res || []);
  };

  const saveConfig = () => {
    const passed = config.every(c => {
      if (c.value_type === 'string') {
        if (c.value_range.indexOf(c.value) === -1) {
          message.error(c.name + '取值无效，请检查');
          return false;
        }
      }
      if (c.value_type === 'int') {
        if (parseInt(c.value) < c.value_range[0] || parseInt(c.value) > c.value_range[1]) {
          message.error(c.name + '取值无效，请检查');
          return false;
        }
      }
      return true;
    });
    if (passed) {
      api.updateConfig(instance.name, config).then(r => {
        if (r === '') {
          message.success('更新成功');
          getConfig();
        }
      });
    }
  };

  const getPhase = phase => {
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
  };

  const showSecret = () => message.info(instance.secret, 5);

  const confirmDelete = () => {
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

  const confirmFlush = () => {
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

  const onDelete = async redis => {
    await api.deleteInstance(redis);
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

  useEffect(() => {
    if (instance.name) {
      getOperLog();
      getConfig();
    }
  }, [instance]);

  const operLogColumns = [
    {
      title: '操作时间',
      dataIndex: 'ts',
      key: 'ts',
    },
    {
      title: '操作人',
      dataIndex: 'realname',
      key: 'realname',
    },
    {
      title: '操作对象',
      dataIndex: 'resource',
      key: 'resource',
    },
    {
      title: '操作名称',
      dataIndex: 'oper',
      key: 'oper',
    },
    {
      title: '操作结果',
      dataIndex: 'status',
      key: 'status',
      render: status => {
        return operStatusMap[status];
      },
    },
  ];

  const configColumns = [
    {
      title: '参数名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '字符类型',
      dataIndex: 'value_type',
      key: 'value_type',
    },
    {
      title: '参数值',
      dataIndex: 'value',
      key: 'value',
      render: (text, record) => {
        if (record.value_type === 'string') {
          return (
            <Input
              defaultValue={text}
              onChange={e => {
                record.value = e.target.value;
              }}
            />
          );
        }
        if (record.value_type === 'int') {
          return (
            <InputNumber
              min={record.value_range[0]}
              max={record.value_range[1]}
              defaultValue={record.value}
              onChange={num => (record.value = num)}
            />
          );
        }
      },
    },
    {
      title: '参数范围',
      dataIndex: 'value_range',
      key: 'value_range',
      width: 400,
      render: (text, record) => {
        if (record.value_type === 'string') {
          return record.value_range.join(',');
        }
        if (record.value_type === 'int') {
          return record.value_range.join('-');
        }
      },
    },
    {
      title: '类型',
      dataIndex: 'redis_conf_type',
      key: 'redis_conf_type',
    },
  ];

  return (
    <div>
      <Drawer title="redis实例详情" width="70%" placement="right" onClose={() => setVisible(false)} visible={visible}>
        <Tabs defaultActiveKey="1" tabPosition="left">
          <TabPane tab="概览" key="1">
            <Space style={{ marginBottom: 16, float: 'right' }}>
              <Button onClick={showSecret} type="primary">
                查看密码
              </Button>
              <Button onClick={() => setSlowLogVisible(true)} type="primary">
                慢查询分析
              </Button>
              <Button onClick={() => setVisibleUpgrade(true)} type="primary">
                升降级
              </Button>
              <Button onClick={confirmFlush} type="primary">
                清理数据
              </Button>
              <Button onClick={confirmDelete} type="danger">
                删除
              </Button>
            </Space>
            <Descriptions bordered>
              <Descriptions.Item span={1} label="实例名称">
                {instance.name ? instance.name.substr(8) : ''}
              </Descriptions.Item>
              <Descriptions.Item span={2} label="状态">
                {getPhase(instance.phase)}
              </Descriptions.Item>
              <Descriptions.Item label="实例地址">{instance.clusterIp}</Descriptions.Item>
              <Descriptions.Item span={2} label="外部IP和端口">
                {instance.externalIp ? instance.externalIp : '无'}
              </Descriptions.Item>
              <Descriptions.Item span={1} label="规格">
                {`${(instance.capacity / 1024).toFixed(2)}GB`}
              </Descriptions.Item>
              <Descriptions.Item span={2} label="已使用">
                {instance.memoryused}
              </Descriptions.Item>
              <Descriptions.Item span={1} label="机房">
                {dcMap[instance.dc]}
              </Descriptions.Item>
              <Descriptions.Item span={1} label="环境">
                {envMap[instance.env]}
              </Descriptions.Item>
              <Descriptions.Item span={1} label="机型">
                {kindMap[instance.kind]}
              </Descriptions.Item>
              <Descriptions.Item span={1} label="负责人">
                {instance.realname}
              </Descriptions.Item>
              <Descriptions.Item span={1} label="创建时间">
                {instance.gmtCreate}
              </Descriptions.Item>
            </Descriptions>
          </TabPane>
         {/* <TabPane tab="操作日志" key="2">
            <Table
              dataSource={operLog}
              columns={operLogColumns}
              rowKey={r => r.ts + '-' + r.cost}
              pagination={{
                defaultCurrent: 1,
                showSizeChanger: true,
              }}
            />
          </TabPane>*/}
          <TabPane tab="配置文件" key="3">
            <Row justify="end" style={{ marginBottom: 10 }}>
              <Col span={2}>
                <Button onClick={saveConfig} type="primary">
                  保存
                </Button>
              </Col>
            </Row>
            <Table
              dataSource={config}
              columns={configColumns}
              rowKey={r => r.name}
              pagination={{
                defaultCurrent: 1,
                showSizeChanger: true,
              }}
            />
          </TabPane>
        </Tabs>
        <UpgradeModal visible={visibleUpgrade} setVisible={setVisibleUpgrade} instance={instance} />
        <SlowLogModal visible={slowLogVisible} setVisible={setSlowLogVisible} instance={instance} />
      </Drawer>
    </div>
  );
};
