import React, { useEffect } from 'react';
import { Modal, Form, Radio, Switch, message } from 'antd';
import Slider from './CustomSlider';
import * as api from '@/services/cache';
import '../../style.less';

const layout = {
  labelCol: { span: 6 },
  wrapperCol: { span: 16 },
};

export default ({ instance = {}, visible, setVisible, onRefresh }) => {
  const { kind, name, capacity, netMode } = instance;
  const [form] = Form.useForm();

  const tipFormatter = value => `${value / 1024}GB`;

  const onUpgrade = async ({ netMode, ...values }) => {
    await api.modifyInstance({
      ...instance,
      netMode: netMode ? 'NodePort' : 'ClusterIP',
      ...values,
    });
    message.success('操作成功');
    if (onRefresh) {
      onRefresh();
    }
    form.resetFields();
    setVisible(false);
  };

  const onOk = () => {
    form.validateFields().then(values => {
      onUpgrade(values);
    });
  };

  return (
    <Modal
      visible={visible}
      title="配置升降级"
      onOk={onOk}
      onCancel={() => {
        form.resetFields();
        setVisible(false);
      }}
      maskClosable={false}
      width={600}>
      <Form
        form={form}
        {...layout}
        initialValues={{
          capacity,
          netMode: netMode === 'NodePort',
        }}>
        <Form.Item label="实例名称">
          <span className="ant-form-text">{name ? name.substr(8) : ''}</span>
        </Form.Item>
        <Form.Item name="capacity" label="扩缩容容量" rules={[{ required: true, message: '扩缩容容量不能为空' }]}>
          {kind === 'standby' ? (
            <Radio.Group className="redis-quota-radio">
              <Radio.Button value={1024}>1GB</Radio.Button>
              <Radio.Button value={2048}>2GB</Radio.Button>
              <Radio.Button value={4096}>4GB</Radio.Button>
              <Radio.Button value={6144}>6GB</Radio.Button>
              <Radio.Button value={8192}>8GB</Radio.Button>
              <Radio.Button value={12288}>12GB</Radio.Button>
              <Radio.Button value={16384}>16GB</Radio.Button>
              <Radio.Button value={24576}>24GB</Radio.Button>
              <Radio.Button value={32768}>32GB</Radio.Button>
            </Radio.Group>
          ) : (
            <Slider
              min={32768}
              max={1048576}
              step={1024}
              marks={{
                32768: '32GB',
                1048576: '1024GB',
              }}
              tipFormatter={tipFormatter}
            />
          )}
        </Form.Item>
        <Form.Item label="是否开启外部IP" name="netMode" valuePropName="checked">
          <Switch checkedChildren="是" unCheckedChildren="否" />
        </Form.Item>
      </Form>
    </Modal>
  );
};
