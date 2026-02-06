// @ts-ignore
/* eslint-disable */
import {request} from '@umijs/max';
import {ResponseStructure} from "@/requestErrorConfig";

/** 审批单查询 POST /api/v1/approvals/my-requests */
export async function myRequests(
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
  if (requestParams.approvableType && typeof requestParams.approvableType === 'string') {
    requestParams.approvableType = parseInt(requestParams.approvableType, 10);
  }
  if (requestParams.status && typeof requestParams.status === 'string') {
    requestParams.status = parseInt(requestParams.status, 10);
  }
  console.info("[Request Params]", requestParams);
  return request<ResponseStructure<API.ApprovalDetailView>>('/api/v1/approvals/my-requests', {
    data: requestParams,
    method: 'POST',
  }).then(response => response.data || {data: [], success: true, total: 0});
}


/** 所有审批单查询 POST /api/v1/approvals/all-requests */
export async function allRequests(
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
  if (requestParams.approvableType && typeof requestParams.approvableType === 'string') {
    requestParams.approvableType = parseInt(requestParams.approvableType, 10);
  }
  if (requestParams.status && typeof requestParams.status === 'string') {
    requestParams.status = parseInt(requestParams.status, 10);
  }
  console.info("[Request Params]", requestParams);
  return request<ResponseStructure<API.ApprovalDetailView>>('/api/v1/approvals/all-requests', {
    data: requestParams,
    method: 'POST',
  }).then(response => response.data || {data: [], success: true, total: 0});
}

/** 审批 POST /api/v1/approvals/{id}/approve */
export async function approveRequest(params: API.ApprovalForm) {
  return request<ResponseStructure<any>>('/api/v1/approvals/' + params.id + '/approve', {
    method: 'POST',
    data: params,
  });
}

/** 模糊查询所有用户 POST /api/v1/user-like */
export async function userLike(params: API.SelUserLike) {
  return request<ResponseStructure<API.CurrentUser[]>>('/api/v1/user-like', {
    method: 'POST',
    data: params,
  });
}


