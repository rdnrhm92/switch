// @ts-ignore
/* eslint-disable */

declare namespace API {

  type SwitchModelList = {
    data: SwitchModel[],
    total: number,
    success: boolean,
  };

  // 版本信息
  type Version = {
    version: number,
  };

  // 下一个环境信息
  type NextEnvInfo = {
    envTag: string,         // 下一个环境标签
    envName: string,        // 下一个环境名称
    approvalStatus: string, // PENDING(审批中), APPROVED(已通过), REJECTED(已拒绝), ""(无审批记录)
    configStatus: string,   // PUBLISHED(已发布), PENDING(待处理)
    buttonText: string,     // 按钮显示文本
    buttonDisabled: boolean, // 按钮是否禁用
  };

  // 开关模型
  type SwitchModel = CommonModel & Version & {
    namespaceTag: string,
    name: string,
    currentEnvTag: string,
    rules: RuleNode,
    description: string,
    useCache: boolean,
    nextEnvInfo?: NextEnvInfo, // 下一个环境信息
    switchConfigs: SwitchConfig[],
  };

  // 规则节点
  type RuleNode = {
    id?:number,
    nodeType?: string,
    children?: RuleNode[],
    factor?: string,
    config?: any,
    description?: string,
  };

  // 创建/更新开关请求
  type CreateUpdateSwitchReq = {
    name: string, // 开关名字
    namespaceTag: string, // 命名空间Tag
    switchId: number, // 开关ID
    description: string, // 开关描述
    rules: RuleNode, // 开关规则
    createSwitchApproversReq: SwitchApproversReq[], // 开关审批人
  };

  // 开关审批人请求
  type SwitchApproversReq = {
    envTag: string,
    approverUsers: number[],
  };

  // 开关详情响应
  type SwitchDetailsResponse = {
    factor: SwitchModel,
    configs: SwitchConfig[],
    approvals: SwitchApproval[],
  };

  // 审核人员配置
  type SwitchApproval = CommonModel & {
    switchID: number,
    envTag: string,
    approverUsers: string,
  };

  // 开关配置
  type SwitchConfig = CommonModel & {
    switchID: number,
    envTag: string,
    isEnabled: boolean,
    configValue: RuleNode,
    status: 'PENDING' | 'PUBLISHED' | 'REJECTED',
  };

  // 提交开关推送请求
  type SubmitSwitchPushReq = {
    switchId: number,
    targetEnvTag: string,
  };

  type UserPermissionsListItem = {
    userInfo: UserInfo;
    userPermissions: Permission[];
  };


}


