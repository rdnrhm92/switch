// @ts-ignore
/* eslint-disable */
import {request} from '@umijs/max';
import {ResponseStructure} from "@/requestErrorConfig";

/** 角色权限列表查询 POST /api/v1/roles-permission */
export async function rolesPermissionList(
  params: any,
  sort: any,
  filter: any,
) {
  const requestParams = {
    ...params,
    ...sort,
    ...filter,
  };
  if (requestParams.current) {
    requestParams.page = requestParams.current;
    delete requestParams.current;
  }
  if (requestParams.pageSize) {
    requestParams.size = requestParams.pageSize;
    delete requestParams.pageSize;
  }
  console.info("[Request Params]", requestParams);
  return request<ResponseStructure<API.RolesList>>('/api/v1/roles-permission', {
    data: requestParams,
    method: 'POST',
  }).then(response => response.data || {data: [], success: true, total: 0});
}

/** 获取所有权限列表 POST /api/v1/permissions */
export async function allPermissions() {
  return request<ResponseStructure<API.RolesListPermissionItem[]>>('/api/v1/permissions', {
    method: 'POST',
  });
}

/** 新增角色 POST /api/v1/roles-create */
export async function insertRole(params: Partial<API.UpsertPermission>) {
  return request<ResponseStructure<any>>('/api/v1/roles-create', {
    method: 'POST',
    data: params,
  });
}

/** 更新角色 POST /api/v1/roles-create */
export async function updateRole(params: Partial<API.UpsertPermission>) {
  return request<ResponseStructure<any>>('/api/v1/roles-edit', {
    method: 'POST',
    data: params,
  });
}


/** 新增权限 POST /api/v1/role */
export async function insertPermission(params: Partial<API.UpsertPermission>) {
  return request<ResponseStructure<any>>('/api/v1/permissions-create', {
    method: 'POST',
    data: params,
  });
}

/** 更新权限 POST /api/v1/permissions-edit */
export async function updatePermission(params: Partial<API.UpsertPermission>) {
  return request<ResponseStructure<any>>('/api/v1/permissions-edit', {
    method: 'POST',
    data: params,
  });
}
