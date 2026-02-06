// @ts-ignore
/* eslint-disable */

declare namespace API {
  type NamespaceMembers = {
    id: number;
    namespaceTag: string;
    createTime: string;
    updateTime: string;
    name: string;
    description: string;
    userRoles: IndexRole[];
  };

  type IndexRole = {
    role: RolesListItem;
  };

  type Permission = {
    name: string;
    description: string;
  };

  type CurrentUser = {
    id?: number;
    username?: string;
    is_super_admin?: boolean;
    namespaceMembers?: NamespaceMembers[];
    avatar?: string;
    select_namespace?:string;
  };

  type SelUserLike = {
    username?: string;
  };

  type LoginResult = API.CurrentUser & {
    token?: string;
  };

  //包含用户 & 加入状态
  type AllNamespaceWithUser = API.NamespaceItem & {
    //PENDING(待审核), REJECTED(拒绝), APPROVED(批准)
    status: string;
  }

  type PageParams = {
    current?: number;
    pageSize?: number;
  };

  type LoginParams = {
    username?: string;
    password?: string;
    autoLogin?: boolean;
  };

  type SearchParams = {
    search?: string;
    all?: boolean;
  };

  type RefreshToken = {
    userId?: number;
    selectNamespace?: string;
  };

}
