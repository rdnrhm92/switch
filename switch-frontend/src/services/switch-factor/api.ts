// @ts-ignore
/* eslint-disable */
import {request} from '@umijs/max';
import {ResponseStructure} from "@/requestErrorConfig";

/** 开关因子的列表 POST /api/v1/switch-factors */
export async function switchFactors(
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
  return request<ResponseStructure<API.SwitchFactor>>('/api/v1/switch-factors', {
    data: requestParams,
    method: 'POST',
  }).then(response => response.data || {data: [], success: true, total: 0});
}

/** 模糊全部开关因子的列表 POST /api/v1/switch-factors-like */
export async function switchFactorsLike(
  params: any,
  sort: any,
  filter: any,
) {
  const requestParams = {
    ...params,
    ...sort,
    ...filter,
  };
  console.info("[Request Params]", requestParams);
  return request<ResponseStructure<API.SwitchFactorItem[]>>('/api/v1/switch-factors-like', {
    data: requestParams,
    method: 'POST',
  }).then(response => response.data || []);
}


/** 新建因子信息 POST /api/v1/factor-meta-create */
export async function createSwitchFactor(data: Partial<API.NamespaceCreateUpdate>) {
  return request<ResponseStructure<null>>('/api/v1/switch-factor-create', {
    method: 'POST',
    data,
  });
}

/** 更新因子信息 POST /api/v1/factor-meta-edit */
export async function updateSwitchFactor(data: Partial<API.NamespaceCreateUpdate>) {
  return request<ResponseStructure<null>>('/api/v1/switch-factor-edit', {
    method: 'POST',
    data,
  });
}
