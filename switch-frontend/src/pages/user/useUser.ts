import {useModel} from '@umijs/max';
import {flushSync} from 'react-dom';
import {useCallback} from 'react';

export const useUserActions = () => {
  const {initialState, setInitialState} = useModel('@@initialState');

  /**
   * 刷新用户信息
   */
  const refreshUserInfo = useCallback(async () => {
    const userInfo = await initialState?.fetchUserInfo?.();
    if (userInfo) {
      flushSync(() => {
        setInitialState((s) => ({
          ...s,
          currentUser: userInfo,
        }));
      });
    }
  }, [initialState, setInitialState]);


  return {
    refreshUserInfo,
  };
};

/**
 * 刷新token
 */
export const refreshToken = (token: string | null | undefined): void => {
  if (token) {
    localStorage.setItem('Authorization', token);
  } else {
    console.error("保存 Token 失败：未能获取到有效的 Authorization");
  }
};
