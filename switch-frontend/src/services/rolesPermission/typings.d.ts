// @ts-ignore
/* eslint-disable */

declare namespace API {
  type RolesList = {
    data: RolesListItem[],
    total: number,
    success: boolean,
  };

  type RolesListItem = CommonModel & {
    name:string,
    description:string,
    namespaceTag: string,
    permissions: RolesListPermissionItem[],
  };

  type RolesListPermissionItem = CommonModel & {
    name:string,
    description:string,
    namespaceTag: string,
  };

  type UpsertPermission = {
    id: number,
    name:string,
    description:string,
  };

}


