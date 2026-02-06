import {ActionType, PageContainer, ProColumns, ProTable} from '@ant-design/pro-components';
import {FormattedMessage, useAccess, useIntl} from '@umijs/max';
import {message, Tag, Tooltip} from 'antd';
import React, {useRef, useEffect} from 'react';
import {userList, assignUserRoles} from '@/services/users/api';
import {useNavigate, useLocation} from "react-router-dom";
import {useErrorHandler} from '@/utils/useErrorHandler';


const UserTableList: React.FC = () => {
  const actionRef = useRef<ActionType | null>(null);
  const [messageApi, contextHolder] = message.useMessage();
  const navigate = useNavigate();
  const location = useLocation();
  const access = useAccess() as Record<string, boolean>;
  const intl = useIntl();
  const { handleError } = useErrorHandler();

  const getJumpInfo = () => {
    const stepData = location.state as JumpData<API.AssignRolesReq>;
    if (stepData) {
      const {flag, data} = stepData;
      return {flag, data};
    }
    return null;
  }

  // 处理从角色权限页面返回的数据
  useEffect(() => {
    let jumpFromData = getJumpInfo();
    if (jumpFromData) {
      switch (jumpFromData.flag) {
        case 'rolesPermission':
          console.log('从用户列表页面过来，数据已处理');
          console.log('从用户列表页面过来，数据是：', jumpFromData.data);
          if (jumpFromData.data.userId && jumpFromData.data.roleIds && jumpFromData.data.roleIds.length>0){
            let userId = jumpFromData.data.userId;
            let roleIds = jumpFromData.data.roleIds;
            console.log('角色分配数据:', {userId, roleIds});
            //角色分配
            handleRoleAssignment(userId, roleIds);
            window.history.replaceState(null, '', window.location.pathname);
          }
          break;
        default:
          console.log('从其他页面过来');
      }
    }
  }, [location.state]);

  // 处理角色分配
  const handleRoleAssignment = async (userId: number, roleIds: number[]) => {
    try {
      await assignUserRoles(userId, roleIds);
      messageApi.success(intl.formatMessage({ id: 'pages.users.action.roleAssignSuccess' }));
      // 刷新用户列表
      actionRef.current?.reload();
    } catch (error: any) {
      handleError(
        error,
        'pages.users.action.roleAssignError',
        { showDetail: true }
      );
    }
  };

  const fetchUserList = async (params: any, sort: any, filter: any) => {
    try {
      const result = await userList(params, sort, filter);
      return {
        ...result,
        success: true,
      };
    } catch (error: any) {
      handleError(
        error,
        'pages.users.message.listLoadError',
        { showDetail: true }
      );
      return {
        data: [],
        success: false,
        total: 0,
      };
    }
  };

  const columns: ProColumns<API.UserListItem>[] = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: '8%',
      hideInSearch: true,
      render: (_, record) => record.userInfo.id,
    },
    {
      title: <FormattedMessage id="pages.users.searchTable.username" />,
      dataIndex: 'username',
      width: '15%',
      fieldProps: {
        style: {
          width: 160,
        },
      },
      render: (_, record) => record.userInfo.username,
    },
    {
      title: <FormattedMessage id="pages.users.searchTable.role" />,
      dataIndex: 'role',
      width: '20%',
      fieldProps: {
        style: {
          width: 160,
        },
      },
      render: (_, record) => (
        <>
          {record.userRoles.map((role) => (
            <Tooltip title={role.description} key={role.id}>
              <Tag color="blue">{role.name}</Tag>
            </Tooltip>
          ))}
        </>
      ),
    },
    {
      title: <FormattedMessage id="pages.users.searchTable.createTime" />,
      sorter: true,
      hideInForm: true,
      width: '20%',
      dataIndex: 'createTime',
      key: 'updatedAt',
      valueType: 'dateTimeRange',
      fieldProps: {
        style: {
          width: 320,
        },
      },
      render: (_, record) => {
        if (!record.userInfo.createTime) return '-';
        return record.userInfo.createTime.replace('T', ' ').substring(0, 19);
      },
    },
    {
      title: <FormattedMessage id="pages.users.searchTable.updateTime" />,
      sorter: true,
      width: '20%',
      hideInSearch: true,
      dataIndex: 'updateTime',
      valueType: 'dateTime',
      render: (_, record) => {
        if (!record.userInfo.updateTime) return '-';
        return record.userInfo.updateTime.replace('T', ' ').substring(0, 19);
      },
    },
    ...(access['users:assign-roles'] ? [
      {
        title: <FormattedMessage id="pages.system.titleOption" />,
        dataIndex: 'option',
        valueType: 'option' as const,
        width: '12%',
        render: (_: any, record: API.UserListItem) => [
          <a
            key="edit"
            onClick={async (e) => {
              e.preventDefault();

              try {
                const roleIds = record.userRoles.map((item) => item.id);
                const jumpData: JumpData = {
                  flag: 'users',
                  data: {
                    userId: record.userInfo.id,
                    roleIds: roleIds,
                  }
                };
                navigate('/user-center/roles-permission', {state: jumpData});
              } catch (error) {
                handleError(
                  error,
                  'pages.users.action.error',
                  { showDetail: true }
                );
              }
            }}
            style={{cursor: 'pointer'}}
          >
            <FormattedMessage id="pages.users.action.assignRoles" />
          </a>
        ],
      }
    ] : []),
  ];

  return (
    <PageContainer>
      {contextHolder}
      <ProTable<API.UserListItem, API.PageParams>
        headerTitle={intl.formatMessage({ id: 'pages.users.searchTable.title' })}
        actionRef={actionRef}
        rowKey={(record) => record.userInfo.id}
        search={{
          labelWidth: 'auto',
          span: {
            xs: 24,
            sm: 12,
            md: 8,
            lg: 8,
            xl: 6,
            xxl: 6,
          },
        }}
        request={fetchUserList}
        columns={columns}
      />
    </PageContainer>
  );
};

export default UserTableList;
