import React from 'react';
import { Link } from 'react-router-dom';

export default () => {
  return (
    <div className="error-wrapper">
      <Link to="/">
        <img src="https://portal-static.tongdun.cn/static-public/assets/images/1.0.1/404.png" alt="404" />
      </Link>
    </div>
  );
};
