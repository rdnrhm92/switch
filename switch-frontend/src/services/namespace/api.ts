// @ts-ignore
/* eslint-disable */
import {request} from '@umijs/max';
import {ResponseStructure} from '@/requestErrorConfig';

/** 获取命名空间列表 POST /api/v1/namespaces */
export async function namespaceList(
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
  const response = await request<ResponseStructure<API.Namespace>>('/api/v1/namespaces', {
    data: requestParams,
    method: 'POST',
  });

  return response.data || {data: [], success: true, total: 0};
}


/** 新建命名空间 POST /api/v1/namespace-create */
export async function createNamespace(data: Partial<API.NamespaceCreateUpdate>) {
  return request<ResponseStructure<API.NamespaceItem>>('/api/v1/namespace-create', {
    method: 'POST',
    data,
  });
}

/** 更新命名空间 POST /api/rule */
export async function updateNamespace(data: Partial<API.NamespaceCreateUpdate>) {
  return request<ResponseStructure<API.NamespaceItem>>('/api/v1/namespace-edit', {
    method: 'POST',
    data,
  });
}

/** 申请加入空间 POST /api/v1/approvals/applyToJoin */
export async function applyToJoin(data: Partial<API.NameSpaceJoin>) {
  return request<ResponseStructure<null>>('/api/v1/approvals/applyToJoin', {
    method: 'POST',
    data,
  });
}
