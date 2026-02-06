/**
 * @see https://umijs.org/docs/max/access#access
 * */
interface AccessPermissions {
  canSuperAdmin: boolean;

  [key: string]: boolean;
}

/**
 * @see https://umijs.org/docs/max/access#access
 * */
export default function access(
  initialState: { currentUser?: API.CurrentUser } | undefined,
): AccessPermissions {
  const {currentUser} = initialState ?? {};

  const permissions: { [key: string]: boolean } = {};

  const selectedNamespaceMember = currentUser?.namespaceMembers?.find(
    (member) => member.namespaceTag === currentUser.select_namespace,
  );
  // <Access accessible={access['{permissions}']}>
  //   <Button>Add User</Button>
  // </Access>
  if (selectedNamespaceMember?.userRoles) {
    selectedNamespaceMember?.userRoles?.forEach((role) => {
      role?.role?.permissions?.forEach((perm) => {
        permissions[perm.name] = true;
      });
    });
  }

  return {
    canSuperAdmin: !!(currentUser && currentUser.is_super_admin),
    ...permissions,
  };
}
