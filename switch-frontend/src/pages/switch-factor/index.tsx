import {ActionType, PageContainer, ProColumns, ProTable,} from '@ant-design/pro-components';
import {Access, FormattedMessage, useAccess, useIntl} from '@umijs/max';
import {Button, message, Modal} from 'antd';
import React, {useRef, useState} from 'react';
import JsonView from '@uiw/react-json-view';
import CreateUpdateForm from "./components/CreateUpdateForm";
import {PlusOutlined} from "@ant-design/icons";
import {switchFactors} from "@/services/switch-factor/api";
import {useErrorHandler} from "@/utils/useErrorHandler";


const TableList: React.FC = () => {
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<API.SwitchFactorItem>();
  const actionRef = useRef<ActionType | null>(null);
  const access = useAccess() as Record<string, boolean>;

  /**
   * @en-US International configuration
   * @zh-CN 国际化配置
   * */
  const intl = useIntl();

  const [messageApi, contextHolder] = message.useMessage();
  const { handleError } = useErrorHandler();

  const fetchSwitchFactorList = async (params: any, sort: any, filter: any) => {
    try {
      const result = await switchFactors(params, sort, filter);
      return {
        ...result,
        success: true,
      };
    } catch (error: any) {
      handleError(
        error,
        'pages.switch-factor.form.listLoadError',
        { showDetail: true }
      );
      return {
        data: [],
        success: false,
        total: 0,
      };
    }
  };

  const columns: ProColumns<API.SwitchFactorItem>[] = [
    {
      title: (
        <FormattedMessage
          id="pages.namespace.searchTable.id"
        />
      ),
      dataIndex: 'id',
      width: 80,
      hideInSearch: true,
    },
    {
      title: (
        <FormattedMessage
          id="pages.switch-factor.searchTable.factorAlias"
        />
      ),
      dataIndex: 'name',
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
          id="pages.switch-factor.searchTable.factor"
        />
      ),
      dataIndex: 'factor',
      width: 150,
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
          id="pages.switch-factor.searchTable.description"
        />
      ),
      dataIndex: 'description',
      width: 150,
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
          id="pages.switch-factor.searchTable.namespaceBind"
        />
      ),
      dataIndex: 'namespaceId',
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
          id="pages.switch-factor.searchTable.createBy"
        />
      ),
      dataIndex: 'createdBy',
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
          id="pages.switch-factor.searchTable.createTime"
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
          id="pages.switch-factor.searchTable.updateBy"
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
          id="pages.switch-factor.searchTable.updateTime"
        />
      ),
      sorter: true,
      hideInForm: true,
      hideInSearch: true,
      width: 170,
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
    {
      title: (
        <FormattedMessage
          id="pages.system.titleOption"
        />
      ),
      dataIndex: 'option',
      valueType: 'option' as const,
      width: 170,
      render: (_: any, record: API.SwitchFactorItem) => {
        const canEdit = access['switch-factors:edit'] && record.namespaceTag !== "";

        return [
          <CreateUpdateForm
            isCreate={false}
            readOnly={!canEdit}
            trigger={
              <a>
                <FormattedMessage
                  id={canEdit ? "pages.switch-factor.searchTable.edit" : "pages.switch-factor.searchTable.view"}
                />
              </a>
            }
            key="edit"
            reload={() => actionRef.current?.reload()}
            values={record}
          />
        ];
      },
    },
  ];

  return (
    <PageContainer>
      {contextHolder}
      <ProTable<API.SwitchFactorItem, API.PageParams>
        headerTitle={intl.formatMessage({
          id: 'pages.switch-factor.searchTable.title',
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
        toolBarRender={() => [
          <Access accessible={!!access['switch-factors:create']}>
            <CreateUpdateForm
              isCreate={true}
              trigger={
                (
                  <Button type={'primary'} icon={<PlusOutlined/>}>
                    <FormattedMessage
                      id="pages.switch-factor.searchTable.createIcon"/>
                  </Button>
                )
              }
              key="create"
              reload={() => actionRef.current?.reload()}
            />
          </Access>
        ]}
        request={fetchSwitchFactorList}
        columns={columns}
      />

      {showDetail && (
        <Modal
          title={intl.formatMessage({ id: 'pages.switch-factor.detail.title' })}
          open={showDetail}
          onCancel={() => setShowDetail(false)}
          footer={null}
          width={800}
          centered
        >
          <JsonView
            value={{
              ...currentRow,
              json_schema: currentRow?.jsonSchema ? (() => {
                try {
                  return JSON.parse(currentRow.jsonSchema);
                } catch (e) {
                  return currentRow.jsonSchema;
                }
              })() : undefined
            }}
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

