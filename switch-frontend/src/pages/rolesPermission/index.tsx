import {ActionType, PageContainer, ProColumns, ProTable} from '@ant-design/pro-components';
import {FormattedMessage, useIntl, useNavigate} from '@umijs/max';
import {Button, message, Space, Tag, Tooltip} from 'antd';
import React, {useEffect, useRef, useState} from 'react';
import {useLocation} from 'react-router-dom';
import {rolesPermissionList} from "@/services/rolesPermission/api";
import {Access, useAccess} from "@@/exports";
import {PlusOutlined} from "@ant-design/icons";
import CreateUpdateForm from "./components/CreateUpdateForm";
import {useErrorHandler} from "@/utils/useErrorHandler";

const RolesPermissionList: React.FC = () => {
  const navigate = useNavigate();
  const intl = useIntl();
  const [selfSelectedRows, setSelfSelectedRows] = useState<API.RolesListItem[]>([]);
  const [selfSelectedRowKeys, setSelfSelectedRowKeys] = useState<React.Key[]>([]);
  const location = useLocation();
  const actionRef = useRef<ActionType | null>(null);
  const [messageApi, contextHolder] = message.useMessage();
  const access = useAccess() as Record<string, boolean>;
  const [createModalOpen, handleModalOpen] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<API.RolesListItem>();
  const [jumpData, setJumpData] = useState<API.AssignRolesReq>();
  const [fromUsers, setFromUsers] = useState<boolean>(false);
  const [hasInitialized, setHasInitialized] = useState<boolean>(false);
  const { handleError } = useErrorHandler();

  const getJumpInfo = () => {
    const stepData = location.state as JumpData<API.AssignRolesReq>;
    if (stepData) {
      const {flag, data} = stepData;
      return {flag, data};
    }
    return null;
  }

  //区别不同页面
  useEffect(() => {
    let jumpFromData = getJumpInfo();
    if (jumpFromData) {
      switch (jumpFromData.flag) {
        case 'users':
          console.log('从用户列表页面过来，数据已处理');
          console.log('从用户列表页面过来，数据是：', jumpFromData.data);
          setFromUsers(true);
          setJumpData(jumpFromData.data);
          break;
        default:
          console.log('从其他页面过来');
      }
    }
  }, [location.state]);


  const handleRemoveSelected = (itemToRemove: API.RolesListItem) => {
    const newSelectedRows = selfSelectedRows.filter(row => row.id !== itemToRemove.id);
    const newRowIds = selfSelectedRowKeys.filter(id => id !== itemToRemove.id);
    setSelfSelectedRows(newSelectedRows);
    //这里完成tag跟列表的联动
    setSelfSelectedRowKeys(newRowIds);
  };

  const fetchRolesPermissionList = async (params: any, sort: any, filter: any) => {
    try {
      const result = await rolesPermissionList(params, sort, filter);

      // 如果有跳转数据且未初始化，回显选中的角色
      if (!hasInitialized && jumpData && jumpData.roleIds && jumpData.roleIds.length > 0 && result.data) {
        const selectedRows = result.data.filter((item: API.RolesListItem) =>
          jumpData.roleIds.includes(item.id)
        );
        setSelfSelectedRows(selectedRows);
        setSelfSelectedRowKeys(jumpData.roleIds);
        setHasInitialized(true);
      }

      return {
        ...result,
        success: true,
      };
    } catch (error: any) {
      handleError(
        error,
        'pages.rolesPermission.message.listLoadError',
        { showDetail: true }
      );
      return {
        data: [],
        success: false,
        total: 0,
      };
    }
  };

  const columns: ProColumns<API.RolesListItem>[] = [
    {
      title: (
        <FormattedMessage
          id="pages.rolesPermission.searchTable.id"
        />
      ),
      dataIndex: 'id',
      width: '8%',
      hideInSearch: true,
    },
    {
      title: (
        <FormattedMessage
          id="pages.rolesPermission.searchTable.roleName"
        />
      ),
      dataIndex: 'name',
      key: 'roleName',
      width: '15%',
    },
    {
      title: (
        <FormattedMessage
          id="pages.rolesPermission.searchTable.permissionName"
        />
      ),
      key: 'permissionName',
      hideInTable: true,
    },
    {
      title: (
        <FormattedMessage
          id="pages.rolesPermission.searchTable.description"
        />
      ),
      dataIndex: 'description',
      width: '20%',
      hideInSearch: true,
    },
    {
      title: (
        <FormattedMessage
          id="pages.rolesPermission.searchTable.permissions"
        />
      ),
      dataIndex: 'permissions',
      width: '30%',
      hideInSearch: true,
      render: (_, record) => (
        <>
          {record.permissions?.map((permission) => (
            <Tooltip title={permission.description} key={permission.id}>
              <Tag color="green">{permission.name}</Tag>
            </Tooltip>
          ))}
        </>
      ),
    },
    {
      title: (
        <FormattedMessage
          id="pages.rolesPermission.searchTable.createTime"
        />
      ),
      sorter: true,
      width: '15%',
      hideInSearch: true,
      dataIndex: 'createTime',
      valueType: 'dateTime',
    },
    ...(access['roles:edit'] ? [
      {
        title: (
          <FormattedMessage
            id="pages.system.titleOption"
          />
        ),
        dataIndex: 'option',
        valueType: 'option' as const,
        width: '12%',
        render: (_: any, record: API.RolesListItem) => [
          record.namespaceTag === "" ? (
            <Tooltip title={intl.formatMessage({id: 'pages.rolesPermission.searchTable.systemRoleTooltip'})} key="edit-disabled">
              <a key="edit" style={{color: 'rgba(0, 0, 0, 0.25)', cursor: 'not-allowed'}}>
                <FormattedMessage id="pages.rolesPermission.searchTable.edit" />
              </a>
            </Tooltip>
          ) : (
            <a key="edit" onClick={() => {
              handleModalOpen(true);
              setCurrentRow(record);
            }}>
              <FormattedMessage id="pages.rolesPermission.searchTable.edit" />
            </a>
          )
        ],
      }
    ]: []),
  ];

  return (
    <PageContainer>
      {contextHolder}
      <ProTable<API.RolesListItem, API.PageParams>
        headerTitle={intl.formatMessage({id: 'pages.rolesPermission.searchTable.title'})}
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 'auto',
        }}
        toolBarRender={() => [
          <Access accessible={!!(access['permissions:create'])} key="create">
            <Button
              type="primary"
              onClick={() => {
                setCurrentRow(undefined);
                handleModalOpen(true);
              }}
            >
              <PlusOutlined/> <FormattedMessage id="pages.rolesPermission.searchTable.create" />
            </Button>
          </Access>,
        ]}
        request={fetchRolesPermissionList}
        columns={columns}
        rowSelection={
          fromUsers ? {
            //key用自己的，因为tag可以被删除，需要用自己的去替换组件内部的然后完成列表的删除跟tag删除的同步
            selectedRowKeys: selfSelectedRowKeys,
            onChange: (selectedRowKeys, selectedRows) => {
              setSelfSelectedRows(selectedRows);
              setSelfSelectedRowKeys(selectedRowKeys);
            },
          } : undefined}
        // 自定义左侧的提示信息
        tableAlertRender={({selectedRowKeys, selectedRows, onCleanSelected}) => {
          return (
            <Space size={24} style={{
              display: "flex",
              alignItems: "flex-start",
              gap: "16px",
              width: "100%"
            }}>
              <div style={{
                flex: "0 0 23",
                flexShrink: 0,
                whiteSpace: "nowrap",
                overflow: "hidden",
                textOverflow: "ellipsis"
              }}>
                <FormattedMessage
                  id="pages.env.searchTable.select"
                  values={{
                    selectedRowKeysLength: selectedRowKeys.length
                  }}
                />
              </div>
              <div style={{
                flex: "1",
                display: "flex",
                flexWrap: "wrap",
                gap: "8px",
                maxHeight: "80px",
                overflowY: "auto"
              }}>
                {selectedRows.map((row) => (
                  <Tag
                    color="processing"
                    key={row.id}
                    closable
                    onClose={() => {
                      handleRemoveSelected(row)
                    }}
                  >
                    {row.name}
                  </Tag>
                ))}
              </div>

            </Space>
          );
        }}
        tableAlertOptionRender={({selectedRowKeys, selectedRows, onCleanSelected}) => {
          return (
            <Space size={16}>
              <a onClick={onCleanSelected}></a>
              <Button
                type="primary"
                key="clear"
                onClick={() => {
                  setSelfSelectedRows([])
                  setSelfSelectedRowKeys([])
                  onCleanSelected()
                }}
              >
                <FormattedMessage
                  id="pages.system.clearAll"/>
              </Button>
              <Button
                type="primary"
                key="confirm"
                onClick={() => {
                  if (selfSelectedRows.length === 0) {
                    messageApi.error(intl.formatMessage({id: 'pages.rolesPermission.message.selectAtLeastOne'}));
                    return;
                  }
                  //合并命名空间的数据一起往回写
                  navigate('/user-center/users', {
                    state: {
                      flag : "rolesPermission",
                      data :{
                        ...(jumpData || {}),
                        roleIds: selfSelectedRowKeys,
                      },
                    }
                  })
                }}
              >
                <FormattedMessage
                  id="pages.system.bandSuccess"/>
              </Button>
            </Space>
          );
        }}
      />
      <CreateUpdateForm
        onCancel={() => handleModalOpen(false)}
        onSubmit={async (value) => {
          handleModalOpen(false);
          if (actionRef.current) {
            actionRef.current.reload();
          }
        }}
        modalOpen={createModalOpen}
        values={currentRow || {}}
      />
    </PageContainer>
  );
};

export default RolesPermissionList;
