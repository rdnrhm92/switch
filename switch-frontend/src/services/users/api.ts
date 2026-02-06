// @ts-ignore
/* eslint-disable */
import {request} from '@umijs/max';
import {ResponseStructure} from "@/requestErrorConfig";

/** 用户列表查询 POST /api/v1/users */
export async function userList(
  params: any,
  sort: any,
  filter: any,
) {
  const requestParams = {
    ...params,
    ...sort,
    ...filter,
  };
  if (requestParams.updatedAt) {
    if (Array.isArray(requestParams.updatedAt) && requestParams.updatedAt.length === 2) {
      requestParams.startTime = new Date(requestParams.updatedAt[0]).toISOString();
      requestParams.endTime = new Date(requestParams.updatedAt[1]).toISOString();
    }
    delete requestParams.updatedAt;
  }
  if (requestParams.current) {
    requestParams.page = requestParams.current;
    delete requestParams.current;
  }
  if (requestParams.pageSize) {
    requestParams.size = requestParams.pageSize;
    delete requestParams.pageSize;
  }
  console.info("[Request Params]", requestParams);
  return request<ResponseStructure<API.UserList>>('/api/v1/users', {
    data: requestParams,
    method: 'POST',
  }).then(response => response.data || {data: [], success: true, total: 0});
}

/** 分配角色给用户 POST /api/v1/assignRoles */
export async function assignUserRoles(
  userId: number,
  roleIds: number[]
) {
  return request<ResponseStructure<any>>('/api/v1/assignRoles', {
    data: {
      userId,
      roleIds
    },
    method: 'POST',
  });
}

