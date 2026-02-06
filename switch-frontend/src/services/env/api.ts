// @ts-ignore
/* eslint-disable */
import {request} from '@umijs/max';
import {ResponseStructure} from '@/requestErrorConfig';

/** 获取环境+配置的列表 POST /api/v1/envs */
export async function envs(
  params: any,
  sort: any,
  filter: any,
) {
  const requestParams = {
    ...params,
    ...sort,
    ...filter,
  };
  console.log('打印',requestParams)
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
  return request<ResponseStructure<API.Env>>('/api/v1/envs', {
    data: requestParams,
    method: 'POST',
  }).then(response => response.data || {data: [], success: true, total: 0});
}


/** 新建环境 POST /api/v1/envs-create */
export async function createEnv(data: Partial<API.EnvCreateUpdate>) {
  return request<ResponseStructure<null>>('/api/v1/env-create', {
    method: 'POST',
    data,
  });
}

/** 更新环境 POST /api/v1/envs-edit */
export async function updateEnv(data: Partial<API.EnvCreateUpdate>) {
  return request<ResponseStructure<null>>('/api/v1/env-edit', {
    method: 'POST',
    data,
  });
}

/** 发布环境 POST /api/v1/envs-publish */
export async function publishEnv(data: Partial<API.EnvPublish>) {
  return request<ResponseStructure<null>>('/api/v1/envs-publish', {
    method: 'POST',
    data,
  });
}

/** 获取命名空间下的环境列表 GET /api/v1/namespaces/:namespaceTag/environments */
export async function namespaceEnvironments(namespaceTag: string) {
  return request<ResponseStructure<API.EnvItem[]>>(`/api/v1/namespaces/${namespaceTag}/environments`, {
    method: 'GET',
  });
}
