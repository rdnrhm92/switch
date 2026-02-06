// @ts-ignore
/* eslint-disable */

declare namespace API {
  type Env = {
    data: EnvItem[],
    total: number,
    success: boolean,
  };

  type Driver = {
    driverConfig: Record<string, any>;
    driverType: string;
    name: string;
    usage: string;
    createTime: string;
    createdBy: string;
    updateBy: string;
    updateTime: string;
    id: number;
  };

  type EnvItem = {
    id: number;
    description: string;
    publish_order: number;
    drivers: Driver[];
    namespaces: NamespaceItem[];
    name: string;
    tag: string;
    createTime: string;
    createBy: string;
    updateTime: string;
    updateBy: string;
    publish: boolean;
  }

  type EnvCreateUpdate = {
    id: number;
    description: string;
    drivers: Driver[];
    name: string;
    tag: string;
    publish_order: number;
    select_namespace?: string;
  }

  type EnvPublish = {
    id: number;
  }
}
