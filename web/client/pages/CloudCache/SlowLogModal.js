import { Modal, Table } from 'antd';
import React, { useEffect, useState } from 'react';
import * as api from '@/services/cache';

export default ({ instance = {}, visible, setVisible, onCancel }) => {
  const [dataSource, setDataSource] = useState([]);

  const loadDataSource = async () => {
    const res = await api.getSlowLog(instance.name);
    setDataSource(res || []);
  };

  useEffect(() => {
    if (visible) {
      loadDataSource();
    }
  }, [visible]);

  const columns = [
    {
      title: '执行时间',
      dataIndex: 'ts',
      key: 'ts',
    },
    {
      title: '阻塞时间',
      dataIndex: 'cost',
      key: 'cost',
    },
    {
      title: '执行命令',
      dataIndex: 'cmd',
      key: 'cmd',
    },
    {
      title: '来源',
      dataIndex: 'src',
      key: 'src',
    },
  ];
  return (
    <Modal
      title="慢查询分析"
      width={1200}
      visible={visible}
      onOk={() => setVisible(false)}
      onCancel={() => {
        setDataSource([]);
        setVisible(false);
      }}
      maskClosable={false}>
      <Table
        dataSource={dataSource}
        columns={columns}
        rowKey={r => r.ts + '-' + r.cost}
        pagination={{
          defaultCurrent: 1,
          showSizeChanger: true,
        }}
      />
    </Modal>
  );
};
