import {ActionType, PageContainer, ProColumns, ProTable,} from '@ant-design/pro-components';
import {FormattedMessage, useIntl} from '@umijs/max';
import {Button, message, Modal} from 'antd';
import React, {useRef, useState} from 'react';
import {useLocation} from 'react-router-dom';
import {namespaceList} from '@/services/namespace/api';
import JsonView from '@uiw/react-json-view';
import CreateUpdateForm from './components/CreateUpdateForm';
import {PlusOutlined} from "@ant-design/icons";
import {useAccess} from "@@/exports";
import {useErrorHandler} from "@/utils/useErrorHandler";

interface LocationState {
  selectedEnv?: API.EnvItem[];
}

const TableList: React.FC = () => {
  const location = useLocation();

  const state: LocationState = location.state || {};

  console.log('location.state?.selectedEnv', state.selectedEnv);
  const actionRef = useRef<ActionType | null>(null);
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<API.NamespaceItem>();
  const { handleError } = useErrorHandler();

  const access = useAccess() as Record<string, boolean>;

  /**
   * @en-US International configuration
   * @zh-CN 国际化配置
   * */
  const intl = useIntl();

  const [messageApi, contextHolder] = message.useMessage();

  const fetchNamespaceList = async (params: any, sort: any, filter: any) => {
    try {
      const result = await namespaceList(params, sort, filter);
      return {
        ...result,
        success: true,
      };
    } catch (error: any) {
      handleError(
        error,
        'pages.namespace.message.listLoadError',
        { showDetail: true }
      );
      return {
        data: [],
        success: false,
        total: 0,
      };
    }
  };

  const columns: ProColumns<API.NamespaceItem>[] = [
    {
      title: (
        <FormattedMessage
          id="pages.namespace.searchTable.id"
        />
      ),
      dataIndex: 'id',
      width: 80,
      hideInSearch: true,
      fixed: 'left',
    },
    {
      title: (
        <FormattedMessage
          id="pages.namespace.searchTable.name"
        />
      ),
      dataIndex: 'name',
      width: 150,
      fixed: 'left',
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
          id="pages.namespace.searchTable.tag"
        />
      ),
      dataIndex: 'tag',
      width: 150,
      fixed: 'left',
      valueType: 'textarea',
      fieldProps: {
        style: {
          width: '150px',
        },
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.namespace.searchTable.description"
        />
      ),
      dataIndex: 'description',
      hideInForm: true,
      width: 250,
      fieldProps: {
        style: {
          width: '150px',
        },
      },
    },
    {
      title: (
        <FormattedMessage
          id="pages.namespace.searchTable.createBy"
        />
      ),
      dataIndex: 'createdBy',
      sorter: true,
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
          id="pages.namespace.searchTable.createTime"
        />
      ),
      sorter: true,
      hideInForm: true,
      width: 170,
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
          id="pages.namespace.searchTable.updateBy"
        />
      ),
      dataIndex: 'updateBy',
      sorter: true,
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
          id="pages.namespace.searchTable.updateTime"
        />
      ),
      sorter: true,
      hideInForm: true,
      hideInSearch: true,
      width: 270,
      dataIndex: 'updateTime',
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

    ...(access['namespaces:edit'] ? [
      {
      title: (
        <FormattedMessage
          id="pages.system.titleOption"
        />
      ),
      fixed: 'right' as const,
      width: 100,
      dataIndex: 'option',
      valueType: 'option' as const,
      render: (_:any, record:API.NamespaceItem) => [
        <CreateUpdateForm
          isCreate={false}
          trigger={<a>
            <FormattedMessage
              id="pages.namespace.searchTable.edit"/>
          </a>}
          key="edit"
          reload={() => actionRef.current?.reload()}
          values={record}/>,
      ],
    }
    ]: []),
  ];

  return (
    <PageContainer>
      {contextHolder}
      <ProTable<API.NamespaceItem, API.PageParams>
        headerTitle={intl.formatMessage({
          id: 'pages.namespace.searchTable.title',
        })}
        scroll={{x: 1300}}
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
        toolBarRender={() => [
          <CreateUpdateForm
            isCreate={true}
            trigger={
              <Button type="primary" icon={<PlusOutlined/>}>
                <FormattedMessage id="pages.namespace.searchTable.createIcon"/>
              </Button>
            }
            key="create"
            reload={() => actionRef.current?.reload()}
          />
        ]}
        request={fetchNamespaceList}
        columns={columns}
      />

      {showDetail && (
        <Modal
          title={intl.formatMessage({id: 'pages.namespace.detail'})}
          open={showDetail}
          onCancel={() => setShowDetail(false)}
          footer={null}
          width={800}
          centered
          destroyOnHidden={true}
        >
          <JsonView
            value={currentRow}
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
