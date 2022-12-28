import React, { useState } from 'react';
import { Modal, Form, Input, Radio, Switch } from 'antd';
import Slider from './CustomSlider';
import * as api from '@/services/cache';

const layout = {
  labelCol: { span: 6 },
  wrapperCol: { span: 16 },
};

export default ({ visible, setVisible, onRefresh }) => {
  const [type, setType] = useState('standby');

  const [form] = Form.useForm();

  const onValuesChange = changedValues => {
    if (changedValues.kind) {
      setType(changedValues.kind);
      form.setFieldsValue({
        capacity: changedValues.kind === 'standby' ? 1024 : 32768,
      });
    }
  };

  const tipFormatter = value => `${value / 1024}GB`;

  const onOk = () => {
    form.validateFields().then(values => {
      onCreate(values);
    });
  };

  const onCreate = async ({ name, kind, netMode, ...others }) => {
    await api.createInstance({
      name: `${kind}-${name}`,
      netMode: netMode ? 'NodePort' : 'ClusterIP',
      kind,
      ...others,
    });
    onRefresh();
    setVisible(false);
  };

  const afterClose = () => {
    form.resetFields();
    setType('standby');
  };

  return (
    <Modal
      visible={visible}
      title="创建Redis实例"
      onOk={onOk}
      onCancel={() => setVisible(false)}
      afterClose={afterClose}
      maskClosable={false}
      width={600}>
      <Form
        form={form}
        {...layout}
        onValuesChange={onValuesChange}
        initialValues={{
          dc: 'hz',
          env: 'production',
          kind: 'standby',
          capacity: type === 'standby' ? 1024 : 32768,
        }}>
        <h3>地域</h3>
        <Form.Item name="dc" label="机房" rules={[{ required: true, message: '机房不能为空' }]}>
          <Radio.Group>
            <Radio.Button value="hz">杭州</Radio.Button>
          </Radio.Group>
        </Form.Item>
        <Form.Item name="env" label="环境" rules={[{ required: true, message: '环境不能为空' }]}>
          <Radio.Group>
            <Radio.Button value="production">生产</Radio.Button>
          </Radio.Group>
        </Form.Item>
        <h3>基本配置</h3>
        <Form.Item name="kind" label="机型" rules={[{ required: true, message: '机型不能为空' }]}>
          <Radio.Group>
            <Radio.Button value="standby">主备</Radio.Button>
            <Radio.Button value="cluster">分布式</Radio.Button>
          </Radio.Group>
        </Form.Item>
        <Form.Item name="capacity" label="实例容量" rules={[{ required: true, message: '实例容量不能为空' }]}>
          {type === 'standby' ? (
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
        <h3>管理配置</h3>
        <Form.Item name="name" label="实例名称" rules={[{ required: true, message: '实例名称不能为空' }]}>
          <Input placeholder="请输入实例名称" />
        </Form.Item>
        <Form.Item label="登录密码" name="secret" rules={[{ required: true, message: '密码不能为空' }]}>
          <Input.Password placeholder="请输入密码" type="secret" autoComplete="new-pwd" />
        </Form.Item>
        <Form.Item
          name="confirm"
          label="确认密码"
          dependencies={['secret']}
          hasFeedback
          rules={[
            {
              required: true,
              message: '请输入确认密码!',
            },
            ({ getFieldValue }) => ({
              validator(_rule, value) {
                if (!value || getFieldValue('secret') === value) {
                  return Promise.resolve();
                }

                return Promise.reject('两次输入的密码不一致!');
              },
            }),
          ]}>
          <Input.Password placeholder="请再次输入密码" autoComplete="new-pwd" />
        </Form.Item>
        <h3>网络设置</h3>
        <Form.Item label="是否开启外部IP" name="netMode" valuePropName="checked">
          <Switch checkedChildren="是" unCheckedChildren="否" />
        </Form.Item>
      </Form>
    </Modal>
  );
};
