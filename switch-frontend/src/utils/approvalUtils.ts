/**
 * 解析 approverUsers 字段
 */
export const parseApproverUsers = (approverUsers: string | number[] | null | undefined): number[] => {
  if (!approverUsers) {
    return [];
  }

  // 如果已经是数组，直接返回
  if (Array.isArray(approverUsers)) {
    return approverUsers;
  }

  if (approverUsers === '') {
    return [];
  }
  try {
    const parsed = JSON.parse(approverUsers);
    return Array.isArray(parsed) ? parsed : [];
  } catch (e) {
    console.error('解析审批人数据失败:', e);
    return [];
  }

};

/**
 * 序列化 approverUsers 字段
 */
export const stringifyApproverUsers = (approverUsers: number[] | string): string => {
  if (typeof approverUsers === 'string') {
    return approverUsers;
  }

  if (Array.isArray(approverUsers)) {
    return JSON.stringify(approverUsers);
  }

  return '[]';
};

/**
 * 将 approverUsers 从字符串转换为数组 用于表单回显
 */
export const processApprovalsForForm = (approvals: any[]): any[] => {
  return (approvals || []).map((approval: any) => ({
    ...approval,
    approverUsers: parseApproverUsers(approval.approverUsers)
  }));
};

/**
 * 将 approverUsers 从数组转换为字符串 用于提交后端
 */
export const processApprovalsForSubmit = (approvals: any[]): any[] => {
  return (approvals || []).map((approval: any) => ({
    ...approval,
    approverUsers: stringifyApproverUsers(approval.approverUsers)
  }));
};
