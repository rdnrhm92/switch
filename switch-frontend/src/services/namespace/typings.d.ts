// @ts-ignore
/* eslint-disable */

declare namespace API {
  type Namespace = {
    data: NamespaceItem[],
    total: number,
    success: boolean,
  };

  type NamespaceItem = {
    id: number;
    createTime: string;
    createBy: string;
    name: string;
    tag: string;
    description: string;
  }


  type NamespaceCreateUpdate = {
    id: number;
    name: string;
    tag: string;
    description: string;
  }

  type NameSpaceJoin = {
    namespaceTag: string;
  }

}
