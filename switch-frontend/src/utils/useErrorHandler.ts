import {message} from 'antd';
import {useIntl} from '@umijs/max';
import {useCallback} from 'react';

interface ErrorHandlerOptions {
  showDetail?: boolean;
  duration?: number;
  key?: string;
  onError?: (error: any) => void;
}

/**
 * 从错误对象中提取错误信息
 * 统一处理各种错误格式
 */
export const extractErrorMessage = (error: any): string => {
  // 处理BizError
  if (error?.name === 'BizError' && error?.info?.message) {
    return error.info.message;
  }

  if (error?.response?.data?.message) {
    return error.response.data.message;
  }

  if (error?.message) {
    return error.message;
  }

  return '未知错误';
};

/**
 * 非 Hook 版本的错误处理函数
 * 用于不能使用 hooks 的地方
 */
export const handleErrorWithoutHook = (error: any, intlMsg :string): string => {
  let errorDetail = extractErrorMessage(error);
  if (intlMsg){
    errorDetail = `${intlMsg}\n详情: ${errorDetail}`
  }
  console.error('Error caught:', errorDetail);
  return errorDetail
};

/**
 * 统一错误处理 Hook
 */
export const useErrorHandler = () => {
  const intl = useIntl();

  const handleError = useCallback((
    error: any,
    messageId: string,
    options?: ErrorHandlerOptions
  ) => {
    const { showDetail = true, duration = 3, onError } = options || {};

    console.error('Error caught:', error);

    // 获取国际化消息
    const localizedMessage = intl.formatMessage({
      id: messageId,
      defaultMessage: '获取不到国际化消息key:' + messageId,
    });

    const errorDetail = extractErrorMessage(error);

    // 组合消息 包含了国际化的加后端响应的
    let fullMessage = localizedMessage;
    if (showDetail && errorDetail && errorDetail !== '未知错误') {
      fullMessage = `${localizedMessage}\n详情: ${errorDetail}`;
    }

    if (fullMessage.length > 200) {
      fullMessage = fullMessage.substring(0, 200) + '......';
    }

    message.error({
      content: fullMessage,
      duration,
      key: options?.key
    });

    if (onError) {
      onError(error);
    }

    return errorDetail;
  }, [intl]);

  return {
    handleError,
  };
};
