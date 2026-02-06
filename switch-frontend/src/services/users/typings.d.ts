// @ts-ignore
/* eslint-disable */

declare namespace API {
  type UserList = {
    data: UserListItem[],
    total: number,
    success: boolean,
  };

  type UserListItem = {
    userInfo: UserInfo;
    userRoles: UserRole[];
  };

  type AssignRolesReq = {
    userId: number,
    roleIds: number[],
  };

  type UserInfo = CommonModel & {
    username: string;
  }

  type UserRole = CommonModel & {
    name: string;
    description: string;
  }

}


