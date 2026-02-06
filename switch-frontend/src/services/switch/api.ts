// @ts-ignore
/* eslint-disable */
import {request} from '@umijs/max';
import {ResponseStructure} from "@/requestErrorConfig";

/** 开关列表查询 POST /api/v1/switches */
export async function switches(
  params: any,
  sort: any,
  filter: any,
) {
  const requestParams = {
    ...params,
  };
  if (requestParams.current) {
    requestParams.page = requestParams.current;
    delete requestParams.current;
  }
  if (requestParams.pageSize) {
    requestParams.size = requestParams.pageSize;
    delete requestParams.pageSize;
  }
  console.info("[Switch List Request Params]", requestParams);
  return request<ResponseStructure<API.SwitchModelList>>('/api/v1/switches', {
    data: requestParams,
    method: 'POST',
  }).then(response => response.data || {data: [], success: true, total: 0});
}

/** 更新开关 POST /api/v1/switch-edit */
export async function updateSwitch(
  params: API.CreateUpdateSwitchReq,
) {
  console.info("[update Switch Request Params]", params);
  return request<ResponseStructure<API.SwitchModel>>('/api/v1/switch-edit', {
    data: params,
    method: 'POST',
  });
}

/** 创建开关 POST /api/v1/switch-create */
export async function createSwitch(
  params: API.CreateUpdateSwitchReq,
) {
  console.info("[create Switch Request Params]", params);
  return request<ResponseStructure<API.SwitchModel>>('/api/v1/switch-create', {
    data: params,
    method: 'POST',
  });
}

/** 获取开关详情 GET /api/v1/switch */
export async function getSwitchDetails(
  switchId: number,
) {
  return request<ResponseStructure<API.SwitchDetailsResponse>>(`/api/v1/switch/${switchId}`, {
    method: 'GET',
  });
}

/** 推送开关变更 POST /api/v1/switch-submit-push */
export async function pushSwitchChange(
  params: API.SubmitSwitchPushReq,
) {
  console.info("[Push Switch Change Request Params]", params);
  return request<ResponseStructure<string>>('/api/v1/switch-submit-push', {
    data: params,
    method: 'POST',
  });
}

/** 用户列表查询 POST /api/v1/users-permission */
export async function userPermissionsList(
  params: any,
) {
  const requestParams = {
    ...params,
  };
  console.info("[Request Params]", requestParams);
  return request<ResponseStructure<API.UserPermissionsListItem>>('/api/v1/users-permission', {
    data: requestParams,
    method: 'POST',
  }).then(response => response.data || []);
}
