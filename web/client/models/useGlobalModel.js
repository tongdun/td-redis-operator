import { useState } from 'react';
import { createModel } from 'hox';

/**
 * 需要使用全局变量时，或多个组件共享状态时，使用model
 */
export default createModel(() => {
  const [user, setUser] = useState({});
  const [loading, setLoading] = useState(false);

  return {
    user,
    setUser,
    loading,
    setLoading,
  };
});
