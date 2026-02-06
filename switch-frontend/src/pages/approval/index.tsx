import {ActionType, PageContainer, ProColumns, ProTable,} from '@ant-design/pro-components';
import {FormattedMessage, useIntl} from '@umijs/max';
import {Card, message, Modal, Popconfirm, Tabs, Tag} from 'antd';
import React, {useRef, useState} from 'react';
import JsonView from '@uiw/react-json-view';
import RejectForm from './components/RejectForm';
import {approveRequest, myRequests, allRequests, userLike} from '@/services/approval/api';
import {useModel} from '@umijs/max';
import {useAccess} from "@@/exports";
import { useErrorHandler } from '@/utils/useErrorHandler';

const TableList: React.FC = () => {
  const { initialState } = useModel('@@initialState');
  const currentUser = initialState?.currentUser;
  const isSuperAdmin = currentUser?.is_super_admin;

  const [activeTab, setActiveTab] = useState(isSuperAdmin ? 'allApprovals' : 'myRequests');
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<API.ApprovalDetailViewItem>();
  const actionRef = useRef<ActionType | null>(null);
  const access = useAccess() as Record<string, boolean>;

  const isMyRequests = () => {
    return activeTab === 'myRequests';
  }

  /**
   * @en-US International configuration
   * @zh-CN 国际化配置
   * */
  const intl = useIntl();
  const { handleError } = useErrorHandler();

  const [messageApi, contextHolder] = message.useMessage();

  const fetchApprovalList = async (params: any, sort: any, filter: any) => {
    let queryParams;

    if (isSuperAdmin) {
      queryParams = {
        ...params,
      };
    } else {
      //invited：0-发起人 1-受邀人
      queryParams = {
        ...params,
        invited: activeTab === 'myRequests' ? 0 : 1,
      };
    }

    try {
      const result = isSuperAdmin ?
        await allRequests(queryParams, sort, filter) :
        await myRequests(queryParams, sort, filter);
      return {
        ...result,
        success: true,
      };
    } catch (error: any) {
      handleError(
        error,
        'pages.approvals.action.listLoadError',
        { showDetail: true }
      );
      return {
        data: [],
        success: false,
        total: 0,
      };
    }
  };

  const columns: ProColumns<API.ApprovalDetailViewItem>[] = [
    {
      title: (
        <FormattedMessage
          id="pages.approvals.searchTable.id"
        />
      ),
      dataIndex: 'id',
      width: 80,
      hideInSearch: true,
    },
    {
      title: (
        <FormattedMessage
          id="pages.approvals.searchTable.approvableType"
        />
      ),
      dataIndex: 'approvableType',
      width: 150,
      order: 10,
      valueType: 'select',
      valueEnum: {
        1: {text: intl.formatMessage({ id: 'pages.approvals.type.namespace' })},
        2: {text: intl.formatMessage({ id: 'pages.approvals.type.switch' })},
      },
      fieldProps: {
        style: {
          width: '150px',
        },
      },
      render: (dom, entity) => {
        return (
          <a
            onClick={() => {
              setCurrentRow(entity);
              setShowDetail(true);
            }}
          >
            {dom}
          </a>
        );
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.approvals.searchTable.approverUsers"
        />
      ),
      dataIndex: 'approverUsersStr',
      hideInSearch: true,
      hideInForm: true,
      width: 150,
      fieldProps: {
        style: {
          width: '150px',
        },
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.approvals.searchTable.approvalNotes"
        />
      ),
      dataIndex: 'approvalNotes',
      hideInForm: true,
      hideInSearch: true,
      width: 150,
      fieldProps: {
        style: {
          width: '150px',
        },
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.approvals.searchTable.requesterUser"
        />
      ),
      dataIndex: 'requesterUser',
      sorter: true,
      hideInSearch: isMyRequests() || isSuperAdmin,
      hideInForm: true,
      order: 9,
      width: 150,
      valueType: 'select',
      fieldProps: {
        showSearch: true,
        style: {
          width: '150px',
        },
      },
      render: (_, record) => record.requesterUserStr,
      request: async ({keyWords}) => {
        try {
          if (!keyWords) {
            return [];
          }
          const res = await userLike({username: keyWords});
          return (res.data || []).map((user: API.CurrentUser) => ({
            label: user.username,
            value: user.id,
          }));
        }catch (error){
          handleError(
            error,
            'pages.approvals.userList.fail',
            {showDetail: true,}
          );
        }
        return [];
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.approvals.searchTable.createTime"
        />
      ),
      sorter: true,
      width: 170,
      hideInSearch: true,
      hideInForm: true,
      order: 1,
      dataIndex: 'createTime',
      key: 'updatedAt',
      valueType: 'dateTimeRange',
      render: (_, record) => {
        if (!record.createTime) return '-';
        return record.createTime.replace('T', ' ').substring(0, 19);
      },
      fieldProps: {
        style: {
          width: '360px',
        },
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.approvals.searchTable.approverUser"
        />
      ),
      dataIndex: 'approverUser',
      sorter: true,
      order: 9,
      hideInSearch: !isMyRequests() || isSuperAdmin,
      hideInForm: true,
      width: 150,
      valueType: 'select',
      render: (_, record) => record.approverUserStr,
      fieldProps: {
        showSearch: true,
        style: {
          width: '150px',
        },
      },
      request: async ({keyWords}) => {
        try {
          if (!keyWords) {
            return [];
          }
          const res = await userLike({username: keyWords});
          return (res.data || []).map((user: API.CurrentUser) => ({
            label: user.username,
            value: user.id,
          }));
        }catch (error){
          handleError(
            error,
            'pages.approvals.userList.fail',
            {showDetail: true,}
          );
        }
        return [];
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.approvals.searchTable.approvalTime"
        />
      ),
      sorter: true,
      hideInSearch: true,
      width: 170,
      dataIndex: 'approvalTime',
      render: (_, record) => {
        if (!record.createTime) return '-';
        return record.createTime.replace('T', ' ').substring(0, 19);
      },
      fieldProps: {
        style: {
          width: '360px',
        },
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.approvals.searchTable.status"
        />
      ),
      dataIndex: 'status',
      hideInForm: true,
      order: 8,
      width: 150,
      valueType: 'select',
      valueEnum: {
        'PENDING': { text: intl.formatMessage({ id: 'pages.approvals.status.pending' }), status: 'Processing' },
        'APPROVED': { text: intl.formatMessage({ id: 'pages.approvals.status.approved' }), status: 'Success' },
        'REJECTED': { text: intl.formatMessage({ id: 'pages.approvals.status.rejected' }), status: 'Error' },
      },
      render: (_, record) => {
        let color = 'default';
        let text: string = record.status;

        if (record.status === 'PENDING') {
          color = 'processing';
          text = intl.formatMessage({ id: 'pages.approvals.status.pending' });
        } else if (record.status === 'APPROVED') {
          color = 'success';
          text = intl.formatMessage({ id: 'pages.approvals.status.approved' });
        } else if (record.status === 'REJECTED') {
          color = 'error';
          text = intl.formatMessage({ id: 'pages.approvals.status.rejected' });
        }
        return <Tag color={color}>{text}</Tag>;
      },
      fieldProps: {
        style: {
          width: '150px',
        },
      },
    },

    // 普通人 待我审批 & 有审批权限 出操作
    // 超管员 全部审批 一律出操作
    ...(((activeTab !== 'myRequests' && access['approvals:approve']) || isSuperAdmin) ? [
      {
        title: (
          <FormattedMessage
            id="pages.system.titleOption"
          />
        ),
        dataIndex: 'option',
        valueType: 'option'  as const,
        width: 170,
        render: (_:any, record:API.ApprovalDetailViewItem) => {
          if (record.status == "PENDING") {
            return [
              <Popconfirm
                key="approve"
                title={intl.formatMessage({ id: 'pages.approvals.action.confirmApprove' })}
                onConfirm={async () => {
                  await approveRequest({id: record.id, notes: "", status: 1});
                  actionRef.current?.reload();
                }}
              >
                <a>
                  <FormattedMessage id="pages.approvals.searchTable.agree"/>
                </a>
              </Popconfirm>,
              <RejectForm
                trigger={
                  <a>
                    <FormattedMessage id="pages.approvals.searchTable.reject"/>
                  </a>
                }
                key="reject"
                reload={() => actionRef.current?.reload()}
                values={record}
              />,
            ];
          }
          return null;
        },
      },
    ]: []),
  ];

  return (
    <PageContainer>
      {contextHolder}
      <Card>
        <Tabs
          activeKey={activeTab}
          onChange={(key) => {
            setActiveTab(key);
            actionRef.current?.reload();
          }}
          items={[
            ...(!isSuperAdmin ? [
              { label: intl.formatMessage({ id: 'pages.approvals.tab.myRequests' }), key: 'myRequests' },
              { label: intl.formatMessage({ id: 'pages.approvals.tab.myApprovals' }), key: 'myApprovals' },
            ] : []),
            ...(isSuperAdmin ? [
              { label: intl.formatMessage({ id: 'pages.approvals.tab.allApprovals' }), key: 'allApprovals' },
            ] : []),
          ]}
        />
        <ProTable<API.ApprovalDetailViewItem, API.PageParams>
          headerTitle={intl.formatMessage({
            id: 'pages.approvals.searchTable.title',
          })}
          actionRef={actionRef}
          rowKey="id"
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
          request={fetchApprovalList}
          columns={columns}
        />
      </Card>

      {showDetail && (
        <Modal
          title={intl.formatMessage({ id: 'pages.approvals.detail.title' })}
          open={showDetail}
          onCancel={() => setShowDetail(false)}
          footer={null}
          width={800}
          centered
          destroyOnHidden={true}
        >
          <JsonView
            value={currentRow?.approvable || {}}
            style={{
              backgroundColor: '#f6f8fa',
              padding: '16px',
              borderRadius: '6px'
            }}
            collapsed={false}
          />
        </Modal>
      )}

    </PageContainer>
  );
};

export default TableList;
