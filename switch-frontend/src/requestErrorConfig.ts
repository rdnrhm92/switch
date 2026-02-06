import type {RequestOptions} from '@@/plugin-request/request';
import type {RequestConfig} from '@umijs/max';

// 与后端约定的响应数据格式
export interface ResponseStructure<T = any> {
  statistics: any;
  trace: any;
  debug_info: any;
  data: T;
  code?: number;
  message?: string;
}

/**
 * @name 错误处理
 * pro 自带的错误处理， 可以在这里做自己的改动
 * @doc https://umijs.org/docs/max/request#配置
 */
export const errorConfig: RequestConfig = {
  // 错误处理： umi@3 的错误处理方案。
  errorConfig: {
    // 错误抛出
    errorThrower: (res) => {
      const {data, code, message} =
        res as unknown as ResponseStructure;
      if (code != 0) {
        const error: any = new Error(message);
        error.name = 'BizError';
        error.info = {code, message, data};
        throw error; // 抛出自制的错误 所有的错误都在这里包装成自己的错误
      }
    },
    // 错误接收及处理
    errorHandler: (error: any, opts: any) => {
      if (opts?.skipErrorHandler) throw error;

      console.error('[API Error]', {
        type: error.name,
        message: error.message,
        status: error.response?.status,
        data: error.response?.data,
      });
    },
  },

  // 请求拦截器
  requestInterceptors: [
    (config: RequestOptions) => {
      // 拦截请求配置，进行个性化处理。
      const token = localStorage.getItem('Authorization');
      if (token) {
        if (config.headers) {
          config.headers.Authorization = `Bearer ${token}`;
        }
      }
      return config;
    },
  ],

  // 响应拦截器
  responseInterceptors: [
    (response) => {
      // 直接返回响应，错误由 errorThrower 处理
      return response;
    },
  ],
};
