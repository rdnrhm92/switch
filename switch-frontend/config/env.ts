/**
 * 自定义环境配置 不关心跨域可以直连，并取消proxy
 */
export default {
  dev: {
    API_URL: 'http://localhost:8000',
  },
  test: {
    API_URL: 'http://localhost:8000', // 请替换为您的测试环境 API 地址
  },
  pre: {
    API_URL: 'http://localhost:8000', // 请替换为您的预发环境 API 地址
  },
  pro: {
    API_URL: 'http://localhost:8000', // 请替换为您的生产环境 API 地址
  },
};
