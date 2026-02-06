// @ts-ignore
/* eslint-disable */
import {request} from '@umijs/max';
import {ResponseStructure} from "@/requestErrorConfig";

/** 获取当前的用户 GET /api/currentUser */
export async function currentUser(options?: { [key: string]: any }) {
  return request<{
    data: ResponseStructure<API.CurrentUser>;
  }>('/api/v1/user', {
    method: 'GET',
    ...(options || {}),
  }).then(response => response.data);
}

/** 获取空间下拉选择列表 POST /api/v1/namespaces-like */
export async function selectNamespaceLike(body: API.SearchParams, options?: { [key: string]: any }) {
  return request<{
    data: API.AllNamespaceWithUser[];
  }>('/api/v1/namespaces-like', {
    method: 'POST',
    data: body,
    ...(options || {}),
  }).then(response => response.data);
}

/** 登录接口 POST /auth/login */
export async function login(body: API.LoginParams, options?: { [key: string]: any }) {
  return request<ResponseStructure<API.LoginResult>>('/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 注册接口 POST /auth/register */
export async function register(body: API.LoginParams, options?: { [key: string]: any }) {
  return request<ResponseStructure<API.LoginResult>>('/auth/register', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 刷新token POST /auth/refresh-token */
export async function refreshUserToken(body: API.RefreshToken, options?: { [key: string]: any }) {
  return request<ResponseStructure<API.LoginResult>>('/auth/refresh-token', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  }).then(response => response.data);
}
