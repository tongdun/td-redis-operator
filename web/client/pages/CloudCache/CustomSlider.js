import React from 'react';
import { InputNumber, Slider, Row, Col } from 'antd';

export default ({ value, onChange, tipFormatter, min, max, step, marks }) => {
  const triggerChange = changedValue => {
    if (onChange) {
      onChange(changedValue);
    }
  };

  const onInputChange = val => triggerChange(val * 1024);

  const formatter = val => `${val || ''} GB`;
  const parser = val => parseInt(val.substr(0, val.length - 3), 10);

  return (
    <Row gutter={24}>
      <Col span={16}>
        <Slider
          min={min}
          max={max}
          step={step}
          marks={marks}
          tipFormatter={tipFormatter}
          value={value}
          onChange={triggerChange}
          included={false}
        />
      </Col>
      <Col span={8}>
        <InputNumber
          min={min / 1024}
          max={max / 1024}
          value={value / 1024}
          step={step / 1024}
          onChange={onInputChange}
          formatter={formatter}
          parser={parser}
        />
      </Col>
    </Row>
  );
};
