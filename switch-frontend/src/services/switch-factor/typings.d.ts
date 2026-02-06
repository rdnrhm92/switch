// @ts-ignore
/* eslint-disable */

declare namespace API {
  type SwitchFactor = {
    data: SwitchFactorItem[],
    total: number,
    success: boolean,
  };

  type SwitchFactorItem = {
    id: number;
    createTime: string;
    createBy: string;
    updateTime: string;
    updateBy: string;
    factor: string;
    name: string;
    namespaceTag: string;
    description: string;
    jsonSchema: string;
  }

}
